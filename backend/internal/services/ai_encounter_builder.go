package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

type EncounterRequest struct {
	PartyLevel       int      `json:"partyLevel"`
	PartySize        int      `json:"partySize"`
	PartyComposition []string `json:"partyComposition"` // ["fighter", "wizard", "cleric", "rogue"]
	Difficulty       string   `json:"difficulty"`       // easy, medium, hard, deadly
	EncounterType    string   `json:"encounterType"`    // combat, social, exploration, puzzle, hybrid
	Location         string   `json:"location"`         // dungeon, forest, city, etc.
	NarrativeContext string   `json:"narrativeContext"` // Story context
	SpecialRequests  string   `json:"specialRequests"`  // Any specific requests
}

type AIEncounterBuilder struct {
	llmProvider LLMProvider
}

func NewAIEncounterBuilder(provider LLMProvider) *AIEncounterBuilder {
	return &AIEncounterBuilder{
		llmProvider: provider,
	}
}

func (b *AIEncounterBuilder) GenerateEncounter(ctx context.Context, req *EncounterRequest) (*models.Encounter, error) {
	// Calculate XP budget based on party
	xpBudget := b.calculateXPBudget(req.PartyLevel, req.PartySize, req.Difficulty)

	// Build prompt for AI
	prompt := b.buildEncounterPrompt(req, xpBudget)

	systemPrompt := `You are a D&D 5th Edition encounter designer creating balanced, engaging encounters.
Your responses must be valid JSON matching the specified format exactly. Do not include any additional text or explanation outside the JSON.`

	// Generate encounter
	response, err := b.llmProvider.GenerateCompletion(ctx, prompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate encounter: %w", err)
	}

	// Parse response
	encounter, err := b.parseEncounterResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encounter: %w", err)
	}

	// Enhance with party-specific details
	b.enhanceEncounterForParty(encounter, req)

	// Calculate and adjust XP
	b.calculateEncounterXP(encounter)

	return encounter, nil
}

