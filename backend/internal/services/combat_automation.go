package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/dice"
)

type CombatAutomationService struct {
	combatRepo    database.CombatAnalyticsRepository
	characterRepo database.CharacterRepository
	npcRepo       database.NPCRepository
	diceRoller    *dice.Roller
}

func NewCombatAutomationService(
	combatRepo database.CombatAnalyticsRepository,
	characterRepo database.CharacterRepository,
	npcRepo database.NPCRepository,
) *CombatAutomationService {
	return &CombatAutomationService{
		combatRepo:    combatRepo,
		characterRepo: characterRepo,
		npcRepo:       npcRepo,
		diceRoller:    dice.NewRoller(),
	}
}

// AutoResolveCombat performs quick combat resolution for minor encounters
func (cas *CombatAutomationService) AutoResolveCombat(
	ctx context.Context,
	sessionID uuid.UUID,
	characters []*models.Character,
	req models.AutoResolveRequest,
) (*models.AutoCombatResolution, error) {
	// Calculate encounter difficulty
	partyLevel := cas.calculateAveragePartyLevel(characters)
	encounterCR := cas.calculateEncounterCR(req.EnemyTypes)

	// Build party and enemy compositions
	partyComp := cas.buildPartyComposition(characters)
	enemyComp := cas.buildEnemyComposition(req.EnemyTypes)

	// Simulate combat
	outcome, rounds, resources := cas.simulateCombat(
		characters,
		req.EnemyTypes,
		partyLevel,
		encounterCR,
		req.UseResources,
	)

	// Generate loot and experience
	loot := cas.generateLoot(req.EncounterDifficulty, req.EnemyTypes)
	experience := cas.calculateExperience(req.EnemyTypes, len(characters))

	// Create narrative summary
	narrative := cas.generateNarrativeSummary(
		outcome,
		rounds,
		req.TerrainType,
		partyLevel,
		encounterCR,
	)

	// Convert resources to JSONB
	resourcesJSON, err := json.Marshal(resources)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resources: %w", err)
	}

	// Convert loot to JSONB
	lootJSON, err := json.Marshal(loot)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal loot: %w", err)
	}

	// Save resolution to database
	resolution := &models.AutoCombatResolution{
		ID:                  uuid.New(),
		GameSessionID:       sessionID,
		EncounterDifficulty: req.EncounterDifficulty,
		PartyComposition:    models.JSONB(partyComp),
		EnemyComposition:    models.JSONB(enemyComp),
		ResolutionType:      "quick",
		Outcome:             outcome,
		RoundsSimulated:     rounds,
		PartyResourcesUsed:  models.JSONB(resourcesJSON),
		LootGenerated:       models.JSONB(lootJSON),
		ExperienceAwarded:   experience,
		NarrativeSummary:    narrative,
		CreatedAt:           time.Now(),
	}

	if err := cas.combatRepo.CreateAutoCombatResolution(resolution); err != nil {
		return nil, fmt.Errorf("failed to save combat resolution: %w", err)
	}

	return resolution, nil
}

// SmartInitiative calculates and assigns initiative automatically
func (cas *CombatAutomationService) SmartInitiative(
	ctx context.Context,
	sessionID uuid.UUID,
	req models.SmartInitiativeRequest,
) ([]models.InitiativeEntry, error) {
	var entries []models.InitiativeEntry

	for _, combatant := range req.Combatants {
		// Get any special initiative rules
		rule, _ := cas.combatRepo.GetInitiativeRule(sessionID, combatant.ID)

		// Calculate initiative bonus
		bonus := combatant.DexterityModifier
		if rule != nil {
			bonus += rule.BaseInitiativeBonus
			if rule.AlertFeat {
				bonus += 5 // Alert feat gives +5 to initiative
			}
		}

		// Roll initiative
		var roll int
		if rule != nil && rule.AdvantageOnInitiative {
			// Roll with advantage
			roll1Result, err := cas.diceRoller.Roll("1d20")
			if err != nil {
				return nil, fmt.Errorf("failed to roll initiative: %w", err)
			}
			roll2Result, err := cas.diceRoller.Roll("1d20")
			if err != nil {
				return nil, fmt.Errorf("failed to roll initiative: %w", err)
			}
			if roll1Result.Total > roll2Result.Total {
				roll = roll1Result.Total
			} else {
				roll = roll2Result.Total
			}
		} else {
			// Normal roll
			rollResult, err := cas.diceRoller.Roll("1d20")
			if err != nil {
				return nil, fmt.Errorf("failed to roll initiative: %w", err)
			}
			roll = rollResult.Total
		}

		total := roll + bonus

		// Apply any special rules
		if rule != nil && rule.SpecialRules != nil {
			// Handle special cases like "always goes first" etc.
			// First unmarshal the JSONB to a map
			var specialRules map[string]interface{}
			if err := json.Unmarshal([]byte(rule.SpecialRules), &specialRules); err == nil {
				if specialPriority, ok := specialRules["priority"].(float64); ok {
					total += int(specialPriority * 100) // Ensure they go first/last
				}
			}
		}

		entries = append(entries, models.InitiativeEntry{
			ID:         combatant.ID,
			Type:       combatant.Type,
			Name:       combatant.Name,
			Initiative: total,
			Roll:       roll,
			Bonus:      bonus,
		})
	}

	// Sort by initiative (highest first)
	cas.sortInitiativeEntries(entries)

	// Handle ties
	cas.resolveTies(entries)

	return entries, nil
}

