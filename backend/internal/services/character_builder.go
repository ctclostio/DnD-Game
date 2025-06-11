package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/your-username/dnd-game/backend/internal/models"
)

type CharacterBuilder struct {
	dataPath string
}

type RaceData struct {
	Name             string                   `json:"name"`
	AbilityIncreases map[string]int           `json:"abilityScoreIncrease"`
	Size             string                   `json:"size"`
	Speed            int                      `json:"speed"`
	Languages        []string                 `json:"languages"`
	Traits           []map[string]interface{} `json:"traits"`
	Subraces         []SubraceData            `json:"subraces"`
}

type SubraceData struct {
	Name             string                   `json:"name"`
	AbilityIncreases map[string]int           `json:"abilityScoreIncrease"`
	Traits           []map[string]interface{} `json:"traits"`
}

type ClassData struct {
	Name                     string                   `json:"name"`
	HitDice                  string                   `json:"hitDice"`
	PrimaryAbility           string                   `json:"primaryAbility"`
	SavingThrowProficiencies []string                 `json:"savingThrowProficiencies"`
	SkillChoices             map[string]interface{}   `json:"skillChoices"`
	Features                 map[string]interface{}   `json:"features"`
	Spellcasting             map[string]interface{}   `json:"spellcasting"`
	Subclasses               []map[string]interface{} `json:"subclasses"`
}

type BackgroundData struct {
	Name               string                 `json:"name"`
	SkillProficiencies []string               `json:"skillProficiencies"`
	Languages          int                    `json:"languages"`
	ToolProficiencies  []string               `json:"toolProficiencies"`
	Equipment          []string               `json:"equipment"`
	Feature            map[string]interface{} `json:"feature"`
}

// validateFileName validates a filename to prevent path traversal attacks
func validateFileName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	
	// Check for path traversal attempts
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return errors.New("invalid characters in name")
	}
	
	// Only allow alphanumeric, dash, and underscore
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return errors.New("name contains invalid characters")
	}
	
	return nil
}

func NewCharacterBuilder(dataPath string) *CharacterBuilder {
	return &CharacterBuilder{
		dataPath: dataPath,
	}
}

func (cb *CharacterBuilder) GetAvailableOptions() (map[string]interface{}, error) {
	races, err := cb.loadRaces()
	if err != nil {
		return nil, err
	}

	classes, err := cb.loadClasses()
	if err != nil {
		return nil, err
	}

	backgrounds, err := cb.loadBackgrounds()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"races":       races,
		"classes":     classes,
		"backgrounds": backgrounds,
		"abilityScoreMethods": []string{
			"standard_array", // 15,14,13,12,10,8
			"point_buy",      // 27 points
			"roll_4d6",       // Roll 4d6, drop lowest
			"custom",         // Custom input
		},
	}, nil
}

func (cb *CharacterBuilder) BuildCharacter(params map[string]interface{}) (*models.Character, error) {
	// Extract parameters
	race, _ := params["race"].(string)
	customRaceID, _ := params["customRaceId"].(string)
	customRaceStats, hasCustomRace := params["customRaceStats"].(map[string]interface{})
	subrace, _ := params["subrace"].(string)
	class := params["class"].(string)
	background := params["background"].(string)
	name := params["name"].(string)
	alignment := params["alignment"].(string)
	abilityScores := params["abilityScores"].(map[string]int)

	// Create character
	character := &models.Character{
		Name:       name,
		Race:       race,
		Subrace:    subrace,
		Class:      class,
		Background: background,
		Alignment:  alignment,
		Level:      1,
	}

	// Handle custom race vs standard race
	var raceData *RaceData
	var err error

	if hasCustomRace && customRaceID != "" {
		// Use custom race data
		character.Race = customRaceStats["name"].(string)
		character.CustomRaceID = &customRaceID
		raceData = cb.convertCustomRaceToRaceData(customRaceStats)
	} else {
		// Load standard race data
		raceData, err = cb.loadRaceData(race)
		if err != nil {
			return nil, err
		}
	}

	classData, err := cb.loadClassData(class)
	if err != nil {
		return nil, err
	}

	backgroundData, err := cb.loadBackgroundData(background)
	if err != nil {
		return nil, err
	}

	// Apply ability scores and racial modifiers
	character.Attributes = cb.calculateFinalAbilityScores(abilityScores, raceData, subrace)

	// Calculate derived stats
	character.ProficiencyBonus = cb.calculateProficiencyBonus(character.Level)
	character.Initiative = cb.calculateModifier(character.Attributes.Dexterity)
	character.Speed = raceData.Speed

	// Apply class features
	cb.applyClassFeatures(character, classData)

	// Apply racial features
	cb.applyRacialFeatures(character, raceData, subrace)

	// Apply background
	cb.applyBackground(character, backgroundData)

	// Calculate saving throws
	character.SavingThrows = cb.calculateSavingThrows(character, classData)

	// Calculate skills
	character.Skills = cb.calculateSkills(character)

	return character, nil
}