func (b *AIEncounterBuilder) buildEncounterPrompt(req *EncounterRequest, xpBudget int) string {
	return fmt.Sprintf(`Create a %s %s encounter for the following party:

Party Details:
- Level: %d
- Size: %d players
- Classes: %s
- XP Budget: %d (for %s difficulty)

Location: %s
Narrative Context: %s
Special Requests: %s

Create a dynamic, engaging encounter that:
1. Fits the narrative context and location
2. Challenges the party appropriately
3. Includes interesting tactical elements
4. Provides multiple ways to approach/resolve it
5. Has environmental features that can be used tactically
6. Includes enemy motivations and personalities

For combat encounters, include:
- Varied enemy types with different roles (damage, tank, support, controller)
- Intelligent enemy tactics based on their nature
- Environmental hazards or features
- Potential reinforcements or escalation
- Non-combat resolution options

Respond with a JSON object in this exact format:
{
  "name": "Encounter Name",
  "description": "Detailed description of the encounter setup",
  "enemies": [
    {
      "name": "Enemy Name",
      "type": "creature type",
      "challengeRating": 0.5,
      "quantity": 2,
      "role": "skirmisher",
      "hitPoints": 22,
      "armorClass": 15,
      "tactics": "Specific tactical behavior",
      "personalityTraits": ["trait1", "trait2"],
      "ideal": "What they believe in",
      "bond": "What they care about",
      "flaw": "Their weakness",
      "abilities": [
        {
          "name": "Ability Name",
          "description": "What it does"
        }
      ],
      "actions": [
        {
          "name": "Action Name",
          "description": "Action description",
          "attackBonus": 4,
          "damage": "1d6+2"
        }
      ]
    }
  ],
  "environmentalFeatures": [
    "Crumbling pillars provide half cover",
    "Narrow bridge over chasm requires DC 10 Acrobatics to cross quickly"
  ],
  "environmentalHazards": [
    {
      "name": "Hazard Name",
      "description": "What it is",
      "trigger": "When it activates",
      "effect": "What happens",
      "saveDC": 13,
      "damage": "2d6"
    }
  ],
  "terrainFeatures": [
    {
      "name": "Difficult Terrain",
      "description": "Rubble and debris",
      "effect": "Movement costs double",
      "location": "Eastern half of the room"
    }
  ],
  "tacticalInfo": {
    "generalStrategy": "Overall enemy strategy",
    "priorityTargets": ["Spellcasters", "Healers"],
    "positioning": "How enemies position themselves",
    "combatPhases": [
      {
        "name": "Opening",
        "trigger": "Combat starts",
        "tactics": "Rush the weakest-looking target"
      },
      {
        "name": "Desperate",
        "trigger": "Below 50%% forces",
        "tactics": "Fighting retreat to reinforcement location"
      }
    ],
    "retreatConditions": "When/how enemies flee"
  },
  "reinforcements": [
    {
      "round": 3,
      "trigger": "If combat lasts to round 3 or alarm is raised",
      "enemies": [
        {
          "name": "Reinforcement",
          "quantity": 2,
          "entrance": "Through the northern door"
        }
      ],
      "announcement": "You hear footsteps approaching from the north!"
    }
  ],
  "nonCombatOptions": {
    "social": [
      {
        "method": "Negotiation",
        "description": "Convince enemies to let you pass",
        "requirements": ["Speak their language", "Offer something valuable"],
        "dc": 15,
        "consequences": "They demand payment or a favor"
      }
    ],
    "stealth": [
      {
        "method": "Sneak past",
        "description": "Use shadows and timing",
        "requirements": ["Group stealth check"],
        "dc": 12,
        "consequences": "If failed, combat starts with surprise round for enemies"
      }
    ],
    "environmental": [
      {
        "method": "Cause distraction",
        "description": "Collapse unstable structure",
        "requirements": ["DC 15 Investigation to notice", "DC 12 Athletics to trigger"],
        "dc": 12,
        "consequences": "Enemies investigate noise, creating opening"
      }
    ]
  },
  "escapeRoutes": [
    {
      "direction": "Back the way you came",
      "difficulty": "Easy",
      "consequence": "None, but objective not completed"
    },
    {
      "direction": "Through the window",
      "difficulty": "Medium",
      "consequence": "Take 2d6 falling damage, but escape pursuit"
    }
  ],
  "objectives": [
    {
      "type": "primary",
      "description": "Defeat or bypass the guards",
      "xpReward": %d,
      "goldReward": %d
    },
    {
      "type": "optional", 
      "description": "Retrieve the stolen artifact",
      "xpReward": %d,
      "itemReward": "Potion of Healing"
    }
  ],
  "scalingOptions": {
    "easier": {
      "removeEnemies": ["Remove one enemy"],
      "hpModifier": -20,
      "damageModifier": -2
    },
    "harder": {
      "addEnemies": ["Add one more enemy"],
      "hpModifier": 20,
      "addHazards": ["Add environmental hazard"]
    }
  },
  "narrativeHooks": [
    "The enemies are searching for the same thing the party seeks",
    "One enemy has information about the main villain"
  ]
}

Make the encounter exciting, tactical, and memorable. Include personality for enemies and multiple resolution paths.`,
		req.EncounterType,
		req.Difficulty,
		req.PartyLevel,
		req.PartySize,
		strings.Join(req.PartyComposition, ", "),
		xpBudget,
		req.Difficulty,
		req.Location,
		req.NarrativeContext,
		req.SpecialRequests,
		xpBudget,
		xpBudget/10,
		xpBudget/4,
	)
}

func (b *AIEncounterBuilder) parseEncounterResponse(response string) (*models.Encounter, error) {
	// Extract JSON from response
	data, err := b.extractJSONFromResponse(response)
	if err != nil {
		return nil, err
	}

	// Build encounter model
	encounter := &models.Encounter{
		Name:        getString(data, "name"),
		Description: getString(data, "description"),
	}

	// Parse all components
	b.parseEnemies(data, encounter)
	b.parseEnvironmentalComponents(data, encounter)
	b.parseTacticalComponents(data, encounter)
	b.parseNonCombatOptions(data, encounter)
	b.parseEscapeAndStory(data, encounter)

	return encounter, nil
}

