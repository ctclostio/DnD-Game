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

	"github.com/ctclostio/DnD-Game/backend/internal/models"
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
	// Extract and validate parameters
	buildParams, err := cb.extractBuildParameters(params)
	if err != nil {
		return nil, err
	}

	// Create base character
	character := cb.createBaseCharacter(buildParams)

	// Load character data (race, class, background)
	characterData, err := cb.loadCharacterData(buildParams, character)
	if err != nil {
		return nil, err
	}

	// Build the complete character
	cb.assembleCharacter(character, characterData, buildParams)

	return character, nil
}

type buildParameters struct {
	race            string
	customRaceID    string
	customRaceStats map[string]interface{}
	hasCustomRace   bool
	subrace         string
	class           string
	background      string
	name            string
	alignment       string
	abilityScores   map[string]int
}

type characterData struct {
	raceData       *RaceData
	classData      *ClassData
	backgroundData *BackgroundData
}

func (cb *CharacterBuilder) extractBuildParameters(params map[string]interface{}) (*buildParameters, error) {
	bp := &buildParameters{}
	
	bp.race, _ = params["race"].(string)
	bp.customRaceID, _ = params["customRaceId"].(string)
	bp.customRaceStats, bp.hasCustomRace = params["customRaceStats"].(map[string]interface{})
	bp.subrace, _ = params["subrace"].(string)
	bp.class, _ = params["class"].(string)
	bp.background, _ = params["background"].(string)
	bp.name, _ = params["name"].(string)
	bp.alignment, _ = params["alignment"].(string)
	bp.abilityScores, _ = params["abilityScores"].(map[string]int)
	
	// Basic validation
	if bp.name == "" {
		return nil, errors.New("character name is required")
	}
	if bp.class == "" {
		return nil, errors.New("character class is required")
	}
	
	return bp, nil
}

func (cb *CharacterBuilder) createBaseCharacter(params *buildParameters) *models.Character {
	return &models.Character{
		Name:       params.name,
		Race:       params.race,
		Subrace:    params.subrace,
		Class:      params.class,
		Background: params.background,
		Alignment:  params.alignment,
		Level:      1,
	}
}