// SaveBattleMap saves a battle map to the database
func (cas *CombatAutomationService) SaveBattleMap(ctx context.Context, battleMap *models.BattleMap) error {
	return cas.combatRepo.CreateBattleMap(battleMap)
}

// GetBattleMap retrieves a battle map by ID
func (cas *CombatAutomationService) GetBattleMap(ctx context.Context, mapID uuid.UUID) (*models.BattleMap, error) {
	return cas.combatRepo.GetBattleMap(mapID)
}

// GetBattleMapsBySession retrieves all battle maps for a session
func (cas *CombatAutomationService) GetBattleMapsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.BattleMap, error) {
	return cas.combatRepo.GetBattleMapsBySession(sessionID)
}

// GetAutoResolutionsBySession retrieves all auto-resolutions for a session
func (cas *CombatAutomationService) GetAutoResolutionsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AutoCombatResolution, error) {
	return cas.combatRepo.GetAutoCombatResolutionsBySession(sessionID)
}

// SetInitiativeRule sets or updates an initiative rule
func (cas *CombatAutomationService) SetInitiativeRule(ctx context.Context, rule *models.SmartInitiativeRule) error {
	return cas.combatRepo.CreateOrUpdateInitiativeRule(rule)
}

// Helper methods

func (cas *CombatAutomationService) calculateAveragePartyLevel(characters []*models.Character) float64 {
	if len(characters) == 0 {
		return 1
	}

	totalLevel := 0
	for _, char := range characters {
		totalLevel += char.Level
	}

	return float64(totalLevel) / float64(len(characters))
}

func (cas *CombatAutomationService) calculateEncounterCR(enemies []models.EnemyInfo) float64 {
	// Convert CR strings to numeric values and calculate effective CR
	totalCR := 0.0
	for _, enemy := range enemies {
		cr := cas.parseCR(enemy.CR)
		totalCR += cr * float64(enemy.Count)
	}

	// Adjust for multiple enemies
	if len(enemies) > 1 {
		totalCR *= 1.2 // Multiple enemy bonus
	}

	return totalCR
}

func (cas *CombatAutomationService) parseCR(cr string) float64 {
	// Handle fractional CRs like "1/2", "1/4", "1/8"
	switch cr {
	case "1/8":
		return 0.125
	case "1/4":
		return 0.25
	case "1/2":
		return 0.5
	default:
		// Try to parse as integer
		var value float64
		_, _ = fmt.Sscanf(cr, "%f", &value)
		return value
	}
}

