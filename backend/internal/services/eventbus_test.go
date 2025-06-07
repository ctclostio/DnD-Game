package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestInMemoryEventBus_Subscribe(t *testing.T) {
	t.Run("successful subscription", func(t *testing.T) {
		logger := log.New(bytes.NewBuffer(nil), "", 0)
		eventBus := NewEventBus(logger)
		
		handler := func(ctx context.Context, event Event) error {
			return nil
		}
		
		err := eventBus.Subscribe(EventCombatStarted, handler)
		
		require.NoError(t, err)
		
		// Verify handler was added
		eb := eventBus.(*InMemoryEventBus)
		eb.mu.RLock()
		handlers, exists := eb.handlers[EventCombatStarted]
		eb.mu.RUnlock()
		
		require.True(t, exists)
		require.Len(t, handlers, 1)
	})

	t.Run("multiple handlers for same event", func(t *testing.T) {
		logger := log.New(bytes.NewBuffer(nil), "", 0)
		eventBus := NewEventBus(logger)
		
		handler1 := func(ctx context.Context, event Event) error { return nil }
		handler2 := func(ctx context.Context, event Event) error { return nil }
		handler3 := func(ctx context.Context, event Event) error { return nil }
		
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, handler1))
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, handler2))
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, handler3))
		
		// Verify all handlers were added
		eb := eventBus.(*InMemoryEventBus)
		eb.mu.RLock()
		handlers, exists := eb.handlers[EventCombatStarted]
		eb.mu.RUnlock()
		
		require.True(t, exists)
		require.Len(t, handlers, 3)
	})

	t.Run("different event types", func(t *testing.T) {
		logger := log.New(bytes.NewBuffer(nil), "", 0)
		eventBus := NewEventBus(logger)
		
		combatHandler := func(ctx context.Context, event Event) error { return nil }
		levelHandler := func(ctx context.Context, event Event) error { return nil }
		questHandler := func(ctx context.Context, event Event) error { return nil }
		
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, combatHandler))
		require.NoError(t, eventBus.Subscribe(EventCharacterLeveled, levelHandler))
		require.NoError(t, eventBus.Subscribe(EventQuestCompleted, questHandler))
		
		// Verify handlers were added to correct event types
		eb := eventBus.(*InMemoryEventBus)
		eb.mu.RLock()
		defer eb.mu.RUnlock()
		
		require.Len(t, eb.handlers[EventCombatStarted], 1)
		require.Len(t, eb.handlers[EventCharacterLeveled], 1)
		require.Len(t, eb.handlers[EventQuestCompleted], 1)
	})

	t.Run("nil handler error", func(t *testing.T) {
		logger := log.New(bytes.NewBuffer(nil), "", 0)
		eventBus := NewEventBus(logger)
		
		err := eventBus.Subscribe(EventCombatStarted, nil)
		
		require.Error(t, err)
		require.Contains(t, err.Error(), "handler cannot be nil")
		
		// Verify no handler was added
		eb := eventBus.(*InMemoryEventBus)
		eb.mu.RLock()
		handlers, exists := eb.handlers[EventCombatStarted]
		eb.mu.RUnlock()
		
		require.False(t, exists)
		require.Len(t, handlers, 0)
	})
}

