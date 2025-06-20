package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Constants for LLM system prompts
const (
	systemPromptWorldBuilding = "You are a creative D&D world-building assistant."
	systemPromptCultureGen = "You are a creative D&D world-building assistant specializing in unique cultural generation."
	systemPromptLanguageGen = "You are a creative D&D world-building assistant specializing in language generation."
)

// ProceduralCultureService generates unique cultures with AI
type ProceduralCultureService struct {
	worldRepo *database.EmergentWorldRepository
	llm       LLMProvider
}

// NewProceduralCultureService creates a new culture generation service
func NewProceduralCultureService(
	worldRepo *database.EmergentWorldRepository,
	llm LLMProvider,
) *ProceduralCultureService {
	return &ProceduralCultureService{
		worldRepo: worldRepo,
		llm:       llm,
	}
}

// GenerateCulture creates a complete unique culture
func (pcs *ProceduralCultureService) GenerateCulture(ctx context.Context, sessionID string, parameters CultureGenParameters) (*models.ProceduralCulture, error) {
	// Generate base culture name
	cultureName := pcs.generateCultureName(parameters)

	// Generate culture foundation using AI
	foundation, err := pcs.generateCultureFoundation(ctx, cultureName, parameters)
	if err != nil {
		return nil, err
	}

	// Create culture components
	culture := &models.ProceduralCulture{
		ID:                uuid.New().String(),
		Name:              cultureName,
		Language:          pcs.generateLanguage(ctx, cultureName, foundation),
		Customs:           pcs.generateCustoms(ctx, cultureName, foundation),
		ArtStyle:          pcs.generateArtStyle(ctx, cultureName, foundation),
		BeliefSystem:      pcs.generateBeliefSystem(ctx, cultureName, foundation),
		Values:            foundation.Values,
		Taboos:            foundation.Taboos,
		Greetings:         pcs.generateGreetings(ctx, cultureName, foundation),
		Architecture:      pcs.generateArchitecture(ctx, cultureName, foundation),
		Cuisine:           pcs.generateCuisine(ctx, cultureName, foundation),
		MusicStyle:        pcs.generateMusicStyle(ctx, cultureName, foundation),
		ClothingStyle:     pcs.generateClothingStyle(ctx, cultureName, foundation),
		NamingConventions: pcs.generateNamingConventions(ctx, cultureName, foundation),
		SocialStructure:   pcs.generateSocialStructure(ctx, cultureName, foundation),
		Metadata: map[string]interface{}{
			"session_id":      sessionID,
			"environment":     parameters.Environment,
			"origin_story":    foundation.OriginStory,
			"cultural_heroes": foundation.CulturalHeroes,
		},
		CreatedAt: time.Now(),
	}

	// Save culture
	if err := pcs.worldRepo.CreateCulture(culture); err != nil {
		return nil, err
	}

	return culture, nil
}

// generateCultureFoundation creates the core cultural identity
func (pcs *ProceduralCultureService) generateCultureFoundation(ctx context.Context, name string, params CultureGenParameters) (*CultureFoundation, error) {
	prompt := fmt.Sprintf(`Create a unique fantasy culture with these parameters:
Name: %s
Environment: %s
Historical Events: %s
Neighboring Cultures: %v
Special Characteristics: %v

Generate a comprehensive cultural foundation with:
1. Core values (5-7 key values with importance 0.0-1.0)
2. Taboos (3-5 forbidden practices)
3. Origin story (2-3 paragraphs)
4. Cultural heroes (2-3 legendary figures)
5. Worldview (how they see reality, magic, nature)
6. Social priorities (what matters most to them)

Return as JSON with keys: values (map), taboos (array), origin_story, cultural_heroes (array of {name, deeds}), worldview, social_priorities`,
		name, params.Environment, params.HistoricalContext,
		params.NeighboringCultures, params.SpecialTraits)

	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptCultureGen)
	if err != nil {
		return pcs.generateDefaultFoundation(name), nil
	}

	var foundation CultureFoundation
	if err := json.Unmarshal([]byte(response), &foundation); err != nil {
		return pcs.generateDefaultFoundation(name), nil
	}

	return &foundation, nil
}

// generateLanguage creates linguistic characteristics
func (pcs *ProceduralCultureService) generateLanguage(ctx context.Context, cultureName string, foundation *CultureFoundation) models.CultureLanguage {
	// Generate phonemes based on culture
	phonemes := pcs.generatePhonemes(cultureName)

	// Generate common words
	prompt := fmt.Sprintf(`Create a language for the %s culture:
Values: %v
Worldview: %s

Generate:
1. Common words (hello, goodbye, yes, no, friend, enemy, honor, home, magic, god) with translations
2. 3 idioms with literal translation and meaning
3. Grammar rule summary (1 sentence)
4. Honorific system (how they address different social ranks)

Return as JSON with keys: common_words (map), idioms (array of {expression, literal, meaning}), grammar_summary, honorifics (array)`,
		cultureName, foundation.Values, foundation.Worldview)

	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptLanguageGen)
	if err != nil {
		// Return default language on error
		return models.CultureLanguage{
			Name:          cultureName + "ish",
			Phonemes:      phonemes,
			WritingSystem: pcs.generateWritingSystem(),
			CommonWords:   pcs.generateDefaultWords(phonemes),
		}
	}

	var langData map[string]interface{}
	_ = json.Unmarshal([]byte(response), &langData)

	// Build language structure
	language := models.CultureLanguage{
		Name:          cultureName + "ish",
		Phonemes:      phonemes,
		WritingSystem: pcs.generateWritingSystem(),
	}

	// Parse common words
	if words, ok := langData["common_words"].(map[string]interface{}); ok {
		language.CommonWords = make(map[string]string)
		for k, v := range words {
			language.CommonWords[k] = v.(string)
		}
	} else {
		language.CommonWords = pcs.generateDefaultWords(phonemes)
	}

	// Parse idioms
	if idioms, ok := langData["idioms"].([]interface{}); ok {
		for _, idiom := range idioms {
			if idiomMap, ok := idiom.(map[string]interface{}); ok {
				language.Idioms = append(language.Idioms, models.LanguageIdiom{
					Expression: idiomMap["expression"].(string),
					Meaning:    idiomMap["meaning"].(string),
					Context:    "general",
					Formality:  "neutral",
				})
			}
		}
	}

	// Grammar rules
	if grammar, ok := langData["grammar_summary"].(string); ok {
		language.GrammarRules = []string{grammar}
	}

	// Honorifics
	if honorifics, ok := langData["honorifics"].([]interface{}); ok {
		for _, h := range honorifics {
			language.HonorificRules = append(language.HonorificRules, h.(string))
		}
	}

	return language
}