func (cas *CombatAutomationService) simulateCombat(
	characters []*models.Character,
	enemies []models.EnemyInfo,
	partyLevel float64,
	encounterCR float64,
	useResources bool,
) (outcome string, rounds int, resources map[string]interface{}) {
	partyStrength := partyLevel * float64(len(characters)) * 10
	encounterStrength := encounterCR * 15

	// Add randomness
	partyRollResult, _ := cas.diceRoller.Roll("1d20")
	enemyRollResult, _ := cas.diceRoller.Roll("1d20")
	partyRoll := float64(partyRollResult.Total)
	enemyRoll := float64(enemyRollResult.Total)

	partyStrength += partyRoll * 5
	encounterStrength += enemyRoll * 5

	// Determine outcome
	if partyStrength > encounterStrength*1.5 {
		outcome = "decisive_victory"
		rounds = 2 + rand.Intn(3)
	} else if partyStrength > encounterStrength {
		outcome = "victory"
		rounds = 3 + rand.Intn(4)
	} else if partyStrength > encounterStrength*0.7 {
		outcome = "costly_victory"
		rounds = 5 + rand.Intn(5)
	} else if partyStrength > encounterStrength*0.5 {
		outcome = "retreat"
		rounds = 3 + rand.Intn(3)
	} else {
		outcome = "defeat"
		rounds = 4 + rand.Intn(4)
	}

	// Calculate resources used
	resources = make(map[string]interface{})

	// HP loss calculation
	hpLossPercent := 0.0
	switch outcome {
	case "decisive_victory":
		hpLossPercent = 0.1 + rand.Float64()*0.1 // 10-20%
	case "victory":
		hpLossPercent = 0.2 + rand.Float64()*0.2 // 20-40%
	case "costly_victory":
		hpLossPercent = 0.4 + rand.Float64()*0.3 // 40-70%
	case "retreat":
		hpLossPercent = 0.3 + rand.Float64()*0.3 // 30-60%
	case "defeat":
		hpLossPercent = 0.6 + rand.Float64()*0.3 // 60-90%
	}

	totalHP := 0
	for _, char := range characters {
		totalHP += char.MaxHitPoints
	}
	resources["hp_lost"] = int(float64(totalHP) * hpLossPercent)

	// Spell slots and abilities used
	if useResources {
		spellSlotsUsed := make(map[int]int)
		maxSlotLevel := int(math.Min(9, math.Ceil(partyLevel/2)))

		for i := 1; i <= maxSlotLevel; i++ {
			if rand.Float64() < 0.3+(float64(rounds)*0.05) {
				spellSlotsUsed[i] = 1 + rand.Intn(2)
			}
		}
		resources["spell_slots_used"] = spellSlotsUsed

		// Other resources
		resources["hit_dice_used"] = rounds / 2
		resources["consumables_used"] = rand.Intn(rounds)
	}

	return outcome, rounds, resources
}

func (cas *CombatAutomationService) generateLoot(difficulty string, enemies []models.EnemyInfo) []map[string]interface{} {
	loot := []map[string]interface{}{}

	// Base gold calculation
	goldMultiplier := map[string]int{
		"trivial": 10,
		"easy":    25,
		"medium":  50,
		"hard":    100,
		"deadly":  200,
	}

	baseGold := goldMultiplier[difficulty]
	if baseGold == 0 {
		baseGold = 50
	}

	totalGold := 0
	for _, enemy := range enemies {
		cr := cas.parseCR(enemy.CR)
		totalGold += int(cr * float64(baseGold) * float64(enemy.Count))
	}

	// Add some variance
	if totalGold > 0 {
		variance := totalGold / 2
		if variance > 0 {
			totalGold = totalGold + rand.Intn(variance) - totalGold/4
		}
	}

	loot = append(loot, map[string]interface{}{
		"type":     "currency",
		"currency": "gp",
		"amount":   totalGold,
	})

	// Chance for items based on difficulty
	itemChance := map[string]float64{
		"trivial": 0.1,
		"easy":    0.2,
		"medium":  0.4,
		"hard":    0.6,
		"deadly":  0.8,
	}

	if rand.Float64() < itemChance[difficulty] {
		// Add a random item
		itemTypes := []string{"potion", "scroll", "weapon", "armor", "trinket"}
		loot = append(loot, map[string]interface{}{
			"type":   "item",
			"name":   fmt.Sprintf("Random %s", itemTypes[rand.Intn(len(itemTypes))]),
			"rarity": cas.getRandomRarity(difficulty),
		})
	}

	return loot
}

func (cas *CombatAutomationService) getRandomRarity(difficulty string) string {
	roll := rand.Float64()

	switch difficulty {
	case "trivial", "easy":
		if roll < 0.95 {
			return "common"
		}
		return "uncommon"
	case "medium":
		if roll < 0.7 {
			return "common"
		} else if roll < 0.95 {
			return "uncommon"
		}
		return "rare"
	case "hard":
		if roll < 0.4 {
			return "common"
		} else if roll < 0.8 {
			return "uncommon"
		} else if roll < 0.95 {
			return "rare"
		}
		return "very_rare"
	case "deadly":
		if roll < 0.2 {
			return "uncommon"
		} else if roll < 0.6 {
			return "rare"
		} else if roll < 0.9 {
			return "very_rare"
		}
		return "legendary"
	default:
		return "common"
	}
}