func TestInMemoryEventBus_Publish(t *testing.T) {
	t.Run("successful event publishing", func(t *testing.T) {
		logger := log.New(bytes.NewBuffer(nil), "", 0)
		eventBus := NewEventBus(logger)
		
		ctx := testutil.TestContext()
		receivedEvents := make(chan Event, 1)
		
		handler := func(ctx context.Context, event Event) error {
			receivedEvents <- event
			return nil
		}
		
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, handler))
		
		event := CombatStartedEvent{
			BaseEvent: BaseEvent{
				EventType: EventCombatStarted,
				EventTime: time.Now(),
			},
			CombatID:  "combat-123",
			SessionID: "session-456",
		}
		
		err := eventBus.Publish(ctx, event)
		require.NoError(t, err)
		
		// Wait for handler to process
		select {
		case received := <-receivedEvents:
			require.Equal(t, EventCombatStarted, received.Type())
			data, ok := received.(*CombatStartedEvent)
			require.True(t, ok)
			require.Equal(t, "combat-123", data.CombatID)
			require.Equal(t, "session-456", data.SessionID)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Handler did not receive event")
		}
	})

	t.Run("multiple handlers receive event", func(t *testing.T) {
		logger := log.New(bytes.NewBuffer(nil), "", 0)
		eventBus := NewEventBus(logger)
		
		ctx := testutil.TestContext()
		var count int32
		wg := sync.WaitGroup{}
		wg.Add(3)
		
		handler := func(ctx context.Context, event Event) error {
			atomic.AddInt32(&count, 1)
			wg.Done()
			return nil
		}
		
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, handler))
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, handler))
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, handler))
		
		event := NewEvent(EventCombatStarted, map[string]string{"test": "data"})
		
		err := eventBus.Publish(ctx, event)
		require.NoError(t, err)
		
		// Wait for all handlers
		done := make(chan bool, 1)
		go func() {
			wg.Wait()
			done <- true
		}()
		
		select {
		case <-done:
			require.Equal(t, int32(3), atomic.LoadInt32(&count))
		case <-time.After(200 * time.Millisecond):
			t.Fatal("Not all handlers received the event")
		}
	})

	t.Run("no handlers for event type", func(t *testing.T) {
		logger := log.New(bytes.NewBuffer(nil), "", 0)
		eventBus := NewEventBus(logger)
		
		ctx := testutil.TestContext()
		
		// Subscribe to different event type
		handler := func(ctx context.Context, event Event) error {
			t.Fatal("Handler should not be called")
			return nil
		}
		require.NoError(t, eventBus.Subscribe(EventCombatEnded, handler))
		
		// Publish event with no handlers
		event := NewEvent(EventCombatStarted, nil)
		
		err := eventBus.Publish(ctx, event)
		require.NoError(t, err)
		
		// Give time to ensure handler is not called
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("handler error is logged", func(t *testing.T) {
		var logBuffer bytes.Buffer
		logger := log.New(&logBuffer, "", 0)
		eventBus := NewEventBus(logger)
		
		ctx := testutil.TestContext()
		handlerCalled := make(chan bool, 1)
		
		handler := func(ctx context.Context, event Event) error {
			handlerCalled <- true
			return errors.New("handler error")
		}
		
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, handler))
		
		event := NewEvent(EventCombatStarted, nil)
		err := eventBus.Publish(ctx, event)
		require.NoError(t, err)
		
		// Wait for handler
		select {
		case <-handlerCalled:
			// Give logger time to write
			time.Sleep(10 * time.Millisecond)
			logOutput := logBuffer.String()
			require.Contains(t, logOutput, "Event handler error")
			require.Contains(t, logOutput, "handler error")
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Handler was not called")
		}
	})

	t.Run("handler panic is recovered", func(t *testing.T) {
		var logBuffer bytes.Buffer
		logger := log.New(&logBuffer, "", 0)
		eventBus := NewEventBus(logger)
		
		ctx := testutil.TestContext()
		handlerCalled := make(chan bool, 1)
		
		panicHandler := func(ctx context.Context, event Event) error {
			handlerCalled <- true
			panic("test panic")
		}
		
		normalHandler := func(ctx context.Context, event Event) error {
			handlerCalled <- true
			return nil
		}
		
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, panicHandler))
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, normalHandler))
		
		event := NewEvent(EventCombatStarted, nil)
		err := eventBus.Publish(ctx, event)
		require.NoError(t, err)
		
		// Both handlers should be called
		for i := 0; i < 2; i++ {
			select {
			case <-handlerCalled:
				// Good
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Not all handlers were called")
			}
		}
		
		// Give logger time to write
		time.Sleep(10 * time.Millisecond)
		logOutput := logBuffer.String()
		require.Contains(t, logOutput, "Event handler panic")
		require.Contains(t, logOutput, "test panic")
	})

	t.Run("concurrent publish and subscribe", func(t *testing.T) {
		logger := log.New(bytes.NewBuffer(nil), "", 0)
		eventBus := NewEventBus(logger)
		
		ctx := testutil.TestContext()
		const numGoroutines = 10
		const numEvents = 100
		
		receivedCount := int32(0)
		wg := sync.WaitGroup{}
		
		// Concurrent subscribers
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				handler := func(ctx context.Context, event Event) error {
					atomic.AddInt32(&receivedCount, 1)
					return nil
				}
				
				eventType := fmt.Sprintf("event.type.%d", id%3)
				if err := eventBus.Subscribe(eventType, handler); err != nil {
					t.Errorf("Subscribe error: %v", err)
				}
			}(i)
		}
		
		// Wait for subscribers
		wg.Wait()
		
		// Concurrent publishers
		wg = sync.WaitGroup{}
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				for j := 0; j < numEvents/numGoroutines; j++ {
					eventType := fmt.Sprintf("event.type.%d", j%3)
					event := NewEvent(eventType, map[string]int{"id": id, "seq": j})
					
					if err := eventBus.Publish(ctx, event); err != nil {
						t.Errorf("Publish error: %v", err)
					}
				}
			}(i)
		}
		
		// Wait for publishers
		wg.Wait()
		
		// Give handlers time to process
		time.Sleep(100 * time.Millisecond)
		
		// Verify events were received
		// Each event type (0, 1, 2) gets ~33 events, with ~3-4 handlers each
		minExpected := int32(numEvents * 3) // At least 3 handlers per event type
		require.GreaterOrEqual(t, atomic.LoadInt32(&receivedCount), minExpected)
	})
}