// generateCustoms creates cultural practices
func (pcs *ProceduralCultureService) generateCustoms(ctx context.Context, cultureName string, foundation *CultureFoundation) []models.CultureCustom {
	customTypes := []string{"ceremony", "daily_practice", "seasonal", "lifecycle"}
	customs := []models.CultureCustom{}

	for _, customType := range customTypes {
		prompt := fmt.Sprintf(`Create a %s custom for the %s culture:
Values: %v
Taboos: %v

Generate a unique cultural practice including:
1. Name of the custom
2. Detailed description (2-3 sentences)
3. When/how often it occurs
4. Who participates
5. Cultural significance

Return as JSON with keys: name, description, frequency, participants, significance`,
			customType, cultureName, foundation.Values, foundation.Taboos)

		response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)

		var customData map[string]interface{}
		if err == nil && json.Unmarshal([]byte(response), &customData) == nil {
			custom := models.CultureCustom{
				Name:         customData["name"].(string),
				Type:         customType,
				Description:  customData["description"].(string),
				Frequency:    customData["frequency"].(string),
				Participants: customData["participants"].(string),
				Significance: 0.5 + rand.Float64()*0.5,
				Requirements: make(map[string]interface{}),
			}
			customs = append(customs, custom)
		}
	}

	return customs
}

// generateArtStyle creates artistic preferences
func (pcs *ProceduralCultureService) generateArtStyle(ctx context.Context, cultureName string, foundation *CultureFoundation) models.CultureArtStyle {
	prompt := fmt.Sprintf(`Design the art style for the %s culture:
Environment: %v
Values: %v
Worldview: %s

Create:
1. Primary art mediums (3-4)
2. Common motifs and symbols (4-5)
3. Color palette (4-5 colors with cultural meaning)
4. 2-3 sacred symbols with descriptions
5. Overall style description (1-2 sentences)

Return as JSON with keys: mediums (array), motifs (array), colors (array of {color, meaning}), sacred_symbols (array of {name, description, meaning}), style_description`,
		cultureName, foundation.Environment, foundation.Values, foundation.Worldview)

	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)

	var artData map[string]interface{}
	artStyle := models.CultureArtStyle{
		Materials:  []string{"stone", "wood", "cloth", "metal"},
		Techniques: make(map[string]string),
		Influences: make(map[string]interface{}),
	}

	if err == nil && json.Unmarshal([]byte(response), &artData) == nil {
		// Parse mediums
		if mediums, ok := artData["mediums"].([]interface{}); ok {
			for _, m := range mediums {
				artStyle.PrimaryMediums = append(artStyle.PrimaryMediums, m.(string))
			}
		}

		// Parse motifs
		if motifs, ok := artData["motifs"].([]interface{}); ok {
			for _, m := range motifs {
				artStyle.CommonMotifs = append(artStyle.CommonMotifs, m.(string))
			}
		}

		// Parse colors
		if colors, ok := artData["colors"].([]interface{}); ok {
			for _, c := range colors {
				if colorMap, ok := c.(map[string]interface{}); ok {
					artStyle.ColorPalette = append(artStyle.ColorPalette, colorMap["color"].(string))
				}
			}
		}

		// Parse sacred symbols
		if symbols, ok := artData["sacred_symbols"].([]interface{}); ok {
			for _, s := range symbols {
				if symbolMap, ok := s.(map[string]interface{}); ok {
					artStyle.SacredSymbols = append(artStyle.SacredSymbols, models.ArtSymbol{
						Name:        symbolMap["name"].(string),
						Description: symbolMap["description"].(string),
						Meaning:     symbolMap["meaning"].(string),
						Usage:       "religious and ceremonial",
					})
				}
			}
		}

		// Style description
		if desc, ok := artData["style_description"].(string); ok {
			artStyle.StyleDescription = desc
		}
	}

	return artStyle
}