func (cb *CharacterBuilder) loadCharacterData(params *buildParameters, character *models.Character) (*characterData, error) {
	data := &characterData{}
	var err error

	// Load race data
	data.raceData, err = cb.loadRaceDataWithCustom(params, character)
	if err != nil {
		return nil, err
	}

	// Load class data
	data.classData, err = cb.loadClassData(params.class)
	if err != nil {
		return nil, err
	}

	// Load background data
	data.backgroundData, err = cb.loadBackgroundData(params.background)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (cb *CharacterBuilder) loadRaceDataWithCustom(params *buildParameters, character *models.Character) (*RaceData, error) {
	if params.hasCustomRace && params.customRaceID != "" {
		// Use custom race data
		character.Race = params.customRaceStats["name"].(string)
		character.CustomRaceID = &params.customRaceID
		return cb.convertCustomRaceToRaceData(params.customRaceStats), nil
	}
	
	// Load standard race data
	return cb.loadRaceData(params.race)
}

func (cb *CharacterBuilder) assembleCharacter(character *models.Character, data *characterData, params *buildParameters) {
	// Apply ability scores and racial modifiers
	character.Attributes = cb.calculateFinalAbilityScores(params.abilityScores, data.raceData, params.subrace)

	// Calculate derived stats
	character.ProficiencyBonus = cb.calculateProficiencyBonus(character.Level)
	character.Initiative = cb.calculateModifier(character.Attributes.Dexterity)
	character.Speed = data.raceData.Speed

	// Apply features
	cb.applyClassFeatures(character, data.classData)
	cb.applyRacialFeatures(character, data.raceData, params.subrace)
	cb.applyBackground(character, data.backgroundData)

	// Calculate final stats
	character.SavingThrows = cb.calculateSavingThrows(character, data.classData)
	character.Skills = cb.calculateSkills(character)
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

	// Calculate each saving throw
	saves.Strength = cb.calculateSingleSavingThrow(character.Attributes.Strength, "Strength", classData.SavingThrowProficiencies, profBonus)
	saves.Dexterity = cb.calculateSingleSavingThrow(character.Attributes.Dexterity, "Dexterity", classData.SavingThrowProficiencies, profBonus)
	saves.Constitution = cb.calculateSingleSavingThrow(character.Attributes.Constitution, "Constitution", classData.SavingThrowProficiencies, profBonus)
	saves.Intelligence = cb.calculateSingleSavingThrow(character.Attributes.Intelligence, "Intelligence", classData.SavingThrowProficiencies, profBonus)
	saves.Wisdom = cb.calculateSingleSavingThrow(character.Attributes.Wisdom, "Wisdom", classData.SavingThrowProficiencies, profBonus)
	saves.Charisma = cb.calculateSingleSavingThrow(character.Attributes.Charisma, "Charisma", classData.SavingThrowProficiencies, profBonus)

	return saves
}

func (cb *CharacterBuilder) calculateSingleSavingThrow(abilityScore int, abilityName string, proficiencies []string, profBonus int) models.SavingThrow {
	save := models.SavingThrow{
		Modifier:    cb.calculateModifier(abilityScore),
		Proficiency: cb.contains(proficiencies, abilityName),
	}
	
	if save.Proficiency {
		save.Modifier += profBonus
	}
	
	return save
}

func (cb *CharacterBuilder) calculateSkills(_ *models.Character) []models.Skill {
	// This would be expanded to calculate all skills based on proficiencies
	// For now, return empty slice
	return []models.Skill{}
}

func (cb *CharacterBuilder) applyClassFeatures(character *models.Character, classData *ClassData) {
	// Apply hit dice and calculate HP
	cb.applyHitDiceAndHP(character, classData)

	// Initialize spell slots for spellcasting classes
	if len(classData.Spellcasting) > 0 {
		cb.applySpellcastingFeatures(character, classData)
	}

	// TODO: Apply other class features based on level
}

func (cb *CharacterBuilder) applyHitDiceAndHP(character *models.Character, classData *ClassData) {
	character.HitDice = classData.HitDice
	
	baseHP := cb.getBaseHPFromHitDice(classData.HitDice)
	character.MaxHitPoints = baseHP + cb.calculateModifier(character.Attributes.Constitution)
	character.HitPoints = character.MaxHitPoints
}

func (cb *CharacterBuilder) getBaseHPFromHitDice(hitDice string) int {
	switch hitDice {
	case "1d6":
		return 6
	case "1d8":
		return 8
	case "1d10":
		return 10
	case "1d12":
		return 12
	default:
		return 8 // Default to d8 if unknown
	}
}

func (cb *CharacterBuilder) applySpellcastingFeatures(character *models.Character, classData *ClassData) {
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
	cb.applyCantripsKnown(character, classData)
}

func (cb *CharacterBuilder) applyCantripsKnown(character *models.Character, classData *ClassData) {
	cantripsKnown, ok := classData.Spellcasting["cantripsKnown"].([]interface{})
	if !ok {
		return
	}

	for _, levelData := range cantripsKnown {
		if cb.extractCantripsForLevel(character, levelData) {
			break
		}
	}
}

func (cb *CharacterBuilder) extractCantripsForLevel(character *models.Character, levelData interface{}) bool {
	levelMap, ok := levelData.(map[string]interface{})
	if !ok {
		return false
	}

	level, ok := levelMap["level"].(float64)
	if !ok || int(level) != character.Level {
		return false
	}

	known, ok := levelMap["known"].(float64)
	if ok {
		character.Spells.CantripsKnown = int(known)
		return true
	}

	return false
}

func (cb *CharacterBuilder) applyRacialFeatures(character *models.Character, raceData *RaceData, _ string) {
	// Apply racial traits
	character.Proficiencies.Languages = append(character.Proficiencies.Languages, raceData.Languages...)

	// TODO: Apply other racial features
}

func (cb *CharacterBuilder) applyBackground(_ *models.Character, backgroundData *BackgroundData) {
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

	cb.convertAbilityScoreIncreases(customRaceStats, raceData)
	cb.convertLanguages(customRaceStats, raceData)
	cb.convertTraits(customRaceStats, raceData)
	cb.addDarkvision(customRaceStats, raceData)
	cb.addResistances(customRaceStats, raceData)
	cb.addImmunities(customRaceStats, raceData)
	cb.addSkillProficiencies(customRaceStats, raceData)

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

// Helper methods for convertCustomRaceToRaceData to reduce cyclomatic complexity

func (cb *CharacterBuilder) convertAbilityScoreIncreases(customRaceStats map[string]interface{}, raceData *RaceData) {
	if asi, ok := customRaceStats["abilityScoreIncreases"].(map[string]interface{}); ok {
		for ability, increase := range asi {
			if val, ok := increase.(float64); ok {
				raceData.AbilityIncreases[ability] = int(val)
			}
		}
	}
}

func (cb *CharacterBuilder) convertLanguages(customRaceStats map[string]interface{}, raceData *RaceData) {
	if languages, ok := customRaceStats["languages"].([]interface{}); ok {
		for _, lang := range languages {
			if langStr, ok := lang.(string); ok {
				raceData.Languages = append(raceData.Languages, langStr)
			}
		}
	}
}

func (cb *CharacterBuilder) convertTraits(customRaceStats map[string]interface{}, raceData *RaceData) {
	if traits, ok := customRaceStats["traits"].([]interface{}); ok {
		for _, trait := range traits {
			if traitMap, ok := trait.(map[string]interface{}); ok {
				raceData.Traits = append(raceData.Traits, traitMap)
			}
		}
	}
}

func (cb *CharacterBuilder) addDarkvision(customRaceStats map[string]interface{}, raceData *RaceData) {
	if darkvision, ok := customRaceStats["darkvision"].(float64); ok && darkvision > 0 {
		raceData.Traits = append(raceData.Traits, map[string]interface{}{
			"name":        "Darkvision",
			"description": fmt.Sprintf("You can see in dim light within %d feet as if it were bright light, and in darkness as if it were dim light.", int(darkvision)),
		})
	}
}

// addListTrait is a generic helper to add traits from list data
func (cb *CharacterBuilder) addListTrait(customRaceStats map[string]interface{}, raceData *RaceData, 
	key string, traitName string, descriptionFormat string) {
	if items, ok := customRaceStats[key].([]interface{}); ok && len(items) > 0 {
		itemList := []string{}
		for _, item := range items {
			if itemStr, ok := item.(string); ok {
				itemList = append(itemList, itemStr)
			}
		}
		if len(itemList) > 0 {
			separator := ", "
			if key == "skillProficiencies" {
				separator = " and "
			}
			raceData.Traits = append(raceData.Traits, map[string]interface{}{
				"name":        traitName,
				"description": fmt.Sprintf(descriptionFormat, strings.Join(itemList, separator)),
			})
		}
	}
}

func (cb *CharacterBuilder) addResistances(customRaceStats map[string]interface{}, raceData *RaceData) {
	cb.addListTrait(customRaceStats, raceData, "resistances", "Damage Resistance", 
		"You have resistance to %s damage.")
}

func (cb *CharacterBuilder) addImmunities(customRaceStats map[string]interface{}, raceData *RaceData) {
	cb.addListTrait(customRaceStats, raceData, "immunities", "Damage Immunity", 
		"You are immune to %s damage.")
}

func (cb *CharacterBuilder) addSkillProficiencies(customRaceStats map[string]interface{}, raceData *RaceData) {
	cb.addListTrait(customRaceStats, raceData, "skillProficiencies", "Skill Proficiencies", 
		"You gain proficiency in %s.")
}