func (b *AIEncounterBuilder) extractJSONFromResponse(response string) (map[string]interface{}, error) {
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return data, nil
}

func (b *AIEncounterBuilder) parseEnemies(data map[string]interface{}, encounter *models.Encounter) {
	if enemies, ok := data["enemies"].([]interface{}); ok {
		for _, enemy := range enemies {
			if enemyMap, ok := enemy.(map[string]interface{}); ok {
				encounterEnemy := b.parseEnemy(enemyMap)
				encounter.Enemies = append(encounter.Enemies, encounterEnemy)
			}
		}
	}
}

func (b *AIEncounterBuilder) parseEnvironmentalComponents(data map[string]interface{}, encounter *models.Encounter) {
	// Parse environmental features
	if features, ok := data["environmentalFeatures"].([]interface{}); ok {
		for _, feature := range features {
			if featureStr, ok := feature.(string); ok {
				encounter.EnvironmentalFeatures = append(encounter.EnvironmentalFeatures, featureStr)
			}
		}
	}

	// Parse environmental hazards
	if hazards, ok := data["environmentalHazards"].([]interface{}); ok {
		for _, hazard := range hazards {
			if hazardMap, ok := hazard.(map[string]interface{}); ok {
				envHazard := models.EnvironmentalHazard{
					Name:        getString(hazardMap, "name"),
					Description: getString(hazardMap, "description"),
					Trigger:     getString(hazardMap, "trigger"),
					Effect:      getString(hazardMap, "effect"),
					SaveDC:      getInt(hazardMap, "saveDC"),
					Damage:      getString(hazardMap, "damage"),
				}
				encounter.EnvironmentalHazards = append(encounter.EnvironmentalHazards, envHazard)
			}
		}
	}
}

func (b *AIEncounterBuilder) parseTacticalComponents(data map[string]interface{}, encounter *models.Encounter) {
	// Parse tactical info
	if tactical, ok := data["tacticalInfo"].(map[string]interface{}); ok {
		encounter.EnemyTactics = b.parseTacticalInfo(tactical)
	}

	// Parse reinforcements
	if reinforcements, ok := data["reinforcements"].([]interface{}); ok {
		encounter.ReinforcementWaves = b.parseReinforcements(reinforcements)
	}
}

func (b *AIEncounterBuilder) parseNonCombatOptions(data map[string]interface{}, encounter *models.Encounter) {
	if nonCombat, ok := data["nonCombatOptions"].(map[string]interface{}); ok {
		encounter.SocialSolutions = b.parseSolutions(nonCombat["social"])
		encounter.StealthOptions = b.parseSolutions(nonCombat["stealth"])
		encounter.EnvironmentalSolutions = b.parseSolutions(nonCombat["environmental"])
	}
}

func (b *AIEncounterBuilder) parseEscapeAndStory(data map[string]interface{}, encounter *models.Encounter) {
	// Parse escape routes
	if escapes, ok := data["escapeRoutes"].([]interface{}); ok {
		for _, escape := range escapes {
			if escapeMap, ok := escape.(map[string]interface{}); ok {
				route := models.EscapeRoute{
					Direction:   getString(escapeMap, "direction"),
					Description: getString(escapeMap, "description"),
					Difficulty:  getString(escapeMap, "difficulty"),
					Consequence: getString(escapeMap, "consequence"),
				}
				encounter.EscapeRoutes = append(encounter.EscapeRoutes, route)
			}
		}
	}

	// Parse narrative hooks
	if hooks, ok := data["narrativeHooks"].([]interface{}); ok {
		for _, hook := range hooks {
			if hookStr, ok := hook.(string); ok {
				encounter.StoryHooks = append(encounter.StoryHooks, hookStr)
			}
		}
	}

	// Parse scaling options
	if scaling, ok := data["scalingOptions"].(map[string]interface{}); ok {
		encounter.ScalingOptions = b.parseScalingOptions(scaling)
	}
}