// generateBeliefSystem creates religious/philosophical beliefs
func (pcs *ProceduralCultureService) generateBeliefSystem(ctx context.Context, cultureName string, foundation *CultureFoundation) models.CultureBeliefSystem {
	beliefTypes := []string{"polytheistic", "monotheistic", "animistic", "philosophical"}
	selectedType := beliefTypes[rand.Intn(len(beliefTypes))]

	prompt := fmt.Sprintf(`Create a %s belief system for the %s culture:
Origin Story: %s
Cultural Heroes: %v
Values: %v

Generate:
1. Name of the belief system
2. 2-3 deities/spirits/principles with domains
3. Core beliefs (3-4 fundamental tenets)
4. Major religious practices (2-3)
5. View of afterlife
6. Creation myth summary (2-3 sentences)

Return as JSON with keys: name, deities (array of {name, title, domains, personality}), core_beliefs (array), practices (array of {name, description, frequency}), afterlife, creation_myth`,
		selectedType, cultureName, foundation.OriginStory,
		foundation.CulturalHeroes, foundation.Values)

	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)

	var beliefData map[string]interface{}
	beliefSystem := models.CultureBeliefSystem{
		Type:        selectedType,
		MoralCode:   make(map[string]string),
		SacredTexts: []string{cultureName + " Codex"},
		ClergyRanks: pcs.generateClergyRanks(selectedType),
		Miracles:    make(map[string]interface{}),
	}

	if err == nil && json.Unmarshal([]byte(response), &beliefData) == nil {
		// Parse name
		if name, ok := beliefData["name"].(string); ok {
			beliefSystem.Name = name
		}

		// Parse deities
		if deities, ok := beliefData["deities"].([]interface{}); ok {
			for _, d := range deities {
				if deityMap, ok := d.(map[string]interface{}); ok {
					domains := []string{}
					if domainList, ok := deityMap["domains"].([]interface{}); ok {
						for _, domain := range domainList {
							domains = append(domains, domain.(string))
						}
					}

					beliefSystem.Deities = append(beliefSystem.Deities, models.CultureDeity{
						Name:        deityMap["name"].(string),
						Title:       deityMap["title"].(string),
						Domain:      domains,
						Personality: deityMap["personality"].(string),
						Symbol:      pcs.generateDeitySymbol(),
						Alignment:   pcs.generateAlignment(),
					})
				}
			}
		}

		// Parse beliefs and other elements
		if beliefs, ok := beliefData["core_beliefs"].([]interface{}); ok {
			for _, b := range beliefs {
				beliefSystem.CoreBeliefs = append(beliefSystem.CoreBeliefs, b.(string))
			}
		}

		if afterlife, ok := beliefData["afterlife"].(string); ok {
			beliefSystem.Afterlife = afterlife
		}

		if creation, ok := beliefData["creation_myth"].(string); ok {
			beliefSystem.CreationMyth = creation
		}

		// Parse practices
		if practices, ok := beliefData["practices"].([]interface{}); ok {
			for _, p := range practices {
				if practiceMap, ok := p.(map[string]interface{}); ok {
					beliefSystem.Practices = append(beliefSystem.Practices, models.ReligiousPractice{
						Name:        practiceMap["name"].(string),
						Type:        "ritual",
						Frequency:   practiceMap["frequency"].(string),
						Description: practiceMap["description"].(string),
						Materials:   []string{},
						Duration:    "varies",
						Effects:     make(map[string]interface{}),
					})
				}
			}
		}
	}

	// Generate holy days
	beliefSystem.HolyDays = pcs.generateHolyDays(cultureName, beliefSystem.Deities)

	return beliefSystem
}

// generateGreetings creates cultural greetings
func (pcs *ProceduralCultureService) generateGreetings(ctx context.Context, cultureName string, foundation *CultureFoundation) map[string]string {
	contexts := []string{"formal", "informal", "morning", "evening", "farewell", "blessing"}
	greetings := make(map[string]string)

	prompt := fmt.Sprintf(`Create greetings for the %s culture:
Values: %v
Social Priorities: %v

Generate appropriate greetings for: %v
Include the cultural significance behind each greeting.

Return as JSON map of context to greeting phrase.`,
		cultureName, foundation.Values, foundation.SocialPriorities, contexts)

	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)

	if err == nil {
		_ = json.Unmarshal([]byte(response), &greetings)
	}

	// Ensure all contexts have greetings
	for _, context := range contexts {
		if _, ok := greetings[context]; !ok {
			greetings[context] = pcs.generateDefaultGreeting(context, cultureName)
		}
	}

	return greetings
}

// generateArchitecture creates building styles
func (pcs *ProceduralCultureService) generateArchitecture(ctx context.Context, cultureName string, foundation *CultureFoundation) models.ArchitectureStyle {
	// Determine materials based on environment
	materials := pcs.getMaterialsForEnvironment(foundation.Environment)

	prompt := fmt.Sprintf(`Design architecture for the %s culture:
Environment: %s
Available Materials: %v
Values: %v

Create:
1. Architecture style name
2. Common building features (4-5)
3. Defensive elements (2-3)
4. Decorative elements (3-4)
5. Typical settlement layout description

Return as JSON with keys: style_name, features (array), defenses (array), decorations (array), layout_description`,
		cultureName, foundation.Environment, materials, foundation.Values)

	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)

	architecture := models.ArchitectureStyle{
		Materials:     materials,
		BuildingTypes: make(map[string]models.BuildingStyle),
	}

	var archData map[string]interface{}
	if err == nil && json.Unmarshal([]byte(response), &archData) == nil {
		if name, ok := archData["style_name"].(string); ok {
			architecture.Name = name
		}

		if features, ok := archData["features"].([]interface{}); ok {
			for _, f := range features {
				architecture.CommonFeatures = append(architecture.CommonFeatures, f.(string))
			}
		}

		if defenses, ok := archData["defenses"].([]interface{}); ok {
			for _, d := range defenses {
				architecture.DefensiveElements = append(architecture.DefensiveElements, d.(string))
			}
		}

		if decorations, ok := archData["decorations"].([]interface{}); ok {
			for _, d := range decorations {
				architecture.Decorations = append(architecture.Decorations, d.(string))
			}
		}

		if layout, ok := archData["layout_description"].(string); ok {
			architecture.TypicalLayout = layout
		}
	}

	// Generate building types
	architecture.BuildingTypes = pcs.generateBuildingTypes(cultureName, architecture)

	return architecture
}

// generateCuisine creates food culture
func (pcs *ProceduralCultureService) generateCuisine(ctx context.Context, cultureName string, foundation *CultureFoundation) []models.CuisineElement {
	cuisineTypes := []string{"staple", "delicacy", "ceremonial", "everyday"}
	cuisine := []models.CuisineElement{}

	for _, cuisineType := range cuisineTypes {
		ingredients := pcs.getIngredientsForEnvironment(foundation.Environment)

		prompt := fmt.Sprintf(`Create a %s dish for the %s culture:
Available ingredients: %v
Taboos: %v

Generate:
1. Dish name
2. Main ingredients (3-4)
3. Preparation method (1-2 sentences)
4. When it's eaten
5. Cultural significance

Return as JSON with keys: name, ingredients (array), preparation, occasion, significance`,
			cuisineType, cultureName, ingredients, foundation.Taboos)

		response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)

		var dishData map[string]interface{}
		if err == nil && json.Unmarshal([]byte(response), &dishData) == nil {
			dish := models.CuisineElement{
				Type: cuisineType,
			}

			if name, ok := dishData["name"].(string); ok {
				dish.Name = name
			}

			if ingredients, ok := dishData["ingredients"].([]interface{}); ok {
				for _, i := range ingredients {
					dish.Ingredients = append(dish.Ingredients, i.(string))
				}
			}

			if prep, ok := dishData["preparation"].(string); ok {
				dish.Preparation = prep
			}

			if occasion, ok := dishData["occasion"].(string); ok {
				dish.Occasion = occasion
			}

			if significance, ok := dishData["significance"].(string); ok {
				dish.Significance = significance
			}

			cuisine = append(cuisine, dish)
		}
	}

	return cuisine
}

