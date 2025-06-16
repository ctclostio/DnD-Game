package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/ctclostio/DnD-Game/backend/internal/config"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// JobType represents different types of background jobs
type JobType string

const (
	// Job types
	JobTypeAIContentGeneration JobType = "ai:content:generate"
	JobTypeEmailNotification   JobType = "email:send"
	JobTypeReportGeneration    JobType = "report:generate"
	JobTypeDataExport          JobType = "data:export"
	JobTypeCharacterBackup     JobType = "character:backup"
	JobTypeCampaignBackup      JobType = "campaign:backup"
	JobTypeImageOptimization   JobType = "image:optimize"
	JobTypeAnalyticsProcess    JobType = "analytics:process"
	JobTypeCleanupExpired      JobType = "cleanup:expired"
	
	// Queue names
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

// JobQueue manages background job processing
type JobQueue struct {
	client    *asynq.Client
	server    *asynq.Server
	mux       *asynq.ServeMux
	redisOpt  asynq.RedisClientOpt
	logger    *logger.LoggerV2
	handlers  map[JobType]JobHandler
	mu        sync.RWMutex
}

// JobHandler processes a specific job type
type JobHandler func(ctx context.Context, task *asynq.Task) error

// JobOptions contains options for enqueuing a job
type JobOptions struct {
	MaxRetry     int           // Maximum number of retries
	Queue        string        // Queue name (critical, default, low)
	ProcessAt    time.Time     // Schedule job for specific time
	ProcessIn    time.Duration // Schedule job after duration
	Deadline     time.Time     // Job must complete by this time
	UniqueFor    time.Duration // Ensure uniqueness for duration
	Retention    time.Duration // How long to keep job results
	TaskID       string        // Custom task ID
}

// DefaultJobOptions returns default job options
func DefaultJobOptions() JobOptions {
	return JobOptions{
		MaxRetry:  3,
		Queue:     QueueDefault,
		Retention: 24 * time.Hour,
	}
}

// NewJobQueue creates a new job queue
func NewJobQueue(cfg *config.RedisConfig, log *logger.LoggerV2) (*JobQueue, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config is required")
	}

	// Create Asynq client
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := asynq.NewClient(redisOpt)

	// Create server with configuration
	serverConfig := asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			QueueCritical: 6,
			QueueDefault:  3,
			QueueLow:      1,
		},
		StrictPriority: true,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			if log != nil {
				log.Error().
					Err(err).
					Str("task_type", task.Type()).
					Bytes("payload", task.Payload()).
					Msg("Task processing failed")
			}
		}),
		RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
			// Exponential backoff with jitter
			return time.Duration(n*n) * time.Second
		},
		Logger: &asynqLogger{logger: log},
		HealthCheckFunc: func(err error) {
			if err != nil && log != nil {
				log.Error().Err(err).Msg("Asynq health check failed")
			}
		},
	}

	server := asynq.NewServer(redisOpt, serverConfig)
	mux := asynq.NewServeMux()

	jq := &JobQueue{
		client:   client,
		server:   server,
		mux:      mux,
		redisOpt: redisOpt,
		logger:   log,
		handlers: make(map[JobType]JobHandler),
	}

	return jq, nil
}

// RegisterHandler registers a handler for a job type
func (jq *JobQueue) RegisterHandler(jobType JobType, handler JobHandler) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	jq.handlers[jobType] = handler
	
	// Register with Asynq mux
	jq.mux.HandleFunc(string(jobType), func(ctx context.Context, task *asynq.Task) error {
		start := time.Now()
		
		// Log job start
		if jq.logger != nil {
			jq.logger.Info().
				Str("job_type", string(jobType)).
				Str("task_id", task.ResultWriter().TaskID()).
				Int("payload_size", len(task.Payload())).
				Msg("Processing job")
		}

		// Process job
		err := handler(ctx, task)
		
		// Log job completion
		if jq.logger != nil {
			event := jq.logger.Info().
				Str("job_type", string(jobType)).
				Str("task_id", task.ResultWriter().TaskID()).
				Dur("duration", time.Since(start))

			if err != nil {
				event.Err(err).Msg("Job failed")
			} else {
				event.Msg("Job completed")
			}
		}

		return err
	})

	if jq.logger != nil {
		jq.logger.Info().
			Str("job_type", string(jobType)).
			Msg("Registered job handler")
	}
}