func TestNewEvent(t *testing.T) {
	t.Run("creates event with correct fields", func(t *testing.T) {
		eventType := "test.event"
		data := map[string]interface{}{
			"field1": "value1",
			"field2": 42,
		}
		
		beforeCreate := time.Now()
		event := NewEvent(eventType, data)
		afterCreate := time.Now()
		
		require.Equal(t, eventType, event.Type())
		require.Equal(t, data, event.Data())
		
		// Verify timestamp is within expected range
		timestamp := event.Timestamp()
		require.True(t, !timestamp.Before(beforeCreate))
		require.True(t, !timestamp.After(afterCreate))
	})

	t.Run("nil data is handled", func(t *testing.T) {
		event := NewEvent("test.event", nil)
		
		require.Equal(t, "test.event", event.Type())
		require.Nil(t, event.Data())
		require.NotZero(t, event.Timestamp())
	})
}

func TestDomainEvents(t *testing.T) {
	t.Run("CombatStartedEvent", func(t *testing.T) {
		event := CombatStartedEvent{
			BaseEvent: BaseEvent{
				EventType: EventCombatStarted,
				EventTime: time.Now(),
			},
			CombatID:  "combat-123",
			SessionID: "session-456",
		}
		
		require.Equal(t, EventCombatStarted, event.Type())
		require.Equal(t, "combat-123", event.CombatID)
		require.Equal(t, "session-456", event.SessionID)
	})

	t.Run("CombatEndedEvent", func(t *testing.T) {
		event := CombatEndedEvent{
			BaseEvent: BaseEvent{
				EventType: EventCombatEnded,
				EventTime: time.Now(),
			},
			CombatID:  "combat-123",
			SessionID: "session-456",
			Victory:   true,
		}
		
		require.Equal(t, EventCombatEnded, event.Type())
		require.Equal(t, "combat-123", event.CombatID)
		require.Equal(t, "session-456", event.SessionID)
		require.True(t, event.Victory)
	})

	t.Run("CharacterLeveledEvent", func(t *testing.T) {
		event := CharacterLeveledEvent{
			BaseEvent: BaseEvent{
				EventType: EventCharacterLeveled,
				EventTime: time.Now(),
			},
			CharacterID: "char-123",
			OldLevel:    5,
			NewLevel:    6,
		}
		
		require.Equal(t, EventCharacterLeveled, event.Type())
		require.Equal(t, "char-123", event.CharacterID)
		require.Equal(t, 5, event.OldLevel)
		require.Equal(t, 6, event.NewLevel)
	})

	t.Run("FactionRelationChangedEvent", func(t *testing.T) {
		event := FactionRelationChangedEvent{
			BaseEvent: BaseEvent{
				EventType: EventFactionRelationChanged,
				EventTime: time.Now(),
			},
			Faction1ID:  "faction-1",
			Faction2ID:  "faction-2",
			OldRelation: 0.5,
			NewRelation: 0.8,
		}
		
		require.Equal(t, EventFactionRelationChanged, event.Type())
		require.Equal(t, "faction-1", event.Faction1ID)
		require.Equal(t, "faction-2", event.Faction2ID)
		require.Equal(t, 0.5, event.OldRelation)
		require.Equal(t, 0.8, event.NewRelation)
	})
}