func (b *AIEncounterBuilder) parseEnemy(data map[string]interface{}) models.EncounterEnemy {
	enemy := b.parseEnemyBasicInfo(data)
	b.parseEnemyPersonalityTraits(data, &enemy)
	b.parseEnemyAbilities(data, &enemy)
	b.parseEnemyActions(data, &enemy)
	return enemy
}

func (b *AIEncounterBuilder) parseEnemyBasicInfo(data map[string]interface{}) models.EncounterEnemy {
	return models.EncounterEnemy{
		Name:            getString(data, "name"),
		Type:            getString(data, "type"),
		ChallengeRating: getFloat(data, "challengeRating"),
		Quantity:        getInt(data, "quantity"),
		Role:            getString(data, "role"),
		HitPoints:       getInt(data, "hitPoints"),
		ArmorClass:      getInt(data, "armorClass"),
		Tactics:         getString(data, "tactics"),
		Ideal:           getString(data, "ideal"),
		Bond:            getString(data, "bond"),
		Flaw:            getString(data, "flaw"),
		IsAlive:         true,
		MoraleThreshold: 50, // Default
	}
}

func (b *AIEncounterBuilder) parseEnemyPersonalityTraits(data map[string]interface{}, enemy *models.EncounterEnemy) {
	if traits, ok := data["personalityTraits"].([]interface{}); ok {
		for _, trait := range traits {
			if traitStr, ok := trait.(string); ok {
				enemy.PersonalityTraits = append(enemy.PersonalityTraits, traitStr)
			}
		}
	}
}

func (b *AIEncounterBuilder) parseEnemyAbilities(data map[string]interface{}, enemy *models.EncounterEnemy) {
	if abilities, ok := data["abilities"].([]interface{}); ok {
		for _, ability := range abilities {
			if abilityMap, ok := ability.(map[string]interface{}); ok {
				ab := models.Ability{
					Name:        getString(abilityMap, "name"),
					Description: getString(abilityMap, "description"),
					Recharge:    getString(abilityMap, "recharge"),
				}
				enemy.Abilities = append(enemy.Abilities, ab)
			}
		}
	}
}

func (b *AIEncounterBuilder) parseEnemyActions(data map[string]interface{}, enemy *models.EncounterEnemy) {
	if actions, ok := data["actions"].([]interface{}); ok {
		for _, action := range actions {
			if actionMap, ok := action.(map[string]interface{}); ok {
				act := models.Action{
					Name:        getString(actionMap, "name"),
					Description: getString(actionMap, "description"),
					AttackBonus: getInt(actionMap, "attackBonus"),
					Damage:      getString(actionMap, "damage"),
					SaveDC:      getInt(actionMap, "saveDC"),
					SaveType:    getString(actionMap, "saveType"),
				}
				enemy.Actions = append(enemy.Actions, act)
			}
		}
	}
}

func (b *AIEncounterBuilder) parseTacticalInfo(data map[string]interface{}) *models.TacticalInfo {
	tactical := &models.TacticalInfo{
		GeneralStrategy:   getString(data, "generalStrategy"),
		Positioning:       getString(data, "positioning"),
		RetreatConditions: getString(data, "retreatConditions"),
		SpecialTactics:    make(map[string]string),
	}

	// Parse priority targets
	if targets, ok := data["priorityTargets"].([]interface{}); ok {
		for _, target := range targets {
			if targetStr, ok := target.(string); ok {
				tactical.PriorityTargets = append(tactical.PriorityTargets, targetStr)
			}
		}
	}

	// Parse combat phases
	if phases, ok := data["combatPhases"].([]interface{}); ok {
		for _, phase := range phases {
			if phaseMap, ok := phase.(map[string]interface{}); ok {
				combatPhase := models.CombatPhase{
					Name:    getString(phaseMap, "name"),
					Trigger: getString(phaseMap, "trigger"),
					Tactics: getString(phaseMap, "tactics"),
				}
				tactical.CombatPhases = append(tactical.CombatPhases, combatPhase)
			}
		}
	}

	return tactical
}