// Enqueue adds a job to the queue
func (jq *JobQueue) Enqueue(ctx context.Context, jobType JobType, payload interface{}, opts ...JobOptions) (*asynq.TaskInfo, error) {
	// Serialize payload
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create task
	task := asynq.NewTask(string(jobType), data)

	// Apply options
	opt := DefaultJobOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Build Asynq options
	var taskOpts []asynq.Option
	
	if opt.MaxRetry > 0 {
		taskOpts = append(taskOpts, asynq.MaxRetry(opt.MaxRetry))
	}
	
	if opt.Queue != "" {
		taskOpts = append(taskOpts, asynq.Queue(opt.Queue))
	}
	
	if !opt.ProcessAt.IsZero() {
		taskOpts = append(taskOpts, asynq.ProcessAt(opt.ProcessAt))
	} else if opt.ProcessIn > 0 {
		taskOpts = append(taskOpts, asynq.ProcessIn(opt.ProcessIn))
	}
	
	if !opt.Deadline.IsZero() {
		taskOpts = append(taskOpts, asynq.Deadline(opt.Deadline))
	}
	
	if opt.UniqueFor > 0 {
		taskOpts = append(taskOpts, asynq.Unique(opt.UniqueFor))
	}
	
	if opt.Retention > 0 {
		taskOpts = append(taskOpts, asynq.Retention(opt.Retention))
	}
	
	if opt.TaskID != "" {
		taskOpts = append(taskOpts, asynq.TaskID(opt.TaskID))
	}

	// Enqueue task
	info, err := jq.client.EnqueueContext(ctx, task, taskOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	if jq.logger != nil {
		jq.logger.Info().
			Str("job_type", string(jobType)).
			Str("task_id", info.ID).
			Str("queue", info.Queue).
			Str("state", string(info.State)).
			Msg("Job enqueued")
	}

	return info, nil
}

// Start begins processing jobs
func (jq *JobQueue) Start() error {
	if jq.logger != nil {
		jq.logger.Info().Msg("Starting job queue processor")
	}

	return jq.server.Start(jq.mux)
}

// Stop gracefully stops the job processor
func (jq *JobQueue) Stop() error {
	if jq.logger != nil {
		jq.logger.Info().Msg("Stopping job queue processor")
	}

	jq.server.Shutdown()
	return jq.client.Close()
}

// GetTaskInfo retrieves information about a task
func (jq *JobQueue) GetTaskInfo(taskID string) (*asynq.TaskInfo, error) {
	inspector := asynq.NewInspector(jq.redisOpt)
	defer inspector.Close()

	return inspector.GetTaskInfo(QueueDefault, taskID)
}

// CancelTask cancels a scheduled or retrying task
func (jq *JobQueue) CancelTask(taskID string) error {
	inspector := asynq.NewInspector(jq.redisOpt)
	defer inspector.Close()

	return inspector.DeleteTask(QueueDefault, taskID)
}

// GetQueueStats returns statistics for all queues
func (jq *JobQueue) GetQueueStats() (map[string]*asynq.QueueInfo, error) {
	inspector := asynq.NewInspector(jq.redisOpt)
	defer inspector.Close()

	queues, err := inspector.Queues()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]*asynq.QueueInfo)
	for _, q := range queues {
		info, err := inspector.GetQueueInfo(q)
		if err != nil {
			return nil, err
		}
		stats[q] = info
	}

	return stats, nil
}

// HealthCheck verifies the job queue is functional
func (jq *JobQueue) HealthCheck(ctx context.Context) error {
	// Try to get queue stats
	stats, err := jq.GetQueueStats()
	if err != nil {
		return fmt.Errorf("failed to get queue stats: %w", err)
	}

	// Check if we can access all expected queues
	expectedQueues := []string{QueueCritical, QueueDefault, QueueLow}
	for _, q := range expectedQueues {
		if _, ok := stats[q]; !ok {
			return fmt.Errorf("queue %s not found", q)
		}
	}

	return nil
}

// asynqLogger adapts our logger to Asynq's logger interface
type asynqLogger struct {
	logger *logger.LoggerV2
}

func (l *asynqLogger) Debug(args ...interface{}) {
	if l.logger != nil {
		l.logger.Debug().Msg(fmt.Sprint(args...))
	}
}

func (l *asynqLogger) Info(args ...interface{}) {
	if l.logger != nil {
		l.logger.Info().Msg(fmt.Sprint(args...))
	}
}

func (l *asynqLogger) Warn(args ...interface{}) {
	if l.logger != nil {
		l.logger.Warn().Msg(fmt.Sprint(args...))
	}
}

func (l *asynqLogger) Error(args ...interface{}) {
	if l.logger != nil {
		l.logger.Error().Msg(fmt.Sprint(args...))
	}
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	if l.logger != nil {
		l.logger.Fatal().Msg(fmt.Sprint(args...))
	}
}

// JobPayload is a generic payload structure
type JobPayload struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Common job payloads

// EmailPayload contains email job data
type EmailPayload struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	HTML    bool     `json:"html"`
}

// AIGenerationPayload contains AI content generation data
type AIGenerationPayload struct {
	UserID      string `json:"user_id"`
	Type        string `json:"type"` // npc, quest, item, etc.
	Prompt      string `json:"prompt"`
	Context     string `json:"context"`
	CallbackURL string `json:"callback_url,omitempty"`
}

// ExportPayload contains data export job information
type ExportPayload struct {
	UserID       string   `json:"user_id"`
	ResourceType string   `json:"resource_type"` // character, campaign, etc.
	ResourceIDs  []string `json:"resource_ids"`
	Format       string   `json:"format"` // json, pdf, etc.
	Email        string   `json:"email"`  // Where to send the export
}

// BackupPayload contains backup job information
type BackupPayload struct {
	UserID       string `json:"user_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

// CleanupPayload contains cleanup job information
type CleanupPayload struct {
	Type      string    `json:"type"` // expired_tokens, old_sessions, etc.
	OlderThan time.Time `json:"older_than"`
}