// generateMusicStyle creates musical traditions
func (pcs *ProceduralCultureService) generateMusicStyle(ctx context.Context, cultureName string, foundation *CultureFoundation) models.MusicStyle {
	prompt := fmt.Sprintf(`Create musical traditions for the %s culture:
Values: %v
Cultural Heroes: %v

Generate:
1. Music style name
2. Traditional instruments (3-4)
3. Common rhythmic patterns (2-3)
4. Musical occasions (3-4)
5. Common themes in songs (3-4)
6. Traditional dances (2-3)

Return as JSON with keys: style_name, instruments (array), rhythms (array), occasions (array), themes (array), dances (array)`,
		cultureName, foundation.Values, foundation.CulturalHeroes)

	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)

	musicStyle := models.MusicStyle{
		Scales: pcs.generateMusicalScales(),
	}

	var musicData map[string]interface{}
	if err == nil && json.Unmarshal([]byte(response), &musicData) == nil {
		if name, ok := musicData["style_name"].(string); ok {
			musicStyle.Name = name
		}

		// Parse arrays
		pcs.parseStringArray(musicData, "instruments", &musicStyle.Instruments)
		pcs.parseStringArray(musicData, "rhythms", &musicStyle.Rhythms)
		pcs.parseStringArray(musicData, "occasions", &musicStyle.Occasions)
		pcs.parseStringArray(musicData, "themes", &musicStyle.Themes)
		pcs.parseStringArray(musicData, "dances", &musicStyle.DanceStyles)
	}

	return musicStyle
}

// generateClothingStyle creates fashion traditions
func (pcs *ProceduralCultureService) generateClothingStyle(ctx context.Context, cultureName string, foundation *CultureFoundation) models.ClothingStyle {
	clothingStyle := models.ClothingStyle{
		EverydayWear:   make(map[string]models.ClothingItem),
		FormalWear:     make(map[string]models.ClothingItem),
		CeremonialWear: make(map[string]models.ClothingItem),
		StatusMarkers:  make(map[string]string),
	}

	// Generate materials and colors based on environment
	clothingStyle.Materials = pcs.getClothingMaterials(foundation.Environment)
	clothingStyle.Colors = pcs.generateColorPalette(cultureName)

	// Generate clothing items for each type and gender
	pcs.generateClothingItems(ctx, cultureName, foundation, &clothingStyle)

	// Generate jewelry
	clothingStyle.Jewelry = pcs.generateJewelry(cultureName, foundation)

	return clothingStyle
}

// generateClothingItems creates clothing for different types and genders
func (pcs *ProceduralCultureService) generateClothingItems(ctx context.Context, cultureName string, foundation *CultureFoundation, clothingStyle *models.ClothingStyle) {
	clothingTypes := []string{"everyday", "formal", "ceremonial"}
	genderRoles := []string{"all", "masculine", "feminine"}

	for _, clothingType := range clothingTypes {
		for _, gender := range genderRoles {
			item := pcs.generateSingleClothingItem(ctx, cultureName, foundation, clothingStyle, clothingType, gender)
			if item.Name != "" {
				pcs.assignClothingItem(clothingStyle, clothingType, gender, item)
			}
		}
	}
}

// generateSingleClothingItem creates a single clothing item
func (pcs *ProceduralCultureService) generateSingleClothingItem(ctx context.Context, cultureName string, foundation *CultureFoundation, clothingStyle *models.ClothingStyle, clothingType, gender string) models.ClothingItem {
	prompt := pcs.buildClothingPrompt(cultureName, foundation, clothingStyle, clothingType, gender)
	
	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)
	if err != nil {
		return models.ClothingItem{}
	}

	var itemData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &itemData); err != nil {
		return models.ClothingItem{}
	}

	return pcs.parseClothingItem(itemData, gender, clothingStyle)
}

// buildClothingPrompt creates the prompt for clothing generation
func (pcs *ProceduralCultureService) buildClothingPrompt(cultureName string, foundation *CultureFoundation, clothingStyle *models.ClothingStyle, clothingType, gender string) string {
	return fmt.Sprintf(`Design %s %s clothing for the %s culture:
Environment: %s
Materials: %v
Values: %v

Create a distinctive garment with name, description, and decorative elements.

Return as JSON with keys: name, description, decorations (array)`,
		gender, clothingType, cultureName, foundation.Environment,
		clothingStyle.Materials, foundation.Values)
}

// parseClothingItem extracts clothing data from LLM response
func (pcs *ProceduralCultureService) parseClothingItem(itemData map[string]interface{}, gender string, clothingStyle *models.ClothingStyle) models.ClothingItem {
	item := models.ClothingItem{
		WornBy:    gender,
		Materials: clothingStyle.Materials[:2],
		Colors:    clothingStyle.Colors[:2],
	}

	if name, ok := itemData["name"].(string); ok {
		item.Name = name
	}

	if desc, ok := itemData["description"].(string); ok {
		item.Description = desc
	}

	if decorations, ok := itemData["decorations"].([]interface{}); ok {
		for _, d := range decorations {
			if decoration, ok := d.(string); ok {
				item.Decorations = append(item.Decorations, decoration)
			}
		}
	}

	return item
}

// assignClothingItem places the item in the appropriate clothing category
func (pcs *ProceduralCultureService) assignClothingItem(clothingStyle *models.ClothingStyle, clothingType, gender string, item models.ClothingItem) {
	key := fmt.Sprintf("%s_%s", gender, clothingType)
	switch clothingType {
	case "everyday":
		clothingStyle.EverydayWear[key] = item
	case "formal":
		clothingStyle.FormalWear[key] = item
	case "ceremonial":
		clothingStyle.CeremonialWear[key] = item
	}
}

