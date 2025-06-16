package services

import (
	"context"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// EmailServiceInterface defines email sending operations
type EmailServiceInterface interface {
	Send(ctx context.Context, to []string, subject, body string, isHTML bool) error
	SendWithAttachment(ctx context.Context, to []string, subject, body string, isHTML bool, attachments []Attachment) error
	SendTemplate(ctx context.Context, to []string, templateName string, data interface{}) error
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// ExportServiceInterface defines data export operations
type ExportServiceInterface interface {
	ExportCharacters(ctx context.Context, userID string, characterIDs []string, format string) (interface{}, error)
	ExportCampaigns(ctx context.Context, userID string, campaignIDs []string, format string) (interface{}, error)
	ExportUserData(ctx context.Context, userID string) (interface{}, error)
}

// CleanupServiceInterface defines cleanup operations
type CleanupServiceInterface interface {
	CleanupExpiredTokens(ctx context.Context, olderThan time.Time) (int, error)
	CleanupOldSessions(ctx context.Context, olderThan time.Time) (int, error)
	CleanupOrphanedData(ctx context.Context) (int, error)
	CleanupTempFiles(ctx context.Context, olderThan time.Time) (int, error)
}

// CampaignServiceInterface defines campaign operations
type CampaignServiceInterface interface {
	GetCampaignByID(ctx context.Context, campaignID string) (*Campaign, error)
	GetUserCampaigns(ctx context.Context, userID string) ([]*Campaign, error)
	CreateCampaign(ctx context.Context, campaign *Campaign) error
	UpdateCampaign(ctx context.Context, campaign *Campaign) error
	DeleteCampaign(ctx context.Context, campaignID string) error
}

// Campaign represents a D&D campaign (placeholder)
type Campaign struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AIServiceInterface defines AI-related operations for job processing
type AIServiceInterface interface {
	GenerateContent(ctx context.Context, prompt string, options map[string]interface{}) (string, error)
	GenerateNPC(ctx context.Context, params map[string]interface{}) (*models.NPC, error)
	GenerateLocation(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
	GenerateEncounter(ctx context.Context, params map[string]interface{}) (*models.Encounter, error)
	GenerateQuest(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
	GenerateItem(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
}

// CharacterServiceInterface defines character-related operations for job processing
type CharacterServiceInterface interface {
	GetCharacterByID(ctx context.Context, characterID string) (*models.Character, error)
	GetUserCharacters(ctx context.Context, userID string) ([]*models.Character, error)
	CreateCharacter(ctx context.Context, character *models.Character) error
	UpdateCharacter(ctx context.Context, character *models.Character) error
	DeleteCharacter(ctx context.Context, characterID string) error
	LevelUp(ctx context.Context, characterID string) error
}