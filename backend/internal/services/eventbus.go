package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// InMemoryEventBus is a simple in-memory event bus implementation
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
	logger   *logger.LoggerV2
}

// NewEventBus creates a new event bus
func NewEventBus(log *logger.LoggerV2) EventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
		logger:   log,
	}
}

// Publish sends an event to all registered handlers
func (eb *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	eb.mu.RLock()
	handlers, exists := eb.handlers[event.Type()]
	eb.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		// No handlers registered for this event type
		return nil
	}

	// Execute handlers asynchronously
	for _, handler := range handlers {
		go eb.executeHandler(ctx, event, handler)
	}

	return nil
}

// executeHandler runs a single event handler with panic recovery and error logging
func (eb *InMemoryEventBus) executeHandler(ctx context.Context, event Event, handler EventHandler) {
	defer eb.recoverHandlerPanic(ctx, event)

	if err := handler(ctx, event); err != nil {
		eb.logHandlerError(ctx, event, err)
	}
}

// recoverHandlerPanic recovers from panics in event handlers
func (eb *InMemoryEventBus) recoverHandlerPanic(ctx context.Context, event Event) {
	if r := recover(); r != nil {
		if eb.logger != nil {
			eb.logger.WithContext(ctx).
				Error().
				Interface("panic", r).
				Str("event_type", event.Type()).
				Msg("Event handler panic")
		}
	}
}

// logHandlerError logs errors from event handlers
func (eb *InMemoryEventBus) logHandlerError(ctx context.Context, event Event, err error) {
	if eb.logger != nil {
		eb.logger.WithContext(ctx).
			Error().
			Err(err).
			Str("event_type", event.Type()).
			Msg("Event handler error")
	}
}

// Subscribe registers a handler for a specific event type
func (eb *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	return nil
}

// NewEvent creates a new event
func NewEvent(eventType string, data interface{}) Event {
	return BaseEvent{
		EventType: eventType,
		EventTime: time.Now(),
		EventData: data,
	}
}

// Example domain events

// CombatStartedEvent is emitted when combat begins
type CombatStartedEvent struct {
	BaseEvent
	CombatID  string `json:"combat_id"`
	SessionID string `json:"session_id"`
}

// CombatEndedEvent is emitted when combat ends
type CombatEndedEvent struct {
	BaseEvent
	CombatID  string `json:"combat_id"`
	SessionID string `json:"session_id"`
	Victory   bool   `json:"victory"`
}

// CharacterLeveledEvent is emitted when a character levels up
type CharacterLeveledEvent struct {
	BaseEvent
	CharacterID string `json:"character_id"`
	OldLevel    int    `json:"old_level"`
	NewLevel    int    `json:"new_level"`
}

// FactionRelationChangedEvent is emitted when faction relationships change
type FactionRelationChangedEvent struct {
	BaseEvent
	Faction1ID  string  `json:"faction1_id"`
	Faction2ID  string  `json:"faction2_id"`
	OldRelation float64 `json:"old_relation"`
	NewRelation float64 `json:"new_relation"`
}