func (cb *CharacterBuilder) RollAbilityScores(method string) (map[string]int, error) {
	scores := make(map[string]int)
	abilities := []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}

	switch method {
	case "standard_array":
		standardArray := []int{15, 14, 13, 12, 10, 8}
		// User would assign these values
		for i, ability := range abilities {
			scores[ability] = standardArray[i]
		}

	case "roll_4d6":
		for _, ability := range abilities {
			scores[ability] = cb.roll4d6DropLowest()
		}

	case "point_buy":
		// Start with all 8s (costs 0 points)
		for _, ability := range abilities {
			scores[ability] = 8
		}
		// User would spend 27 points to increase

	default:
		return nil, fmt.Errorf("invalid ability score method: %s", method)
	}

	return scores, nil
}

func (cb *CharacterBuilder) roll4d6DropLowest() int {
	dice := make([]int, 4)
	for i := 0; i < 4; i++ {
		dice[i] = rand.Intn(6) + 1
	}
	sort.Ints(dice)
	// Sum the highest 3
	return dice[1] + dice[2] + dice[3]
}

func (cb *CharacterBuilder) calculateModifier(score int) int {
	return (score - 10) / 2
}

func (cb *CharacterBuilder) calculateProficiencyBonus(level int) int {
	return ((level - 1) / 4) + 2
}

func (cb *CharacterBuilder) calculateFinalAbilityScores(base map[string]int, raceData *RaceData, subrace string) models.Attributes {
	// Apply racial modifiers
	for ability, increase := range raceData.AbilityIncreases {
		base[strings.ToLower(ability)] += increase
	}

	// Apply subrace modifiers if applicable
	if subrace != "" {
		for _, sr := range raceData.Subraces {
			if sr.Name == subrace {
				for ability, increase := range sr.AbilityIncreases {
					base[strings.ToLower(ability)] += increase
				}
				break
			}
		}
	}

	return models.Attributes{
		Strength:     base["strength"],
		Dexterity:    base["dexterity"],
		Constitution: base["constitution"],
		Intelligence: base["intelligence"],
		Wisdom:       base["wisdom"],
		Charisma:     base["charisma"],
	}
}