// generateNamingConventions creates naming traditions
func (pcs *ProceduralCultureService) generateNamingConventions(ctx context.Context, cultureName string, foundation *CultureFoundation) models.NamingConventions {
	prompt := fmt.Sprintf(`Create naming conventions for the %s culture:
Language patterns: %v
Values: %v
Social structure: %v

Generate:
1. Given name patterns (2-3 examples with meaning)
2. Family name patterns (2-3 examples)  
3. Title formats (2-3 for different ranks)
4. Nickname rules
5. Taboo naming practices
6. Naming ceremony description

Return as JSON with keys: given_patterns (array), family_patterns (array), titles (array), nickname_rules, taboos (array), ceremony`,
		cultureName, pcs.getLanguagePatterns(cultureName),
		foundation.Values, foundation.SocialPriorities)

	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)

	namingConventions := models.NamingConventions{
		NameMeanings: make(map[string]string),
	}

	var namingData map[string]interface{}
	if err == nil && json.Unmarshal([]byte(response), &namingData) == nil {
		pcs.parseStringArray(namingData, "given_patterns", &namingConventions.GivenNamePatterns)
		pcs.parseStringArray(namingData, "family_patterns", &namingConventions.FamilyNamePatterns)
		pcs.parseStringArray(namingData, "titles", &namingConventions.TitleFormats)
		pcs.parseStringArray(namingData, "taboos", &namingConventions.TabooNames)

		if rules, ok := namingData["nickname_rules"].(string); ok {
			namingConventions.NicknameRules = []string{rules}
		}

		if ceremony, ok := namingData["ceremony"].(string); ok {
			namingConventions.NamingCeremonies = []string{ceremony}
		}
	}

	return namingConventions
}

// generateSocialStructure creates societal organization
func (pcs *ProceduralCultureService) generateSocialStructure(ctx context.Context, cultureName string, foundation *CultureFoundation) models.SocialStructure {
	structureTypes := []string{"caste", "class", "egalitarian", "meritocratic", "theocratic"}
	selectedType := structureTypes[rand.Intn(len(structureTypes))]

	prompt := fmt.Sprintf(`Design a %s social structure for the %s culture:
Values: %v
Social Priorities: %v

Generate:
1. 3-5 social classes/groups with names, privileges, and restrictions
2. Social mobility description
3. Leadership structure
4. Family unit type
5. How outsiders are treated

Return as JSON with keys: classes (array of {name, rank, privileges, restrictions, occupations}), mobility, leadership, family_unit, outsider_treatment`,
		selectedType, cultureName, foundation.Values, foundation.SocialPriorities)

	response, err := pcs.llm.GenerateContent(ctx, prompt, systemPromptWorldBuilding)

	socialStructure := models.SocialStructure{
		Type:        selectedType,
		GenderRoles: make(map[string]string),
		AgeRoles:    make(map[string]string),
	}

	if err != nil {
		return socialStructure
	}

	var structureData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &structureData); err != nil {
		return socialStructure
	}

	// Parse classes
	if classes, ok := structureData["classes"].([]interface{}); ok {
		for i, c := range classes {
			classMap, ok := c.(map[string]interface{})
			if !ok {
				continue
			}

			class := models.SocialClass{
				Rank: i + 1,
			}

			if name, ok := classMap["name"].(string); ok {
				class.Name = name
			}

			pcs.parseStringArray(classMap, "privileges", &class.Privileges)
			pcs.parseStringArray(classMap, "restrictions", &class.Restrictions)
			pcs.parseStringArray(classMap, "occupations", &class.Occupations)

			// Add visual markers
			class.Markers = []string{
				fmt.Sprintf("%s clothing colors", class.Name),
				fmt.Sprintf("%s symbols", class.Name),
			}

			socialStructure.Classes = append(socialStructure.Classes, class)
		}
	}

	if mobility, ok := structureData["mobility"].(string); ok {
		socialStructure.Mobility = mobility
	}

	if leadership, ok := structureData["leadership"].(string); ok {
		socialStructure.Leadership = leadership
	}

	if family, ok := structureData["family_unit"].(string); ok {
		socialStructure.FamilyUnit = family
	}

	if outsiders, ok := structureData["outsider_treatment"].(string); ok {
		socialStructure.Outsiders = outsiders
	}

	// Generate gender and age roles
	socialStructure.GenderRoles = pcs.generateGenderRoles(cultureName, foundation)
	socialStructure.AgeRoles = pcs.generateAgeRoles(cultureName, foundation)

	return socialStructure
}

// RespondToPlayerAction updates culture based on player interactions
func (pcs *ProceduralCultureService) RespondToPlayerAction(ctx context.Context, cultureID string, action PlayerCulturalAction) error {
	culture, err := pcs.worldRepo.GetCulture(cultureID)
	if err != nil {
		return err
	}

	// Determine cultural response based on action and values
	response := pcs.evaluateCulturalResponse(culture, action)

	// Update cultural aspects if significant impact
	if response.Impact > 0.3 {
		switch response.AffectedAspect {
		case constants.CultureValues:
			pcs.adjustCulturalValues(culture, action, response)
		case constants.AspectCustoms:
			pcs.modifyCustoms(culture, action, response)
		case constants.AspectSocialStructure:
			pcs.adjustSocialStructure(culture, action, response)
		}
	}

	// Generate cultural event if major impact
	if response.Impact > 0.5 {
		event := pcs.generateCulturalResponseEvent(ctx, culture, action, response)
		_ = event // placeholder for future event handling
	}

	return pcs.worldRepo.UpdateCulture(culture)
}

// Helper functions and types

type CultureGenParameters struct {
	Environment         string
	HistoricalContext   string
	NeighboringCultures []string
	SpecialTraits       []string
}

type CultureFoundation struct {
	Values           map[string]float64 `json:"values"`
	Taboos           []string           `json:"taboos"`
	OriginStory      string             `json:"origin_story"`
	CulturalHeroes   []CulturalHero     `json:"cultural_heroes"`
	Worldview        string             `json:"worldview"`
	SocialPriorities []string           `json:"social_priorities"`
	Environment      string
}