func (b *AIEncounterBuilder) parseSolutions(data interface{}) []models.Solution {
	var solutions []models.Solution

	solutionList, ok := data.([]interface{})
	if !ok {
		return solutions
	}

	for _, solution := range solutionList {
		if sol := b.parseSingleSolution(solution); sol != nil {
			solutions = append(solutions, *sol)
		}
	}

	return solutions
}

func (b *AIEncounterBuilder) parseSingleSolution(solution interface{}) *models.Solution {
	solutionMap, ok := solution.(map[string]interface{})
	if !ok {
		return nil
	}

	sol := &models.Solution{
		Method:       getString(solutionMap, "method"),
		Description:  getString(solutionMap, "description"),
		DC:           getInt(solutionMap, "dc"),
		Consequences: getString(solutionMap, "consequences"),
	}

	b.parseSolutionRequirements(solutionMap, sol)
	return sol
}

func (b *AIEncounterBuilder) parseSolutionRequirements(solutionMap map[string]interface{}, sol *models.Solution) {
	reqs, ok := solutionMap["requirements"].([]interface{})
	if !ok {
		return
	}

	for _, req := range reqs {
		if reqStr, ok := req.(string); ok {
			sol.Requirements = append(sol.Requirements, reqStr)
		}
	}
}

func (b *AIEncounterBuilder) parseReinforcements(data []interface{}) []models.ReinforcementWave {
	var waves []models.ReinforcementWave

	for _, wave := range data {
		if waveMap, ok := wave.(map[string]interface{}); ok {
			reinforcement := models.ReinforcementWave{
				Round:        getInt(waveMap, "round"),
				Trigger:      getString(waveMap, "trigger"),
				Announcement: getString(waveMap, "announcement"),
			}

			// Parse reinforcement enemies
			if enemies, ok := waveMap["enemies"].([]interface{}); ok {
				for _, enemy := range enemies {
					if enemyMap, ok := enemy.(map[string]interface{}); ok {
						// Create simplified enemy for reinforcement
						reinforcementEnemy := models.EncounterEnemy{
							Name:     getString(enemyMap, "name"),
							Quantity: getInt(enemyMap, "quantity"),
						}
						reinforcement.Enemies = append(reinforcement.Enemies, reinforcementEnemy)
						reinforcement.Entrance = getString(enemyMap, "entrance")
					}
				}
			}

			waves = append(waves, reinforcement)
		}
	}

	return waves
}

func (b *AIEncounterBuilder) parseScalingOptions(data map[string]interface{}) *models.ScalingOptions {
	options := &models.ScalingOptions{}

	if easier, ok := data["easier"].(map[string]interface{}); ok {
		options.Easy = b.parseScalingAdjustment(easier)
	}

	if harder, ok := data["harder"].(map[string]interface{}); ok {
		options.Hard = b.parseScalingAdjustment(harder)
	}

	// Generate medium and deadly based on easy/hard
	options.Medium = models.ScalingAdjustment{} // No changes for medium
	options.Deadly = models.ScalingAdjustment{
		HPModifier:     options.Hard.HPModifier * 2,
		DamageModifier: options.Hard.DamageModifier * 2,
		AddEnemies:     append(options.Hard.AddEnemies, options.Hard.AddEnemies...),
		AddHazards:     options.Hard.AddHazards,
	}

	return options
}

