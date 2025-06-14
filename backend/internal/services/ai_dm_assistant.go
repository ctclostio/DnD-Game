package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/google/uuid"
)

// AIDMAssistantService handles AI-powered DM assistance.
type AIDMAssistantService struct {
	llmProvider LLMProvider
}

// NewAIDMAssistantService creates a new AI DM assistant service.
func NewAIDMAssistantService(llmProvider LLMProvider) *AIDMAssistantService {
	return &AIDMAssistantService{
		llmProvider: llmProvider,
	}
}

// GenerateNPCDialogue generates contextual dialogue for an NPC.
func (s *AIDMAssistantService) GenerateNPCDialogue(ctx context.Context, req models.NPCDialogueRequest) (string, error) {
	systemPrompt := `You are a Dungeon Master helping to generate NPC dialogue for a D&D game. 
Generate dialogue that:
1. Matches the NPC's personality traits and speaking style
2. Responds appropriately to the player's input
3. Stays in character
4. Advances the story or provides useful information
5. Is engaging and memorable

Keep responses concise (1-3 sentences unless the situation calls for more).
Do not include actions or descriptions, only spoken dialogue.`

	userPrompt := fmt.Sprintf(`NPC: %s
Personality: %s
Speaking Style: %s
Current Situation: %s
Player Said/Did: %s
Previous Context: %s

Generate an appropriate response from this NPC.`,
		req.NPCName,
		strings.Join(req.NPCPersonality, ", "),
		req.DialogueStyle,
		req.Situation,
		req.PlayerInput,
		req.PreviousContext)

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate NPC dialogue: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// GenerateLocationDescription creates immersive location descriptions.
func (s *AIDMAssistantService) GenerateLocationDescription(ctx context.Context, req models.LocationDescriptionRequest) (*models.AILocation, error) {
	systemPrompt := `You are a Dungeon Master creating vivid location descriptions for a D&D game.
Your descriptions should:
1. Paint a clear picture using all five senses
2. Include interesting details that players can interact with
3. Set the appropriate mood and atmosphere
4. Suggest potential actions without being prescriptive
5. Include at least one hidden detail or secret

Return your response as a JSON object with the following structure:
{
  "description": "Main description of the location",
  "atmosphere": "The mood and feeling of the place",
  "notableFeatures": ["feature1", "feature2", "feature3"],
  "availableActions": ["action1", "action2", "action3"],
  "secretsAndHidden": [
    {
      "description": "A hidden detail",
      "discoveryDC": 15,
      "discoveryHint": "What might tip off observant players"
    }
  ],
  "environmentalEffects": "Any ongoing effects like weather, magical auras, etc."
}`

	features := ""
	if len(req.SpecialFeatures) > 0 {
		features = "\nSpecial Features to Include: " + strings.Join(req.SpecialFeatures, ", ")
	}

	userPrompt := fmt.Sprintf(`Create a detailed description for:
Location Type: %s
Location Name: %s
Atmosphere: %s%s
Time of Day: %s
Weather: %s

Make it immersive and interactive.`,
		req.LocationType,
		req.LocationName,
		req.Atmosphere,
		features,
		req.TimeOfDay,
		req.Weather)

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate location description: %w", err)
	}

	// Parse the JSON response.
	var locationData struct {
		Description          string                `json:"description"`
		Atmosphere           string                `json:"atmosphere"`
		NotableFeatures      []string              `json:"notableFeatures"`
		AvailableActions     []string              `json:"availableActions"`
		SecretsAndHidden     []models.SecretDetail `json:"secretsAndHidden"`
		EnvironmentalEffects string                `json:"environmentalEffects"`
	}

	if err := json.Unmarshal([]byte(response), &locationData); err != nil {
		// Fallback to simple text response.
		return &models.AILocation{
			Name:        req.LocationName,
			Type:        req.LocationType,
			Description: response,
			Atmosphere:  req.Atmosphere,
		}, nil
	}

	location := &models.AILocation{
		ID:                   uuid.New(),
		Name:                 req.LocationName,
		Type:                 req.LocationType,
		Description:          locationData.Description,
		Atmosphere:           locationData.Atmosphere,
		NotableFeatures:      locationData.NotableFeatures,
		AvailableActions:     locationData.AvailableActions,
		SecretsAndHidden:     locationData.SecretsAndHidden,
		EnvironmentalEffects: locationData.EnvironmentalEffects,
	}

	return location, nil
}