func (cb *CharacterBuilder) calculateSavingThrows(character *models.Character, classData *ClassData) models.SavingThrows {
	saves := models.SavingThrows{}
	profBonus := character.ProficiencyBonus

	// Strength
	saves.Strength.Modifier = cb.calculateModifier(character.Attributes.Strength)
	saves.Strength.Proficiency = cb.contains(classData.SavingThrowProficiencies, "Strength")
	if saves.Strength.Proficiency {
		saves.Strength.Modifier += profBonus
	}

	// Dexterity
	saves.Dexterity.Modifier = cb.calculateModifier(character.Attributes.Dexterity)
	saves.Dexterity.Proficiency = cb.contains(classData.SavingThrowProficiencies, "Dexterity")
	if saves.Dexterity.Proficiency {
		saves.Dexterity.Modifier += profBonus
	}

	// Constitution
	saves.Constitution.Modifier = cb.calculateModifier(character.Attributes.Constitution)
	saves.Constitution.Proficiency = cb.contains(classData.SavingThrowProficiencies, "Constitution")
	if saves.Constitution.Proficiency {
		saves.Constitution.Modifier += profBonus
	}

	// Intelligence
	saves.Intelligence.Modifier = cb.calculateModifier(character.Attributes.Intelligence)
	saves.Intelligence.Proficiency = cb.contains(classData.SavingThrowProficiencies, "Intelligence")
	if saves.Intelligence.Proficiency {
		saves.Intelligence.Modifier += profBonus
	}

	// Wisdom
	saves.Wisdom.Modifier = cb.calculateModifier(character.Attributes.Wisdom)
	saves.Wisdom.Proficiency = cb.contains(classData.SavingThrowProficiencies, "Wisdom")
	if saves.Wisdom.Proficiency {
		saves.Wisdom.Modifier += profBonus
	}

	// Charisma
	saves.Charisma.Modifier = cb.calculateModifier(character.Attributes.Charisma)
	saves.Charisma.Proficiency = cb.contains(classData.SavingThrowProficiencies, "Charisma")
	if saves.Charisma.Proficiency {
		saves.Charisma.Modifier += profBonus
	}

	return saves
}

func (cb *CharacterBuilder) calculateSkills(character *models.Character) []models.Skill {
	// This would be expanded to calculate all skills based on proficiencies
	// For now, return empty slice
	return []models.Skill{}
}

func (cb *CharacterBuilder) applyClassFeatures(character *models.Character, classData *ClassData) {
	// Parse hit dice
	character.HitDice = classData.HitDice

	// Calculate HP (max at level 1)
	var baseHP int
	switch classData.HitDice {
	case "1d6":
		baseHP = 6
	case "1d8":
		baseHP = 8
	case "1d10":
		baseHP = 10
	case "1d12":
		baseHP = 12
	}
	character.MaxHitPoints = baseHP + cb.calculateModifier(character.Attributes.Constitution)
	character.HitPoints = character.MaxHitPoints

	// Initialize spell slots for spellcasting classes
	if classData.Spellcasting != nil && len(classData.Spellcasting) > 0 {
		// Extract spellcasting ability
		if ability, ok := classData.Spellcasting["ability"].(string); ok {
			character.Spells.SpellcastingAbility = ability

			// Calculate spell save DC and attack bonus
			abilityMod := cb.getAbilityModifier(character, ability)
			character.Spells.SpellSaveDC = 8 + character.ProficiencyBonus + abilityMod
			character.Spells.SpellAttackBonus = character.ProficiencyBonus + abilityMod
		}

		// Initialize spell slots directly
		character.Spells.SpellSlots = InitializeSpellSlots(character.Class, character.Level)

		// Set cantrips known if applicable
		if cantripsKnown, ok := classData.Spellcasting["cantripsKnown"].([]interface{}); ok {
			for _, levelData := range cantripsKnown {
				if levelMap, ok := levelData.(map[string]interface{}); ok {
					if level, ok := levelMap["level"].(float64); ok && int(level) == character.Level {
						if known, ok := levelMap["known"].(float64); ok {
							character.Spells.CantripsKnown = int(known)
							break
						}
					}
				}
			}
		}
	}

	// TODO: Apply other class features based on level
}

func (cb *CharacterBuilder) applyRacialFeatures(character *models.Character, raceData *RaceData, subrace string) {
	// Apply racial traits
	character.Proficiencies.Languages = append(character.Proficiencies.Languages, raceData.Languages...)

	// TODO: Apply other racial features
}

func (cb *CharacterBuilder) applyBackground(character *models.Character, backgroundData *BackgroundData) {
	// Apply background proficiencies
	// TODO: Add skill proficiencies from background

	// TODO: Apply other background features
}

// Helper functions

func (cb *CharacterBuilder) loadRaces() ([]string, error) {
	files, err := os.ReadDir(filepath.Join(cb.dataPath, "races"))
	if err != nil {
		return nil, err
	}

	var races []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			races = append(races, strings.TrimSuffix(file.Name(), ".json"))
		}
	}
	return races, nil
}