func TestEventBusIntegration(t *testing.T) {
	t.Run("complete event flow", func(t *testing.T) {
		logger := log.New(bytes.NewBuffer(nil), "", 0)
		eventBus := NewEventBus(logger)
		ctx := testutil.TestContext()
		
		// Track events received by each handler
		combatEvents := make(chan *CombatStartedEvent, 1)
		analyticsEvents := make(chan Event, 1)
		auditEvents := make(chan Event, 1)
		
		// Combat handler - processes combat-specific logic
		combatHandler := func(ctx context.Context, event Event) error {
			if combat, ok := event.(*CombatStartedEvent); ok {
				combatEvents <- combat
			}
			return nil
		}
		
		// Analytics handler - tracks all events
		analyticsHandler := func(ctx context.Context, event Event) error {
			analyticsEvents <- event
			return nil
		}
		
		// Audit handler - logs important events
		auditHandler := func(ctx context.Context, event Event) error {
			if strings.Contains(event.Type(), "combat") {
				auditEvents <- event
			}
			return nil
		}
		
		// Subscribe handlers
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, combatHandler))
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, analyticsHandler))
		require.NoError(t, eventBus.Subscribe(EventCombatStarted, auditHandler))
		
		// Publish event
		combatEvent := &CombatStartedEvent{
			BaseEvent: BaseEvent{
				EventType: EventCombatStarted,
				EventTime: time.Now(),
			},
			CombatID:  "combat-xyz",
			SessionID: "session-abc",
		}
		
		require.NoError(t, eventBus.Publish(ctx, combatEvent))
		
		// Verify all handlers received the event
		timeout := time.After(200 * time.Millisecond)
		
		select {
		case received := <-combatEvents:
			require.Equal(t, "combat-xyz", received.CombatID)
		case <-timeout:
			t.Fatal("Combat handler did not receive event")
		}
		
		select {
		case received := <-analyticsEvents:
			require.Equal(t, EventCombatStarted, received.Type())
		case <-timeout:
			t.Fatal("Analytics handler did not receive event")
		}
		
		select {
		case received := <-auditEvents:
			require.Equal(t, EventCombatStarted, received.Type())
		case <-timeout:
			t.Fatal("Audit handler did not receive event")
		}
	})
}

// Benchmark tests
func BenchmarkEventBus_Publish(b *testing.B) {
	logger := log.New(bytes.NewBuffer(nil), "", 0)
	eventBus := NewEventBus(logger)
	ctx := context.Background()
	
	// Add a simple handler
	handler := func(ctx context.Context, event Event) error {
		return nil
	}
	eventBus.Subscribe(EventCombatStarted, handler)
	
	event := NewEvent(EventCombatStarted, map[string]string{"test": "data"})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eventBus.Publish(ctx, event)
	}
}

func BenchmarkEventBus_Subscribe(b *testing.B) {
	logger := log.New(bytes.NewBuffer(nil), "", 0)
	eventBus := NewEventBus(logger)
	
	handler := func(ctx context.Context, event Event) error {
		return nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eventType := fmt.Sprintf("event.type.%d", i)
		eventBus.Subscribe(eventType, handler)
	}
}