type CulturalHero struct {
	Name  string `json:"name"`
	Deeds string `json:"deeds"`
}

type PlayerCulturalAction struct {
	Type        string // trade, diplomacy, conflict, influence
	Target      string // specific custom, belief, etc
	Approach    string // respectful, aggressive, subversive
	Magnitude   float64
	Description string
}

type CulturalResponse struct {
	Acceptance     float64
	Resistance     float64
	Impact         float64
	AffectedAspect string
	Description    string
}

// Implementation of helper functions

func (pcs *ProceduralCultureService) generateCultureName(_ CultureGenParameters) string {
	prefixes := []string{"Zar", "Keth", "Mor", "Val", "Syl", "Dra", "Ith", "Nar", "Bel", "Tor"}
	suffixes := []string{"ani", "ari", "eshi", "ovan", "ukai", "enti", "ashi", "orim", "ethi", "alor"}

	return prefixes[rand.Intn(len(prefixes))] + suffixes[rand.Intn(len(suffixes))]
}

func (pcs *ProceduralCultureService) generateDefaultFoundation(name string) *CultureFoundation {
	return &CultureFoundation{
		Values: map[string]float64{
			"honor":     0.5 + rand.Float64()*0.5,
			"tradition": 0.5 + rand.Float64()*0.5,
			"strength":  0.5 + rand.Float64()*0.5,
			"wisdom":    0.5 + rand.Float64()*0.5,
			"unity":     0.5 + rand.Float64()*0.5,
		},
		Taboos:           []string{"breaking oaths", "disrespecting elders", "wasting resources"},
		OriginStory:      fmt.Sprintf("The %s people emerged from ancient times...", name),
		CulturalHeroes:   []CulturalHero{{Name: "The First " + name, Deeds: "Founded the civilization"}},
		Worldview:        "The world is a place of balance between order and chaos",
		SocialPriorities: []string{"family", "community", "tradition"},
	}
}

func (pcs *ProceduralCultureService) generatePhonemes(cultureName string) []string {
	consonants := []string{"p", "t", "k", "b", "d", "g", "m", "n", "l", "r", "s", "sh", "z", "zh", "h", "w", "y"}
	vowels := []string{"a", "e", "i", "o", "u", "ai", "ei", "ou"}

	// Select subset based on culture name hash
	hash := 0
	for _, r := range cultureName {
		hash += int(r)
	}

	selectedConsonants := []string{}
	selectedVowels := []string{}

	for i := 0; i < 8+hash%4; i++ {
		selectedConsonants = append(selectedConsonants, consonants[(hash+i)%len(consonants)])
	}

	for i := 0; i < 3+hash%2; i++ {
		selectedVowels = append(selectedVowels, vowels[(hash+i)%len(vowels)])
	}

	return append(selectedConsonants, selectedVowels...)
}

func (pcs *ProceduralCultureService) generateWritingSystem() string {
	systems := []string{"alphabetic", "syllabic", "logographic", "mixed", "runic", "hieroglyphic"}
	return systems[rand.Intn(len(systems))]
}

func (pcs *ProceduralCultureService) generateDefaultWords(phonemes []string) map[string]string {
	words := make(map[string]string)
	basicWords := []string{"hello", "goodbye", "yes", "no", "friend", "enemy", "honor", "home"}

	for _, word := range basicWords {
		// Generate pseudo-word from phonemes
		length := 2 + rand.Intn(3)
		generated := ""
		for i := 0; i < length; i++ {
			generated += phonemes[rand.Intn(len(phonemes))]
		}
		words[word] = generated
	}

	return words
}

func (pcs *ProceduralCultureService) getMaterialsForEnvironment(environment string) []string {
	materialMap := map[string][]string{
		"forest":   {"wood", "thatch", "bark", "rope"},
		"mountain": {"stone", "slate", "iron", "copper"},
		"desert":   {"sandstone", "adobe", "canvas", "glass"},
		"coastal":  {"driftwood", "coral", "shells", "rope"},
		"plains":   {"sod", "timber", "hide", "bone"},
		"swamp":    {"reeds", "mud brick", "wood", "vines"},
		"tundra":   {"ice", "stone", "hide", "bone"},
		"volcanic": {"obsidian", "basalt", "pumice", "metal"},
	}

	if materials, ok := materialMap[strings.ToLower(environment)]; ok {
		return materials
	}
	return []string{"wood", "stone", "clay", "thatch"}
}

func (pcs *ProceduralCultureService) generateClergyRanks(beliefType string) []string {
	rankMap := map[string][]string{
		"polytheistic":  {"Initiate", "Acolyte", "Priest/Priestess", "High Priest/Priestess", "Oracle"},
		"monotheistic":  {"Novice", "Brother/Sister", "Father/Mother", "Bishop", "Archbishop"},
		"animistic":     {"Seeker", "Spirit Walker", "Shaman", "Elder Shaman", "Spirit Master"},
		"philosophical": {"Student", "Scholar", "Master", "Sage", "Enlightened One"},
	}

	if ranks, ok := rankMap[beliefType]; ok {
		return ranks
	}
	return []string{"Initiate", "Adept", "Master", "Elder", "Supreme"}
}

func (pcs *ProceduralCultureService) generateDeitySymbol() string {
	symbols := []string{"sun", "moon", "star", "tree", "mountain", "wave", "flame", "eye", "hand", "sword", "shield", "crown"}
	return symbols[rand.Intn(len(symbols))]
}

func (pcs *ProceduralCultureService) generateAlignment() string {
	alignments := []string{"Lawful Good", "Neutral Good", "Chaotic Good", "Lawful Neutral", "True Neutral", "Chaotic Neutral"}
	return alignments[rand.Intn(len(alignments))]
}

