package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// JobHandlers contains all job handler implementations
type JobHandlers struct {
	logger           *logger.LoggerV2
	aiService        services.AIServiceInterface
	emailService     services.EmailServiceInterface
	characterService services.CharacterServiceInterface
	campaignService  services.CampaignServiceInterface
	exportService    services.ExportServiceInterface
	cleanupService   services.CleanupServiceInterface
}

// NewJobHandlers creates a new job handlers instance
func NewJobHandlers(
	logger *logger.LoggerV2,
	aiService services.AIServiceInterface,
	emailService services.EmailServiceInterface,
	characterService services.CharacterServiceInterface,
	campaignService services.CampaignServiceInterface,
	exportService services.ExportServiceInterface,
	cleanupService services.CleanupServiceInterface,
) *JobHandlers {
	return &JobHandlers{
		logger:           logger,
		aiService:        aiService,
		emailService:     emailService,
		characterService: characterService,
		campaignService:  campaignService,
		exportService:    exportService,
		cleanupService:   cleanupService,
	}
}

// RegisterAll registers all job handlers with the queue
func (jh *JobHandlers) RegisterAll(queue *JobQueue) {
	queue.RegisterHandler(JobTypeAIContentGeneration, jh.HandleAIGeneration)
	queue.RegisterHandler(JobTypeEmailNotification, jh.HandleEmailNotification)
	queue.RegisterHandler(JobTypeReportGeneration, jh.HandleReportGeneration)
	queue.RegisterHandler(JobTypeDataExport, jh.HandleDataExport)
	queue.RegisterHandler(JobTypeCharacterBackup, jh.HandleCharacterBackup)
	queue.RegisterHandler(JobTypeCampaignBackup, jh.HandleCampaignBackup)
	queue.RegisterHandler(JobTypeImageOptimization, jh.HandleImageOptimization)
	queue.RegisterHandler(JobTypeAnalyticsProcess, jh.HandleAnalyticsProcess)
	queue.RegisterHandler(JobTypeCleanupExpired, jh.HandleCleanupExpired)
}

// HandleAIGeneration processes AI content generation jobs
func (jh *JobHandlers) HandleAIGeneration(ctx context.Context, task *asynq.Task) error {
	var payload AIGenerationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Log the job
	jh.logger.Info().
		Str("user_id", payload.UserID).
		Str("type", payload.Type).
		Str("task_id", task.ResultWriter().TaskID()).
		Msg("Processing AI generation job")

	// Process based on type
	var result interface{}
	var err error

	switch payload.Type {
	case "npc":
		// Generate NPC using AI service
		params := map[string]interface{}{
			"prompt": payload.Prompt,
			"context": payload.Context,
		}
		npc, npcErr := jh.aiService.GenerateNPC(ctx, params)
		if npcErr != nil {
			err = npcErr
		} else {
			result = npc
		}
		
	case "quest":
		// Generate quest
		params := map[string]interface{}{
			"prompt": payload.Prompt,
			"context": payload.Context,
		}
		result, err = jh.aiService.GenerateQuest(ctx, params)
		
	case "item":
		// Generate item
		params := map[string]interface{}{
			"prompt": payload.Prompt,
			"context": payload.Context,
		}
		result, err = jh.aiService.GenerateItem(ctx, params)
		
	case "encounter":
		// Generate encounter
		params := map[string]interface{}{
			"prompt": payload.Prompt,
			"context": payload.Context,
		}
		encounter, encErr := jh.aiService.GenerateEncounter(ctx, params)
		if encErr != nil {
			err = encErr
		} else {
			result = encounter
		}
		
	default:
		return fmt.Errorf("unknown AI generation type: %s", payload.Type)
	}

	if err != nil {
		return fmt.Errorf("AI generation failed: %w", err)
	}

	// Store result for retrieval
	resultData, _ := json.Marshal(result)
	if _, err := task.ResultWriter().Write(resultData); err != nil {
		jh.logger.Error().Err(err).Msg("Failed to write task result")
	}

	// If callback URL provided, notify completion
	if payload.CallbackURL != "" {
		// TODO: Implement webhook notification
		jh.logger.Debug().
			Str("callback_url", payload.CallbackURL).
			Msg("Would notify callback URL")
	}

	return nil
}

