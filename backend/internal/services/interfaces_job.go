package services

import (
	"context"
	"time"
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