func (b *AIEncounterBuilder) parseScalingAdjustment(data map[string]interface{}) models.ScalingAdjustment {
	adj := models.ScalingAdjustment{
		HPModifier:     getInt(data, "hpModifier"),
		DamageModifier: getInt(data, "damageModifier"),
	}

	// Parse arrays
	adj.AddEnemies = parseStringArray(data, "addEnemies")
	adj.RemoveEnemies = parseStringArray(data, "removeEnemies")
	adj.AddHazards = parseStringArray(data, "addHazards")
	adj.AddTerrain = parseStringArray(data, "addTerrain")
	adj.AddObjectives = parseStringArray(data, "addObjectives")

	return adj
}

func (b *AIEncounterBuilder) calculateXPBudget(level, size int, difficulty string) int {
	// XP thresholds per character level (easy, medium, hard, deadly)
	thresholds := map[int][4]int{
		1:  {25, 50, 75, 100},
		2:  {50, 100, 150, 200},
		3:  {75, 150, 225, 400},
		4:  {125, 250, 375, 500},
		5:  {250, 500, 750, 1100},
		6:  {300, 600, 900, 1400},
		7:  {350, 750, 1100, 1700},
		8:  {450, 900, 1400, 2100},
		9:  {550, 1100, 1600, 2400},
		10: {600, 1200, 1900, 2800},
		11: {800, 1600, 2400, 3600},
		12: {1000, 2000, 3000, 4500},
		13: {1100, 2200, 3400, 5100},
		14: {1250, 2500, 3800, 5700},
		15: {1400, 2800, 4300, 6400},
		16: {1600, 3200, 4800, 7200},
		17: {2000, 3900, 5900, 8800},
		18: {2100, 4200, 6300, 9500},
		19: {2400, 4900, 7300, 10900},
		20: {2800, 5700, 8500, 12700},
	}

	difficultyIndex := map[string]int{
		constants.DifficultyEasy:   0,
		constants.DifficultyMedium: 1,
		constants.DifficultyHard:   2,
		constants.DifficultyDeadly: 3,
	}

	idx := difficultyIndex[difficulty]
	if idx == 0 && difficulty != constants.DifficultyEasy {
		idx = 1 // Default to medium
	}

	threshold := thresholds[level][idx]
	return threshold * size
}

func (b *AIEncounterBuilder) calculateEncounterXP(encounter *models.Encounter) {
	totalXP, enemyCount := b.calculateBaseXP(encounter)
	encounter.TotalXP = totalXP

	multiplier := b.getXPMultiplier(enemyCount, encounter.PartySize)
	encounter.AdjustedXP = int(float64(totalXP) * multiplier)
}

func (b *AIEncounterBuilder) calculateBaseXP(encounter *models.Encounter) (int, int) {
	// CR to XP mapping
	crToXP := map[float64]int{
		0:     10,
		0.125: 25,
		0.25:  50,
		0.5:   100,
		1:     200,
		2:     450,
		3:     700,
		4:     1100,
		5:     1800,
		6:     2300,
		7:     2900,
		8:     3900,
		9:     5000,
		10:    5900,
		11:    7200,
		12:    8400,
		13:    10000,
		14:    11500,
		15:    13000,
		16:    15000,
		17:    18000,
		18:    20000,
		19:    22000,
		20:    25000,
	}

	totalXP := 0
	enemyCount := 0

	for i := range encounter.Enemies {
		if xp, ok := crToXP[encounter.Enemies[i].ChallengeRating]; ok {
			totalXP += xp * encounter.Enemies[i].Quantity
			enemyCount += encounter.Enemies[i].Quantity
		}
	}

	return totalXP, enemyCount
}

func (b *AIEncounterBuilder) getXPMultiplier(enemyCount, partySize int) float64 {
	// Calculate base multiplier based on enemy count
	multiplier := 1.0
	switch {
	case enemyCount == 1:
		multiplier = 1.0
	case enemyCount == 2:
		multiplier = 1.5
	case enemyCount >= 3 && enemyCount <= 6:
		multiplier = 2.0
	case enemyCount >= 7 && enemyCount <= 10:
		multiplier = 2.5
	case enemyCount >= 11 && enemyCount <= 14:
		multiplier = 3.0
	case enemyCount >= 15:
		multiplier = 4.0
	}

	// Adjust multiplier for party size
	if partySize < 3 {
		multiplier *= 1.5
	} else if partySize > 5 {
		multiplier *= 0.5
	}

	return multiplier
}