func (cb *CharacterBuilder) loadClasses() ([]string, error) {
	files, err := os.ReadDir(filepath.Join(cb.dataPath, "classes"))
	if err != nil {
		return nil, err
	}

	var classes []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			classes = append(classes, strings.TrimSuffix(file.Name(), ".json"))
		}
	}
	return classes, nil
}

func (cb *CharacterBuilder) loadBackgrounds() ([]string, error) {
	files, err := os.ReadDir(filepath.Join(cb.dataPath, "backgrounds"))
	if err != nil {
		return nil, err
	}

	var backgrounds []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			backgrounds = append(backgrounds, strings.TrimSuffix(file.Name(), ".json"))
		}
	}
	return backgrounds, nil
}

func (cb *CharacterBuilder) loadRaceData(race string) (*RaceData, error) {
	// Validate input to prevent path traversal
	if err := validateFileName(race); err != nil {
		return nil, fmt.Errorf("invalid race name: %w", err)
	}
	
	data, err := os.ReadFile(filepath.Join(cb.dataPath, "races", race+".json"))
	if err != nil {
		return nil, err
	}

	var raceData RaceData
	if err := json.Unmarshal(data, &raceData); err != nil {
		return nil, err
	}

	return &raceData, nil
}

func (cb *CharacterBuilder) loadClassData(class string) (*ClassData, error) {
	// Validate input to prevent path traversal
	if err := validateFileName(class); err != nil {
		return nil, fmt.Errorf("invalid class name: %w", err)
	}
	
	data, err := os.ReadFile(filepath.Join(cb.dataPath, "classes", class+".json"))
	if err != nil {
		return nil, err
	}

	var classData ClassData
	if err := json.Unmarshal(data, &classData); err != nil {
		return nil, err
	}

	return &classData, nil
}

func (cb *CharacterBuilder) loadBackgroundData(background string) (*BackgroundData, error) {
	// Validate input to prevent path traversal
	if err := validateFileName(background); err != nil {
		return nil, fmt.Errorf("invalid background name: %w", err)
	}
	
	data, err := os.ReadFile(filepath.Join(cb.dataPath, "backgrounds", background+".json"))
	if err != nil {
		return nil, err
	}

	var backgroundData BackgroundData
	if err := json.Unmarshal(data, &backgroundData); err != nil {
		return nil, err
	}

	return &backgroundData, nil
}