func (pcs *ProceduralCultureService) generateHolyDays(_ string, deities []models.CultureDeity) []models.HolyDay {
	holyDays := []models.HolyDay{}
	seasons := []string{"Spring Equinox", "Summer Solstice", "Autumn Equinox", "Winter Solstice"}

	// Seasonal celebrations
	for i, season := range seasons {
		holyDays = append(holyDays, models.HolyDay{
			Name:         fmt.Sprintf("Festival of %s", season),
			Date:         season,
			Duration:     fmt.Sprintf("%d days", 1+rand.Intn(3)),
			Celebration:  "Feasting, dancing, and offerings",
			Restrictions: []string{},
			Traditions:   []string{"Bonfires", "Ritual cleansing", "Community feast"},
		})

		if i < len(deities) {
			holyDays[i].Name = fmt.Sprintf("Festival of %s", deities[i].Name)
		}
	}

	return holyDays
}

func (pcs *ProceduralCultureService) generateDefaultGreeting(context, _ string) string {
	greetingMap := map[string]string{
		"formal":   "May the ancestors guide you",
		"informal": "Well met, friend",
		"morning":  "The sun greets you",
		"evening":  "The stars watch over you",
		"farewell": "Until paths cross again",
		"blessing": "Walk in harmony",
	}

	if greeting, ok := greetingMap[context]; ok {
		return greeting
	}
	return "Greetings"
}

func (pcs *ProceduralCultureService) generateBuildingTypes(_ string, architecture models.ArchitectureStyle) map[string]models.BuildingStyle {
	buildingTypes := make(map[string]models.BuildingStyle)

	types := []string{"dwelling", "temple", "market", "fortress", "hall"}

	for _, buildingType := range types {
		building := models.BuildingStyle{
			Purpose:     buildingType,
			Materials:   architecture.Materials[:2],
			Features:    []string{},
			Inhabitants: pcs.getBuildingInhabitants(buildingType),
		}

		switch buildingType {
		case "dwelling":
			building.Size = constants.SizeSmall
			building.Features = []string{"hearth", "sleeping area", "storage"}
		case "temple":
			building.Size = constants.SizeLarge
			building.Features = append(architecture.CommonFeatures[:2], "altar", "sacred space")
		case "market":
			building.Size = constants.SizeMedium
			building.Features = []string{"stalls", "covered walkways", "central plaza"}
		case "fortress":
			building.Size = "massive"
			building.Features = architecture.DefensiveElements
		case "hall":
			building.Size = constants.SizeLarge
			building.Features = []string{"great hearth", "high ceiling", "gathering space"}
		}

		buildingTypes[buildingType] = building
	}

	return buildingTypes
}

func (pcs *ProceduralCultureService) getBuildingInhabitants(buildingType string) string {
	inhabitants := map[string]string{
		"dwelling": "families",
		"temple":   "clergy",
		"market":   "merchants",
		"fortress": "soldiers",
		"hall":     "nobles",
	}

	if inhab, ok := inhabitants[buildingType]; ok {
		return inhab
	}
	return "various"
}

func (pcs *ProceduralCultureService) getIngredientsForEnvironment(environment string) []string {
	ingredientMap := map[string][]string{
		"forest":   {"mushrooms", "berries", "venison", "herbs", "nuts", "honey"},
		"mountain": {"goat", "root vegetables", "hardy grains", "preserved meats", "cheese"},
		"desert":   {"dates", "figs", "goat", "flatbread", "spices", "cactus"},
		"coastal":  {"fish", "shellfish", "seaweed", "salt", "citrus", "coconut"},
		"plains":   {"grains", "beef", "vegetables", "dairy", "herbs", "wild game"},
		"swamp":    {"rice", "fish", "waterfowl", "tubers", "greens", "spices"},
	}

	if ingredients, ok := ingredientMap[strings.ToLower(environment)]; ok {
		return ingredients
	}
	return []string{"grain", "meat", "vegetables", "herbs", "fruit"}
}

func (pcs *ProceduralCultureService) generateMusicalScales() []string {
	scales := []string{"pentatonic", "heptatonic", "chromatic", "modal", "microtonal"}
	selected := []string{}

	numScales := 1 + rand.Intn(3)
	for i := 0; i < numScales; i++ {
		selected = append(selected, scales[rand.Intn(len(scales))])
	}

	return selected
}

func (pcs *ProceduralCultureService) parseStringArray(data map[string]interface{}, key string, target *[]string) {
	if arr, ok := data[key].([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				*target = append(*target, str)
			}
		}
	}
}

func (pcs *ProceduralCultureService) getClothingMaterials(environment string) []string {
	materialMap := map[string][]string{
		"forest":   {"leather", "wool", "linen", "fur"},
		"mountain": {"wool", "leather", "fur", "felt"},
		"desert":   {"linen", "cotton", "silk", "leather"},
		"coastal":  {"cotton", "linen", "sailcloth", "shells"},
		"plains":   {"leather", "wool", "cotton", "hide"},
		"swamp":    {"reeds", "cotton", "leather", "fiber"},
	}

	if materials, ok := materialMap[strings.ToLower(environment)]; ok {
		return materials
	}
	return []string{"cloth", "leather", "wool", "linen"}
}

func (pcs *ProceduralCultureService) generateColorPalette(_ string) []string {
	allColors := []string{
		"crimson", "azure", "emerald", "gold", "silver", "obsidian",
		"ivory", "amber", "violet", "turquoise", "ochre", "indigo",
	}

	// Select 4-6 colors
	numColors := 4 + rand.Intn(3)
	selected := []string{}

	for i := 0; i < numColors; i++ {
		color := allColors[rand.Intn(len(allColors))]
		if !containsInCulture(selected, color) {
			selected = append(selected, color)
		}
	}

	return selected
}

func (pcs *ProceduralCultureService) generateJewelry(_ string, _ *CultureFoundation) []string {
	jewelry := []string{}
	types := []string{"rings", "necklaces", "bracelets", "earrings", "brooches", "circlets", "anklets"}

	// Select based on values
	numTypes := 3 + rand.Intn(3)
	for i := 0; i < numTypes; i++ {
		jewelry = append(jewelry, types[rand.Intn(len(types))])
	}

	return jewelry
}