// GenerateCombatNarration creates dynamic combat descriptions.
func (s *AIDMAssistantService) GenerateCombatNarration(ctx context.Context, req models.CombatNarrationRequest) (string, error) {
	intensity := "normal"
	if req.IsCritical {
		intensity = "epic and dramatic"
	} else if req.TargetHP <= req.TargetMaxHP/4 {
		intensity = "desperate and tense"
	}

	systemPrompt := fmt.Sprintf(`You are narrating combat for a D&D game. 
Create %s descriptions that:
1. Are visceral and exciting without being gratuitously violent
2. Reflect the weapon/spell being used
3. Show the impact and consequences
4. Keep the pace moving
5. Add cinematic flair

Keep it to 1-2 sentences. Be creative and varied.`, intensity)

	var userPrompt string
	if req.IsHit {
		if req.TargetHP <= 0 {
			userPrompt = fmt.Sprintf(`%s defeats %s with %s dealing %d damage. 
Create a dramatic death/defeat description.`,
				req.AttackerName, req.TargetName, req.WeaponOrSpell, req.Damage)
		} else {
			userPrompt = fmt.Sprintf(`%s hits %s with %s for %d damage. 
%s has %d/%d HP remaining. Describe the impact.`,
				req.AttackerName, req.TargetName, req.WeaponOrSpell, req.Damage,
				req.TargetName, req.TargetHP, req.TargetMaxHP)
		}
	} else {
		userPrompt = fmt.Sprintf(`%s misses %s with %s. 
Describe the near-miss in an exciting way.`,
			req.AttackerName, req.TargetName, req.WeaponOrSpell)
	}

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate combat narration: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// GeneratePlotTwist creates unexpected story developments.
func (s *AIDMAssistantService) GeneratePlotTwist(ctx context.Context, currentContext map[string]interface{}) (*models.AIStoryElement, error) {
	systemPrompt := `You are a master storyteller creating plot twists for a D&D campaign.
Your twists should:
1. Be surprising but make sense in hindsight
2. Create new opportunities for adventure
3. Challenge player assumptions
4. Connect to existing story elements
5. Have clear consequences

Return as JSON:
{
  "title": "Brief title for the twist",
  "description": "Detailed explanation of the twist",
  "suggestedTiming": "When/how to reveal this",
  "prerequisites": ["What needs to happen first"],
  "consequences": ["What happens as a result"],
  "foreshadowingHints": ["Subtle hints to drop beforehand"],
  "impactLevel": "minor|moderate|major|campaign-changing"
}`

	contextJSON, _ := json.Marshal(currentContext)
	userPrompt := fmt.Sprintf(`Based on this campaign context:
%s

Generate an engaging plot twist that fits naturally into the story.`, string(contextJSON))

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plot twist: %w", err)
	}

	var twistData struct {
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		SuggestedTiming    string   `json:"suggestedTiming"`
		Prerequisites      []string `json:"prerequisites"`
		Consequences       []string `json:"consequences"`
		ForeshadowingHints []string `json:"foreshadowingHints"`
		ImpactLevel        string   `json:"impactLevel"`
	}

	if err := json.Unmarshal([]byte(response), &twistData); err != nil {
		return nil, fmt.Errorf("failed to parse plot twist: %w", err)
	}

	return &models.AIStoryElement{
		ID:                 uuid.New(),
		Type:               models.StoryElementPlotTwist,
		Title:              twistData.Title,
		Description:        twistData.Description,
		Context:            currentContext,
		ImpactLevel:        twistData.ImpactLevel,
		SuggestedTiming:    twistData.SuggestedTiming,
		Prerequisites:      twistData.Prerequisites,
		Consequences:       twistData.Consequences,
		ForeshadowingHints: twistData.ForeshadowingHints,
	}, nil
}