func (b *AIEncounterBuilder) enhanceEncounterForParty(encounter *models.Encounter, req *EncounterRequest) {
	encounter.PartyLevel = req.PartyLevel
	encounter.PartySize = req.PartySize
	encounter.Location = req.Location
	encounter.NarrativeContext = req.NarrativeContext
	encounter.EncounterType = req.EncounterType
	encounter.Difficulty = req.Difficulty

	// Store party composition
	composition := make(map[string]interface{})
	classCounts := make(map[string]int)
	for _, class := range req.PartyComposition {
		classCounts[class]++
	}
	composition["classes"] = classCounts
	composition["totalSize"] = req.PartySize
	encounter.PartyComposition = composition

	// Calculate average CR for the encounter
	totalCR := 0.0
	enemyCount := 0
	for i := range encounter.Enemies {
		totalCR += encounter.Enemies[i].ChallengeRating * float64(encounter.Enemies[i].Quantity)
		enemyCount += encounter.Enemies[i].Quantity
	}
	if enemyCount > 0 {
		encounter.ChallengeRating = totalCR / float64(enemyCount)
	}

	// Add party-specific tactical considerations
	b.addPartySpecificTactics(encounter, req.PartyComposition)
}

func (b *AIEncounterBuilder) addPartySpecificTactics(encounter *models.Encounter, classes []string) {
	if encounter.EnemyTactics == nil {
		encounter.EnemyTactics = &models.TacticalInfo{
			SpecialTactics: make(map[string]string),
		}
	}

	// Add tactics based on party composition
	hasHealer := false
	hasCaster := false
	hasTank := false

	for _, class := range classes {
		switch class {
		case constants.ClassCleric, constants.ClassDruid, constants.ClassBard:
			hasHealer = true
		case constants.ClassWizard, constants.ClassSorcerer, constants.ClassWarlock:
			hasCaster = true
		case constants.ClassFighter, constants.ClassPaladin, constants.ClassBarbarian:
			hasTank = true
		}
	}

	if hasHealer {
		encounter.EnemyTactics.SpecialTactics["vs_healer"] = "Focus fire on the healer when they're exposed"
	}
	if hasCaster {
		encounter.EnemyTactics.SpecialTactics["vs_caster"] = "Use cover and spread out to avoid area spells"
	}
	if hasTank {
		encounter.EnemyTactics.SpecialTactics["vs_tank"] = "Use mobility to bypass front-line fighters"
	}
}

func (b *AIEncounterBuilder) GenerateTacticalSuggestion(ctx context.Context, encounter *models.Encounter, situation string) (string, error) {
	systemPrompt := "You are a D&D tactical advisor providing concise, actionable tactical suggestions for combat encounters."

	prompt := fmt.Sprintf(`Provide a specific tactical suggestion for the enemies in this situation:

Encounter: %s
Current Situation: %s
Enemy Forces: %d enemies remaining
Enemy Strategy: %s

Provide a brief (2-3 sentences) tactical suggestion for what the enemies should do next. Consider their personalities, abilities, and the current tactical situation. Make it specific and actionable.`,
		encounter.Name,
		situation,
		len(encounter.Enemies),
		encounter.EnemyTactics.GeneralStrategy,
	)

	response, err := b.llmProvider.GenerateCompletion(ctx, prompt, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate tactical suggestion: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// Helper functions
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	if val, ok := data[key].(int); ok {
		return val
	}
	return 0
}

func getFloat(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	if val, ok := data[key].(int); ok {
		return float64(val)
	}
	return 0
}

func parseStringArray(data map[string]interface{}, key string) []string {
	var result []string
	if arr, ok := data[key].([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
	}
	return result
}