// HandleEmailNotification sends email notifications
func (jh *JobHandlers) HandleEmailNotification(ctx context.Context, task *asynq.Task) error {
	var payload EmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Log the job
	jh.logger.Info().
		Strs("to", payload.To).
		Str("subject", payload.Subject).
		Msg("Sending email notification")

	// Send email
	if err := jh.emailService.Send(ctx, payload.To, payload.Subject, payload.Body, payload.HTML); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// HandleReportGeneration generates reports
func (jh *JobHandlers) HandleReportGeneration(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		UserID     string    `json:"user_id"`
		ReportType string    `json:"report_type"`
		StartDate  time.Time `json:"start_date"`
		EndDate    time.Time `json:"end_date"`
		Format     string    `json:"format"` // pdf, csv, json
		Email      string    `json:"email"`
	}
	
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	jh.logger.Info().
		Str("user_id", payload.UserID).
		Str("report_type", payload.ReportType).
		Str("format", payload.Format).
		Msg("Generating report")

	// TODO: Implement report generation based on type
	// For now, just log
	jh.logger.Debug().Msg("Report generation not yet implemented")

	// Send email with report
	if payload.Email != "" {
		emailPayload := EmailPayload{
			To:      []string{payload.Email},
			Subject: fmt.Sprintf("Your %s Report is Ready", payload.ReportType),
			Body:    "Your report has been generated and is attached to this email.",
			HTML:    true,
		}
		
		return jh.HandleEmailNotification(ctx, asynq.NewTask(string(JobTypeEmailNotification), mustMarshal(emailPayload)))
	}

	return nil
}

// HandleDataExport exports user data
func (jh *JobHandlers) HandleDataExport(ctx context.Context, task *asynq.Task) error {
	var payload ExportPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	jh.logger.Info().
		Str("user_id", payload.UserID).
		Str("resource_type", payload.ResourceType).
		Str("format", payload.Format).
		Int("resource_count", len(payload.ResourceIDs)).
		Msg("Processing data export")

	// Export based on resource type
	var exportData interface{}
	var err error

	switch payload.ResourceType {
	case "character":
		if jh.exportService != nil {
			exportData, err = jh.exportService.ExportCharacters(ctx, payload.UserID, payload.ResourceIDs, payload.Format)
		}
		
	case "campaign":
		if jh.exportService != nil {
			exportData, err = jh.exportService.ExportCampaigns(ctx, payload.UserID, payload.ResourceIDs, payload.Format)
		}
		
	default:
		return fmt.Errorf("unknown export type: %s", payload.ResourceType)
	}

	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	// Store result
	if exportData != nil {
		resultData, err := json.Marshal(exportData)
		if err != nil {
			return fmt.Errorf("failed to marshal export data: %w", err)
		}
		if _, err := task.ResultWriter().Write(resultData); err != nil {
			return fmt.Errorf("failed to write result: %w", err)
		}
	}

	// Send email with export link
	if payload.Email != "" {
		// TODO: Generate download link
		downloadLink := fmt.Sprintf("https://api.example.com/exports/%s", task.ResultWriter().TaskID())
		
		emailPayload := EmailPayload{
			To:      []string{payload.Email},
			Subject: fmt.Sprintf("Your %s Export is Ready", payload.ResourceType),
			Body:    fmt.Sprintf("Your export is ready for download: %s\n\nThis link will expire in 24 hours.", downloadLink),
			HTML:    false,
		}
		
		return jh.HandleEmailNotification(ctx, asynq.NewTask(string(JobTypeEmailNotification), mustMarshal(emailPayload)))
	}

	return nil
}