func (cb *CharacterBuilder) convertCustomRaceToRaceData(customRaceStats map[string]interface{}) *RaceData {
	// Convert custom race stats to RaceData format
	raceData := &RaceData{
		Name:             customRaceStats["name"].(string),
		AbilityIncreases: make(map[string]int),
		Size:             customRaceStats["size"].(string),
		Speed:            int(customRaceStats["speed"].(float64)),
		Languages:        []string{},
		Traits:           []map[string]interface{}{},
	}

	// Convert ability score increases
	if asi, ok := customRaceStats["abilityScoreIncreases"].(map[string]interface{}); ok {
		for ability, increase := range asi {
			if val, ok := increase.(float64); ok {
				raceData.AbilityIncreases[ability] = int(val)
			}
		}
	}

	// Convert languages
	if languages, ok := customRaceStats["languages"].([]interface{}); ok {
		for _, lang := range languages {
			if langStr, ok := lang.(string); ok {
				raceData.Languages = append(raceData.Languages, langStr)
			}
		}
	}

	// Convert traits
	if traits, ok := customRaceStats["traits"].([]interface{}); ok {
		for _, trait := range traits {
			if traitMap, ok := trait.(map[string]interface{}); ok {
				raceData.Traits = append(raceData.Traits, traitMap)
			}
		}
	}

	// Add additional features from custom race
	// Darkvision
	if darkvision, ok := customRaceStats["darkvision"].(float64); ok && darkvision > 0 {
		raceData.Traits = append(raceData.Traits, map[string]interface{}{
			"name":        "Darkvision",
			"description": fmt.Sprintf("You can see in dim light within %d feet as if it were bright light, and in darkness as if it were dim light.", int(darkvision)),
		})
	}

	// Resistances
	if resistances, ok := customRaceStats["resistances"].([]interface{}); ok && len(resistances) > 0 {
		resistanceList := []string{}
		for _, res := range resistances {
			if resStr, ok := res.(string); ok {
				resistanceList = append(resistanceList, resStr)
			}
		}
		if len(resistanceList) > 0 {
			raceData.Traits = append(raceData.Traits, map[string]interface{}{
				"name":        "Damage Resistance",
				"description": fmt.Sprintf("You have resistance to %s damage.", strings.Join(resistanceList, ", ")),
			})
		}
	}

	// Immunities
	if immunities, ok := customRaceStats["immunities"].([]interface{}); ok && len(immunities) > 0 {
		immunityList := []string{}
		for _, imm := range immunities {
			if immStr, ok := imm.(string); ok {
				immunityList = append(immunityList, immStr)
			}
		}
		if len(immunityList) > 0 {
			raceData.Traits = append(raceData.Traits, map[string]interface{}{
				"name":        "Damage Immunity",
				"description": fmt.Sprintf("You are immune to %s damage.", strings.Join(immunityList, ", ")),
			})
		}
	}

	// Skill proficiencies
	if skills, ok := customRaceStats["skillProficiencies"].([]interface{}); ok && len(skills) > 0 {
		skillList := []string{}
		for _, skill := range skills {
			if skillStr, ok := skill.(string); ok {
				skillList = append(skillList, skillStr)
			}
		}
		if len(skillList) > 0 {
			raceData.Traits = append(raceData.Traits, map[string]interface{}{
				"name":        "Skill Proficiencies",
				"description": fmt.Sprintf("You gain proficiency in %s.", strings.Join(skillList, " and ")),
			})
		}
	}

	return raceData
}

func (cb *CharacterBuilder) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (cb *CharacterBuilder) getAbilityModifier(character *models.Character, ability string) int {
	switch strings.ToLower(ability) {
	case "strength":
		return cb.calculateModifier(character.Attributes.Strength)
	case "dexterity":
		return cb.calculateModifier(character.Attributes.Dexterity)
	case "constitution":
		return cb.calculateModifier(character.Attributes.Constitution)
	case "intelligence":
		return cb.calculateModifier(character.Attributes.Intelligence)
	case "wisdom":
		return cb.calculateModifier(character.Attributes.Wisdom)
	case "charisma":
		return cb.calculateModifier(character.Attributes.Charisma)
	default:
		return 0
	}
}

