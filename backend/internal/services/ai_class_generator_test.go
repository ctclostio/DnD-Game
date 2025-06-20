package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services/mocks"
)

// Test error messages
const (
	errFailedToGenerateClass = "failed to generate class"
	errFailedToParseClassResponse = "failed to parse class response"
	errClassValidationFailed = "class validation failed"
	errAPIError = "API error"
	errInvalidHitDie = "invalid hit die:"
	errInvalidSavingThrows = "classes must have exactly 2 saving throw proficiencies"
	errNoLevel1Features = "class must have at least one level 1 feature"
	errInvalidPrimaryAbility = "invalid primary ability"
	testFeature1Name = "Feature 1"
)

func TestNewAIClassGenerator(t *testing.T) {
	mockLLM := &mocks.MockLLMProvider{}
	generator := NewAIClassGenerator(mockLLM)

	require.NotNil(t, generator)
	require.Equal(t, mockLLM, generator.llmProvider)
}

func TestAIClassGenerator_GenerateCustomClass(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		// Prepare valid AI response
		aiClass := map[string]interface{}{
			"name":                     "Shadowdancer",
			"description":              "Masters of stealth and shadow magic",
			"hitDie":                   8,
			"primaryAbility":           "Dexterity",
			"savingThrowProficiencies": []string{"Dexterity", "Charisma"},
			"skillProficiencies":       []string{"Acrobatics", "Deception", "Stealth"},
			"skillChoices":             3,
			"armorProficiencies":       []string{"Light armor"},
			"weaponProficiencies":      []string{"Simple weapons", "shortswords", "rapiers", "hand crossbows"},
			"toolProficiencies":        []string{"Thieves' tools"},
			"classFeatures": []map[string]interface{}{
				{
					"level":       1,
					"name":        "Shadow Step",
					"description": "You can teleport between shadows",
				},
				{
					"level":       1,
					"name":        "Darkvision",
					"description": "You can see in darkness within 60 feet",
				},
				{
					"level":       3,
					"name":        "Shadow Clone",
					"description": "Create an illusory duplicate of yourself",
				},
			},
			"spellcastingAbility": "Charisma",
			"spellList":           []string{"minor illusion", "silent image", "darkness", "shadow blade"},
			"ritualCasting":       false,
			"subclasses": []map[string]interface{}{
				{
					"name":        "Path of Shadows",
					"description": "Focus on stealth and assassination",
					"features": []map[string]interface{}{
						{
							"level":       3,
							"name":        "Assassinate",
							"description": "Critical hits on surprised creatures",
						},
					},
				},
			},
		}

		aiResponse, _ := json.Marshal(aiClass)
		mockLLM := &mocks.MockLLMProvider{}
		mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything, mock.Anything).
			Return(string(aiResponse), nil)

		generator := NewAIClassGenerator(mockLLM)

		req := CustomClassRequest{
			Name:        "Shadowdancer",
			Description: "A class that manipulates shadows",
			Role:        "stealth damage dealer",
			Style:       "balanced",
			Features:    "shadow magic and stealth",
		}

		ctx := context.Background()
		class, err := generator.GenerateCustomClass(ctx, &req)

		require.NoError(t, err)
		require.NotNil(t, class)
		require.Equal(t, "Shadowdancer", class.Name)
		require.Equal(t, "Masters of stealth and shadow magic", class.Description)
		require.Equal(t, 8, class.HitDie)
		require.Equal(t, "Dexterity", class.PrimaryAbility)
		require.Len(t, class.ClassFeatures, 3)

		mockLLM.AssertExpectations(t)
	})

	t.Run("LLM provider error", func(t *testing.T) {
		mockLLM := &mocks.MockLLMProvider{}
		mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything, mock.Anything).
			Return("", errors.New(errAPIError))

		generator := NewAIClassGenerator(mockLLM)

		req := CustomClassRequest{
			Name:        "Test Class",
			Description: "Test description",
			Role:        "tank",
		}

		ctx := context.Background()
		class, err := generator.GenerateCustomClass(ctx, &req)

		require.Error(t, err)
		require.Nil(t, class)
		require.Contains(t, err.Error(), errFailedToGenerateClass)

		mockLLM.AssertExpectations(t)
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		mockLLM := &mocks.MockLLMProvider{}
		mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything, mock.Anything).
			Return("invalid json", nil)

		generator := NewAIClassGenerator(mockLLM)

		req := CustomClassRequest{
			Name:        "Test Class",
			Description: "Test description",
			Role:        "healer",
		}

		ctx := context.Background()
		class, err := generator.GenerateCustomClass(ctx, &req)

		require.Error(t, err)
		require.Nil(t, class)
		require.Contains(t, err.Error(), errFailedToParseClassResponse)

		mockLLM.AssertExpectations(t)
	})

	t.Run("validation failure - overpowered", func(t *testing.T) {
		// Create an overpowered class
		aiClass := map[string]interface{}{
			"name":                     "Godslayer",
			"description":              "Too powerful",
			"hitDie":                   12,
			"primaryAbility":           "Strength",
			"savingThrowProficiencies": []string{"Strength", "Dexterity", "Constitution"}, // Too many saves
			"skillProficiencies":       []string{},
			"skillChoices":             2,
			"classFeatures": []map[string]interface{}{
				{"level": 1, "name": "Divine Strike", "description": "Deal extra damage"},
				{"level": 1, "name": "Immortality", "description": "Cannot die"},
				{"level": 1, "name": "All-Seeing", "description": "See everything"},
				{"level": 1, "name": "Time Stop", "description": "Stop time"},
				{"level": 1, "name": "Reality Warp", "description": "Change reality"},
			},
		}

		aiResponse, _ := json.Marshal(aiClass)
		mockLLM := &mocks.MockLLMProvider{}
		mockLLM.On("GenerateCompletion", mock.Anything, mock.Anything, mock.Anything).
			Return(string(aiResponse), nil)

		generator := NewAIClassGenerator(mockLLM)

		req := CustomClassRequest{
			Name:        "Godslayer",
			Description: "An overpowered class",
			Role:        "damage dealer",
			Style:       "powerful",
		}

		ctx := context.Background()
		class, err := generator.GenerateCustomClass(ctx, &req)

		require.Error(t, err)
		require.Nil(t, class)
		require.Contains(t, err.Error(), errClassValidationFailed)

		mockLLM.AssertExpectations(t)
	})
}