// HandleCharacterBackup backs up character data
func (jh *JobHandlers) HandleCharacterBackup(ctx context.Context, task *asynq.Task) error {
	var payload BackupPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	jh.logger.Info().
		Str("user_id", payload.UserID).
		Str("character_id", payload.ResourceID).
		Msg("Backing up character")

	// Get character data
	character, err := jh.characterService.GetCharacterByID(ctx, payload.ResourceID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Verify ownership
	if character.UserID != payload.UserID {
		return fmt.Errorf("unauthorized: user does not own character")
	}

	// Create backup
	backup := map[string]interface{}{
		"character":   character,
		"timestamp":   time.Now().UTC(),
		"version":     "1.0",
		"backup_type": "automatic",
	}

	// TODO: Store backup in object storage
	backupData, _ := json.Marshal(backup)
	jh.logger.Debug().
		Int("backup_size", len(backupData)).
		Msg("Character backup created")

	return nil
}

// HandleCampaignBackup backs up campaign data
func (jh *JobHandlers) HandleCampaignBackup(ctx context.Context, task *asynq.Task) error {
	var payload BackupPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	jh.logger.Info().
		Str("user_id", payload.UserID).
		Str("campaign_id", payload.ResourceID).
		Msg("Backing up campaign")

	// TODO: Implement campaign backup
	jh.logger.Debug().Msg("Campaign backup not yet implemented")

	return nil
}

// HandleImageOptimization optimizes uploaded images
func (jh *JobHandlers) HandleImageOptimization(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		UserID    string   `json:"user_id"`
		ImageURLs []string `json:"image_urls"`
		MaxWidth  int      `json:"max_width"`
		MaxHeight int      `json:"max_height"`
		Quality   int      `json:"quality"`
	}
	
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	jh.logger.Info().
		Str("user_id", payload.UserID).
		Int("image_count", len(payload.ImageURLs)).
		Msg("Optimizing images")

	// TODO: Implement image optimization
	// - Download images
	// - Resize/compress
	// - Upload to CDN
	// - Update references

	jh.logger.Debug().Msg("Image optimization not yet implemented")

	return nil
}

// HandleAnalyticsProcess processes analytics data
func (jh *JobHandlers) HandleAnalyticsProcess(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		Type      string    `json:"type"` // daily, weekly, monthly
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	}
	
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	jh.logger.Info().
		Str("type", payload.Type).
		Time("start_time", payload.StartTime).
		Time("end_time", payload.EndTime).
		Msg("Processing analytics")

	// TODO: Implement analytics processing
	// - Aggregate user activity
	// - Calculate metrics
	// - Store results

	jh.logger.Debug().Msg("Analytics processing not yet implemented")

	return nil
}

// HandleCleanupExpired cleans up expired data
func (jh *JobHandlers) HandleCleanupExpired(ctx context.Context, task *asynq.Task) error {
	var payload CleanupPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	jh.logger.Info().
		Str("type", payload.Type).
		Time("older_than", payload.OlderThan).
		Msg("Cleaning up expired data")

	if jh.cleanupService == nil {
		jh.logger.Debug().Msg("Cleanup service not available")
		return nil
	}

	var count int
	var err error

	switch payload.Type {
	case "expired_tokens":
		count, err = jh.cleanupService.CleanupExpiredTokens(ctx, payload.OlderThan)
		
	case "old_sessions":
		count, err = jh.cleanupService.CleanupOldSessions(ctx, payload.OlderThan)
		
	case "orphaned_data":
		count, err = jh.cleanupService.CleanupOrphanedData(ctx)
		
	default:
		return fmt.Errorf("unknown cleanup type: %s", payload.Type)
	}

	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	jh.logger.Info().
		Str("type", payload.Type).
		Int("items_cleaned", count).
		Msg("Cleanup completed")

	return nil
}

// Helper function to marshal data
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}