// InitializeSpellSlots creates spell slots based on class and level
func InitializeSpellSlots(class string, level int) []models.SpellSlot {
	// Spell slots by level for full casters (wizard, cleric, druid, bard, sorcerer)
	fullCasterSlots := map[int][]int{
		1:  {2, 0, 0, 0, 0, 0, 0, 0, 0},
		2:  {3, 0, 0, 0, 0, 0, 0, 0, 0},
		3:  {4, 2, 0, 0, 0, 0, 0, 0, 0},
		4:  {4, 3, 0, 0, 0, 0, 0, 0, 0},
		5:  {4, 3, 2, 0, 0, 0, 0, 0, 0},
		6:  {4, 3, 3, 0, 0, 0, 0, 0, 0},
		7:  {4, 3, 3, 1, 0, 0, 0, 0, 0},
		8:  {4, 3, 3, 2, 0, 0, 0, 0, 0},
		9:  {4, 3, 3, 3, 1, 0, 0, 0, 0},
		10: {4, 3, 3, 3, 2, 0, 0, 0, 0},
		11: {4, 3, 3, 3, 2, 1, 0, 0, 0},
		12: {4, 3, 3, 3, 2, 1, 0, 0, 0},
		13: {4, 3, 3, 3, 2, 1, 1, 0, 0},
		14: {4, 3, 3, 3, 2, 1, 1, 0, 0},
		15: {4, 3, 3, 3, 2, 1, 1, 1, 0},
		16: {4, 3, 3, 3, 2, 1, 1, 1, 0},
		17: {4, 3, 3, 3, 2, 1, 1, 1, 1},
		18: {4, 3, 3, 3, 3, 1, 1, 1, 1},
		19: {4, 3, 3, 3, 3, 2, 1, 1, 1},
		20: {4, 3, 3, 3, 3, 2, 2, 1, 1},
	}

	// Half casters (ranger, paladin) get spells at level 2
	halfCasterSlots := map[int][]int{
		1:  {0, 0, 0, 0, 0, 0, 0, 0, 0},
		2:  {2, 0, 0, 0, 0, 0, 0, 0, 0},
		3:  {3, 0, 0, 0, 0, 0, 0, 0, 0},
		4:  {3, 0, 0, 0, 0, 0, 0, 0, 0},
		5:  {4, 2, 0, 0, 0, 0, 0, 0, 0},
		6:  {4, 2, 0, 0, 0, 0, 0, 0, 0},
		7:  {4, 3, 0, 0, 0, 0, 0, 0, 0},
		8:  {4, 3, 0, 0, 0, 0, 0, 0, 0},
		9:  {4, 3, 2, 0, 0, 0, 0, 0, 0},
		10: {4, 3, 2, 0, 0, 0, 0, 0, 0},
		11: {4, 3, 3, 0, 0, 0, 0, 0, 0},
		12: {4, 3, 3, 0, 0, 0, 0, 0, 0},
		13: {4, 3, 3, 1, 0, 0, 0, 0, 0},
		14: {4, 3, 3, 1, 0, 0, 0, 0, 0},
		15: {4, 3, 3, 2, 0, 0, 0, 0, 0},
		16: {4, 3, 3, 2, 0, 0, 0, 0, 0},
		17: {4, 3, 3, 3, 1, 0, 0, 0, 0},
		18: {4, 3, 3, 3, 1, 0, 0, 0, 0},
		19: {4, 3, 3, 3, 2, 0, 0, 0, 0},
		20: {4, 3, 3, 3, 2, 0, 0, 0, 0},
	}

	// Warlock has unique spell slot progression (Pact Magic)
	warlockSlots := map[int][]int{
		1:  {1, 0, 0, 0, 0, 0, 0, 0, 0},
		2:  {2, 0, 0, 0, 0, 0, 0, 0, 0},
		3:  {0, 2, 0, 0, 0, 0, 0, 0, 0},
		4:  {0, 2, 0, 0, 0, 0, 0, 0, 0},
		5:  {0, 0, 2, 0, 0, 0, 0, 0, 0},
		6:  {0, 0, 2, 0, 0, 0, 0, 0, 0},
		7:  {0, 0, 0, 2, 0, 0, 0, 0, 0},
		8:  {0, 0, 0, 2, 0, 0, 0, 0, 0},
		9:  {0, 0, 0, 0, 2, 0, 0, 0, 0},
		10: {0, 0, 0, 0, 2, 0, 0, 0, 0},
		11: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		12: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		13: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		14: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		15: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		16: {0, 0, 0, 0, 3, 0, 0, 0, 0},
		17: {0, 0, 0, 0, 4, 0, 0, 0, 0},
		18: {0, 0, 0, 0, 4, 0, 0, 0, 0},
		19: {0, 0, 0, 0, 4, 0, 0, 0, 0},
		20: {0, 0, 0, 0, 4, 0, 0, 0, 0},
	}

	var slotsTable map[int][]int
	switch strings.ToLower(class) {
	case "wizard", "cleric", "druid", "bard", "sorcerer":
		slotsTable = fullCasterSlots
	case "ranger", "paladin":
		slotsTable = halfCasterSlots
	case "warlock":
		slotsTable = warlockSlots
	default:
		// Non-casters have no spell slots
		return []models.SpellSlot{}
	}

	slots := []models.SpellSlot{}
	if slotCounts, ok := slotsTable[level]; ok {
		for i, count := range slotCounts {
			if count > 0 {
				slots = append(slots, models.SpellSlot{
					Level:     i + 1,
					Total:     count,
					Remaining: count,
				})
			}
		}
	}

	return slots
}