func TestAIClassGenerator_validateClass(t *testing.T) {
	generator := NewAIClassGenerator(&mocks.MockLLMProvider{})

	tests := []struct {
		name    string
		class   *models.CustomClass
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid balanced class",
			class: &models.CustomClass{
				Name:                     "Valid Class",
				Description:              "A balanced class",
				HitDie:                   8,
				PrimaryAbility:           "Wisdom",
				BalanceScore:             6,
				SavingThrowProficiencies: []string{"Wisdom", "Charisma"},
				SkillChoices:             2,
				ClassFeatures: []models.ClassFeature{
					{Level: 1, Name: testFeature1Name},
					{Level: 3, Name: "Feature 2"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid hit dice",
			class: &models.CustomClass{
				Name:                     "Bad Dice",
				HitDie:                   13,
				PrimaryAbility:           "Strength",
				SavingThrowProficiencies: []string{"Strength", "Constitution"},
			},
			wantErr: true,
			errMsg:  errInvalidHitDie,
		},
		{
			name: "no saving throws",
			class: &models.CustomClass{
				Name:                     "No Saves",
				HitDie:                   8,
				PrimaryAbility:           "Intelligence",
				SavingThrowProficiencies: []string{},
			},
			wantErr: true,
			errMsg:  errInvalidSavingThrows,
		},
		{
			name: "no level 1 features",
			class: &models.CustomClass{
				Name:                     "No Features",
				HitDie:                   8,
				PrimaryAbility:           "Wisdom",
				SavingThrowProficiencies: []string{"Wisdom", "Charisma"},
				ClassFeatures: []models.ClassFeature{
					{Level: 3, Name: testFeature1Name},
				},
			},
			wantErr: true,
			errMsg:  errNoLevel1Features,
		},
		{
			name: "invalid primary ability",
			class: &models.CustomClass{
				Name:                     "Bad Ability",
				HitDie:                   8,
				PrimaryAbility:           "luck",
				SavingThrowProficiencies: []string{"Wisdom", "Charisma"},
			},
			wantErr: true,
			errMsg:  errInvalidPrimaryAbility,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.validateClass(tt.class)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAIClassGenerator_calculateBalanceScore(t *testing.T) {
	generator := NewAIClassGenerator(&mocks.MockLLMProvider{})

	tests := []struct {
		name          string
		class         *models.CustomClass
		expectedScore int
	}{
		{
			name: "basic fighter-like class",
			class: &models.CustomClass{
				HitDie:         10,
				PrimaryAbility: "Strength",
				ClassFeatures: []models.ClassFeature{
					{Level: 1, Name: "Fighting Style"},
					{Level: 2, Name: "Second Wind"},
				},
			},
			expectedScore: 5,
		},
		{
			name: "caster class",
			class: &models.CustomClass{
				HitDie:              6,
				PrimaryAbility:      "Intelligence",
				SpellcastingAbility: "Intelligence",
				ClassFeatures: []models.ClassFeature{
					{Level: 1, Name: "Spellcasting"},
				},
			},
			expectedScore: 4,
		},
		{
			name: "hybrid class with many features",
			class: &models.CustomClass{
				HitDie:              8,
				PrimaryAbility:      "Charisma",
				SpellcastingAbility: "Charisma",
				ClassFeatures: []models.ClassFeature{
					{Level: 1, Name: testFeature1Name},
					{Level: 1, Name: "Feature 2"},
					{Level: 3, Name: "Feature 3"},
					{Level: 5, Name: "Feature 4"},
					{Level: 7, Name: "Feature 5"},
				},
			},
			expectedScore: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := generator.calculateBalanceScore(tt.class)
			require.Equal(t, tt.expectedScore, score)
		})
	}
}