// GenerateEnvironmentalHazard creates location-appropriate challenges.
func (s *AIDMAssistantService) GenerateEnvironmentalHazard(ctx context.Context, locationType string, difficulty int) (*models.AIEnvironmentalHazard, error) {
	systemPrompt := `You are creating environmental hazards for a D&D game.
Create hazards that:
1. Fit naturally in the environment
2. Provide interesting tactical challenges
3. Can be detected and avoided by clever players
4. Have clear mechanical effects
5. Add tension without being unfair

Return as JSON:
{
  "name": "Hazard name",
  "description": "What it looks like",
  "triggerCondition": "What causes it to activate",
  "effectDescription": "What happens when triggered",
  "mechanicalEffects": {
    "save": "DEX/STR/etc",
    "difficultyClass": 10-20,
    "damage": "dice formula like 2d6",
    "damageType": "fire/cold/etc",
    "additionalEffects": "any status effects"
  },
  "avoidanceHints": "How perceptive players might notice it",
  "isTrap": true/false,
  "isNatural": true/false
}`

	userPrompt := fmt.Sprintf(`Create a hazard for a %s location.
Target difficulty: %d (scale 1-10, where 10 is deadly)
Make it thematically appropriate and mechanically interesting.`, locationType, difficulty)

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate environmental hazard: %w", err)
	}

	var hazardData struct {
		Name              string                 `json:"name"`
		Description       string                 `json:"description"`
		TriggerCondition  string                 `json:"triggerCondition"`
		EffectDescription string                 `json:"effectDescription"`
		MechanicalEffects map[string]interface{} `json:"mechanicalEffects"`
		AvoidanceHints    string                 `json:"avoidanceHints"`
		IsTrap            bool                   `json:"isTrap"`
		IsNatural         bool                   `json:"isNatural"`
	}

	if err := json.Unmarshal([]byte(response), &hazardData); err != nil {
		return nil, fmt.Errorf("failed to parse environmental hazard: %w", err)
	}

	// Extract DC and damage formula from mechanical effects.
	dc := 12 // default
	damageFormula := "1d6"

	if dcVal, ok := hazardData.MechanicalEffects["difficultyClass"].(float64); ok {
		dc = int(dcVal)
	}
	if damage, ok := hazardData.MechanicalEffects["damage"].(string); ok {
		damageFormula = damage
	}

	return &models.AIEnvironmentalHazard{
		ID:                uuid.New(),
		Name:              hazardData.Name,
		Description:       hazardData.Description,
		TriggerCondition:  hazardData.TriggerCondition,
		EffectDescription: hazardData.EffectDescription,
		MechanicalEffects: hazardData.MechanicalEffects,
		DifficultyClass:   dc,
		DamageFormula:     damageFormula,
		AvoidanceHints:    hazardData.AvoidanceHints,
		IsTrap:            hazardData.IsTrap,
		IsNatural:         hazardData.IsNatural,
		IsActive:          true,
	}, nil
}

// GenerateNPC creates a full NPC with personality and motivations.
func (s *AIDMAssistantService) GenerateNPC(ctx context.Context, role string, context map[string]interface{}) (*models.AINPC, error) {
	systemPrompt := `You are creating memorable NPCs for a D&D game.
Create NPCs that:
1. Have clear motivations and goals
2. Possess unique personality traits and mannerisms
3. Speak in a distinctive way
4. Have secrets or hidden depths
5. Can drive the story forward

Return as JSON:
{
  "name": "Full name",
  "race": "D&D race",
  "occupation": "What they do",
  "personalityTraits": ["trait1", "trait2", "trait3"],
  "appearance": "Physical description",
  "voiceDescription": "How they sound",
  "motivations": "What drives them",
  "secrets": "What they're hiding",
  "dialogueStyle": "How they speak (accent, vocabulary, mannerisms)",
  "relationshipToParty": "Initial attitude/connection"
}`

	contextJSON, _ := json.Marshal(context)
	userPrompt := fmt.Sprintf(`Create an NPC for the role of: %s

Campaign context: %s

Make them interesting and three-dimensional.`, role, string(contextJSON))

	response, err := s.llmProvider.GenerateCompletion(ctx, userPrompt, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate NPC: %w", err)
	}

	var npcData struct {
		Name                string   `json:"name"`
		Race                string   `json:"race"`
		Occupation          string   `json:"occupation"`
		PersonalityTraits   []string `json:"personalityTraits"`
		Appearance          string   `json:"appearance"`
		VoiceDescription    string   `json:"voiceDescription"`
		Motivations         string   `json:"motivations"`
		Secrets             string   `json:"secrets"`
		DialogueStyle       string   `json:"dialogueStyle"`
		RelationshipToParty string   `json:"relationshipToParty"`
	}

	if err := json.Unmarshal([]byte(response), &npcData); err != nil {
		return nil, fmt.Errorf("failed to parse NPC data: %w", err)
	}

	return &models.AINPC{
		ID:                  uuid.New(),
		Name:                npcData.Name,
		Race:                npcData.Race,
		Occupation:          npcData.Occupation,
		PersonalityTraits:   npcData.PersonalityTraits,
		Appearance:          npcData.Appearance,
		VoiceDescription:    npcData.VoiceDescription,
		Motivations:         npcData.Motivations,
		Secrets:             npcData.Secrets,
		DialogueStyle:       npcData.DialogueStyle,
		RelationshipToParty: npcData.RelationshipToParty,
		GeneratedDialogue:   []models.DialogueEntry{},
	}, nil
}