func (pcs *ProceduralCultureService) getLanguagePatterns(cultureName string) []string {
	// Simple pattern generation based on culture name
	patterns := []string{}

	if strings.Contains(cultureName, "ar") {
		patterns = append(patterns, "flowing sounds")
	}
	if strings.Contains(cultureName, "k") || strings.Contains(cultureName, "th") {
		patterns = append(patterns, "harsh consonants")
	}

	patterns = append(patterns, "compound words", "suffix-based")

	return patterns
}

func (pcs *ProceduralCultureService) generateGenderRoles(_ string, foundation *CultureFoundation) map[string]string {
	// Generate based on cultural values
	roles := make(map[string]string)

	if foundation.Values["tradition"] > 0.7 {
		roles["masculine"] = "warriors and leaders"
		roles["feminine"] = "healers and keepers of lore"
		roles["other"] = "spiritual guides"
	} else if foundation.Values["equality"] > 0.7 {
		roles["all"] = "any role based on aptitude"
	} else {
		roles["various"] = "roles determined by clan and calling"
	}

	return roles
}

func (pcs *ProceduralCultureService) generateAgeRoles(_ string, _ *CultureFoundation) map[string]string {
	return map[string]string{
		"children":     "learn and play",
		"youth":        "prove themselves",
		"adults":       "provide and protect",
		"elders":       "guide and teach",
		"ancient_ones": "revered advisors",
	}
}

func (pcs *ProceduralCultureService) evaluateCulturalResponse(_ *models.ProceduralCulture, action PlayerCulturalAction) CulturalResponse {
	response := CulturalResponse{
		AffectedAspect: constants.CultureValues,
	}

	// Calculate acceptance based on cultural values and action approach
	switch action.Approach {
	case "respectful":
		response.Acceptance = 0.6 + rand.Float64()*0.3
		response.Resistance = 0.1 + rand.Float64()*0.2
	case "aggressive":
		response.Acceptance = 0.1 + rand.Float64()*0.2
		response.Resistance = 0.6 + rand.Float64()*0.3
	default:
		response.Acceptance = 0.3 + rand.Float64()*0.3
		response.Resistance = 0.3 + rand.Float64()*0.3
	}

	response.Impact = (response.Acceptance - response.Resistance) * action.Magnitude

	// Determine affected aspect
	switch action.Type {
	case constants.ActionTrade:
		response.AffectedAspect = constants.AspectCustoms
	case constants.ActionDiplomacy:
		response.AffectedAspect = constants.CultureValues
	case constants.EventConflict:
		response.AffectedAspect = constants.AspectSocialStructure
	case constants.CultureInfluence:
		response.AffectedAspect = action.Target
	}

	return response
}

func (pcs *ProceduralCultureService) adjustCulturalValues(culture *models.ProceduralCulture, action PlayerCulturalAction, response CulturalResponse) {
	// Adjust values based on player influence
	impactMagnitude := response.Impact * 0.1 // Max 10% change

	// Example: trade might increase value of wealth
	if action.Type == constants.ActionTrade {
		culture.Values["wealth"] = math.Min(1.0, culture.Values["wealth"]+impactMagnitude)
		culture.Values["isolation"] = math.Max(0.0, culture.Values["isolation"]-impactMagnitude)
	}
}

func (pcs *ProceduralCultureService) modifyCustoms(culture *models.ProceduralCulture, action PlayerCulturalAction, response CulturalResponse) {
	// Potentially add new customs or modify existing ones
	if response.Impact > 0.7 && action.Type == constants.CultureInfluence {
		// Create new custom influenced by players
		newCustom := models.CultureCustom{
			Name:         fmt.Sprintf("Festival of %s", action.Target),
			Type:         "ceremonial",
			Description:  "A new tradition inspired by outsider influence",
			Frequency:    "annual",
			Participants: "all",
			Significance: response.Impact,
			Requirements: map[string]interface{}{
				"origin": "player_influenced",
			},
		}
		culture.Customs = append(culture.Customs, newCustom)
	}
}

func (pcs *ProceduralCultureService) adjustSocialStructure(culture *models.ProceduralCulture, action PlayerCulturalAction, response CulturalResponse) {
	// Modify social mobility or class structure based on actions
	if action.Type == constants.EventConflict && response.Impact < -0.5 {
		culture.SocialStructure.Mobility = "more rigid due to external threats"
	} else if action.Type == constants.ActionDiplomacy && response.Impact > 0.5 {
		culture.SocialStructure.Outsiders = "viewed with cautious respect"
	}
}

func (pcs *ProceduralCultureService) generateCulturalResponseEvent(_ context.Context, culture *models.ProceduralCulture, action PlayerCulturalAction, response CulturalResponse) *models.EmergentWorldEvent {
	eventTypes := map[string]string{
		constants.ActionTrade:      "commercial_shift",
		constants.ActionDiplomacy:  "diplomatic_development",
		constants.EventConflict:    "cultural_resistance",
		constants.CultureInfluence: "cultural_evolution",
	}

	return &models.EmergentWorldEvent{
		ID:        uuid.New().String(),
		SessionID: culture.Metadata["session_id"].(string),
		EventType: "cultural_response",
		Title:     fmt.Sprintf("%s Responds to Outside Influence", culture.Name),
		Description: fmt.Sprintf("The %s culture has %s in response to %s",
			culture.Name,
			pcs.getResponseDescription(response),
			action.Description),
		Impact: map[string]interface{}{
			"culture_id":      culture.ID,
			"action_type":     action.Type,
			"response_type":   eventTypes[action.Type],
			"acceptance":      response.Acceptance,
			"resistance":      response.Resistance,
			"affected_aspect": response.AffectedAspect,
		},
		AffectedEntities: []string{culture.ID},
		IsPlayerVisible:  true,
		OccurredAt:       time.Now(),
	}
}

func (pcs *ProceduralCultureService) getResponseDescription(response CulturalResponse) string {
	if response.Impact > 0.5 {
		return "embraced the changes enthusiastically"
	} else if response.Impact > 0 {
		return "cautiously adapted to the new influences"
	} else if response.Impact > -0.5 {
		return "resisted the outside influence"
	} else {
		return "strongly rejected the foreign ways"
	}
}

func containsInCulture(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