func (cas *CombatAutomationService) calculateExperience(enemies []models.EnemyInfo, partySize int) int {
	// XP by CR mapping (simplified)
	xpByCR := map[string]int{
		"1/8": 25,
		"1/4": 50,
		"1/2": 100,
		"1":   200,
		"2":   450,
		"3":   700,
		"4":   1100,
		"5":   1800,
		"6":   2300,
		"7":   2900,
		"8":   3900,
		"9":   5000,
		"10":  5900,
	}

	totalXP := 0
	enemyCount := 0

	for _, enemy := range enemies {
		xp := xpByCR[enemy.CR]
		if xp == 0 {
			// Try to parse as integer for higher CRs
			var cr int
			_, _ = fmt.Sscanf(enemy.CR, "%d", &cr)
			if cr > 10 {
				xp = 5900 + (cr-10)*1000
			} else {
				xp = 200 // Default
			}
		}
		totalXP += xp * enemy.Count
		enemyCount += enemy.Count
	}

	// Apply encounter multiplier
	multiplier := 1.0
	if enemyCount >= 2 {
		multiplier = 1.5
	}
	if enemyCount >= 3 {
		multiplier = 2.0
	}
	if enemyCount >= 7 {
		multiplier = 2.5
	}
	if enemyCount >= 11 {
		multiplier = 3.0
	}
	if enemyCount >= 15 {
		multiplier = 4.0
	}

	// Adjust for party size
	if partySize < 3 {
		multiplier *= 1.5
	} else if partySize > 5 {
		multiplier *= 0.5
	}

	return int(float64(totalXP) * multiplier)
}

func (cas *CombatAutomationService) generateNarrativeSummary(
	outcome string,
	rounds int,
	terrainType string,
	partyLevel float64,
	encounterCR float64,
) string {
	narratives := map[string][]string{
		"decisive_victory": {
			"The party swiftly overwhelmed their foes with coordinated strikes and superior tactics.",
			"With barely a scratch, the adventurers dispatched their enemies in a display of martial prowess.",
			"The encounter was over almost before it began, the party's skill far exceeding the challenge.",
		},
		"victory": {
			"After a brief but intense skirmish, the party emerged victorious.",
			"The adventurers fought well, overcoming their foes through teamwork and determination.",
			"Though the enemies put up a fight, the party's strength proved superior.",
		},
		"costly_victory": {
			"The battle was hard-fought, with the party paying a steep price for their victory.",
			"Bloodied but unbowed, the adventurers managed to defeat their foes after a grueling combat.",
			"Victory came at a cost, with several party members bearing serious wounds.",
		},
		"retreat": {
			"Recognizing the danger, the party made a tactical withdrawal from the battlefield.",
			"The adventurers fought a retreating action, escaping with their lives if not their pride.",
			"Discretion proved the better part of valor as the party retreated from overwhelming odds.",
		},
		"defeat": {
			"The encounter proved too much for the party, forcing a desperate escape.",
			"Overwhelmed by their foes, the adventurers were routed from the field.",
			"The party suffered a crushing defeat, barely escaping with their lives.",
		},
	}

	narrative := narratives[outcome][rand.Intn(len(narratives[outcome]))]

	// Add terrain flavor
	if terrainType != "" {
		terrainFlavor := map[string]string{
			"forest":   " The dense foliage provided both cover and obstacles during the fight.",
			"dungeon":  " The cramped corridors limited mobility and tactical options.",
			"open":     " The open terrain allowed for fluid movement and ranged attacks.",
			"urban":    " The city streets and buildings created a complex battlefield.",
			"mountain": " The rocky terrain and elevation changes added complexity to the combat.",
		}

		if flavor, ok := terrainFlavor[terrainType]; ok {
			narrative += flavor
		}
	}

	// Add duration note
	narrative += fmt.Sprintf(" The combat lasted %d rounds.", rounds)

	return narrative
}

func (cas *CombatAutomationService) buildPartyComposition(characters []*models.Character) []byte {
	party := make([]map[string]interface{}, len(characters))

	for i, char := range characters {
		party[i] = map[string]interface{}{
			"id":    char.ID,
			"name":  char.Name,
			"class": char.Class,
			"level": char.Level,
			"hp":    char.MaxHitPoints,
		}
	}

	data, _ := json.Marshal(party)
	return data
}

func (cas *CombatAutomationService) buildEnemyComposition(enemies []models.EnemyInfo) []byte {
	data, _ := json.Marshal(enemies)
	return data
}

func (cas *CombatAutomationService) sortInitiativeEntries(entries []models.InitiativeEntry) {
	// Sort by initiative descending
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Initiative > entries[i].Initiative {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

func (cas *CombatAutomationService) resolveTies(entries []models.InitiativeEntry) {
	// When initiatives are tied, order by dexterity modifier (bonus)
	for i := 0; i < len(entries)-1; i++ {
		if entries[i].Initiative == entries[i+1].Initiative {
			if entries[i+1].Bonus > entries[i].Bonus {
				entries[i], entries[i+1] = entries[i+1], entries[i]
			}
		}
	}
}
