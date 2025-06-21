package services

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Test constants
const (
	testDMID            = "dm-123"
	testPlayerID        = testPlayerID
	testCampaignName    = "Test Campaign"
	testSessionNotFound = testSessionNotFound
	testNonExistentID   = testNonExistentID
)

func TestGameService_CreateSession(t *testing.T) {
	t.Run("successful session creation", func(t *testing.T) {
		service := NewGameService()

		session := &models.GameSession{
			DMID:        testDMID,
			Name:        "The Lost Mines",
			Description: "An adventure into the lost mines of Phandelver",
		}

		beforeCreate := time.Now()
		result, err := service.CreateSession(session)
		afterCreate := time.Now()

		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify ID was generated
		require.NotEmpty(t, result.ID)

		// Verify status was set
		require.Equal(t, models.GameStatusActive, result.Status)

		// Verify timestamp
		require.True(t, !result.CreatedAt.Before(beforeCreate))
		require.True(t, !result.CreatedAt.After(afterCreate))

		// Verify session was stored
		stored, exists := service.sessions[result.ID]
		require.True(t, exists)
		require.Equal(t, result, stored)

		// Verify events slice was initialized
		events, exists := service.events[result.ID]
		require.True(t, exists)
		require.NotNil(t, events)
		require.Len(t, events, 0)
	})

	t.Run("multiple sessions with unique IDs", func(t *testing.T) {
		service := NewGameService()

		session1 := &models.GameSession{
			DMID: "dm-1",
			Name: "Campaign 1",
		}

		session2 := &models.GameSession{
			DMID: "dm-2",
			Name: "Campaign 2",
		}

		result1, err1 := service.CreateSession(session1)
		result2, err2 := service.CreateSession(session2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		// Verify unique IDs
		require.NotEqual(t, result1.ID, result2.ID)

		// Verify both are stored
		require.Len(t, service.sessions, 2)
		require.Contains(t, service.sessions, result1.ID)
		require.Contains(t, service.sessions, result2.ID)
	})

	t.Run("preserves original session fields", func(t *testing.T) {
		service := NewGameService()

		session := &models.GameSession{
			DMID:        "dm-456",
			Name:        "Curse of Strahd",
			Description: "Gothic horror in Barovia",
			State: map[string]interface{}{
				"currentLocation": "Village of Barovia",
				"partyLevel":      3,
			},
		}

		result, err := service.CreateSession(session)

		require.NoError(t, err)
		require.Equal(t, "dm-456", result.DMID)
		require.Equal(t, "Curse of Strahd", result.Name)
		require.Equal(t, "Gothic horror in Barovia", result.Description)
		require.Equal(t, session.State, result.State)
	})
}

func TestGameService_GetSessionByID(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		service := NewGameService()

		// Create a session first
		session := &models.GameSession{
			DMID: testDMID,
			Name: testCampaignName,
		}
		created, _ := service.CreateSession(session)

		// Retrieve it
		retrieved, err := service.GetSessionByID(created.ID)

		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, created, retrieved)
	})

	t.Run(testSessionNotFound, func(t *testing.T) {
		service := NewGameService()

		retrieved, err := service.GetSessionByID("non-existent-id")

		require.Error(t, err)
		require.Nil(t, retrieved)
		require.Contains(t, err.Error(), testSessionNotFound)
	})

	t.Run("empty ID", func(t *testing.T) {
		service := NewGameService()

		retrieved, err := service.GetSessionByID("")

		require.Error(t, err)
		require.Nil(t, retrieved)
		require.Contains(t, err.Error(), testSessionNotFound)
	})
}

func TestGameService_AddPlayerToSession(t *testing.T) {
	t.Run("successful player addition", func(t *testing.T) {
		service := NewGameService()

		// Create a session first
		session := &models.GameSession{
			DMID: testDMID,
			Name: testCampaignName,
		}
		created, _ := service.CreateSession(session)

		// Add player
		player := &models.Player{
			ID:          "player-456",
			Name:        "Aragorn",
			CharacterID: "char-789",
			IsOnline:    true,
		}

		beforeAdd := time.Now()
		err := service.AddPlayerToSession(created.ID, player.ID, player)
		afterAdd := time.Now()

		require.NoError(t, err)

		// Verify JoinedAt was set
		require.True(t, !player.JoinedAt.Before(beforeAdd))
		require.True(t, !player.JoinedAt.After(afterAdd))

		// Note: Current implementation doesn't actually store players
		// This is marked as TODO in the code
	})

	t.Run(testSessionNotFound, func(t *testing.T) {
		service := NewGameService()

		player := &models.Player{
			ID:   testPlayerID,
			Name: "Test Player",
		}

		err := service.AddPlayerToSession(testNonExistentID, player.ID, player)

		require.Error(t, err)
		require.Contains(t, err.Error(), testSessionNotFound)
	})

	t.Run("empty session ID", func(t *testing.T) {
		service := NewGameService()

		player := &models.Player{
			ID:   testPlayerID,
			Name: "Test Player",
		}

		err := service.AddPlayerToSession("", player.ID, player)

		require.Error(t, err)
		require.Contains(t, err.Error(), testSessionNotFound)
	})
}

func TestGameService_RecordGameEvent(t *testing.T) {
	t.Run("successful event recording", func(t *testing.T) {
		service := NewGameService()

		// Create a session first
		session := &models.GameSession{
			DMID: testDMID,
			Name: testCampaignName,
		}
		created, _ := service.CreateSession(session)

		// Record event
		event := &models.GameEvent{
			SessionID: created.ID,
			Type:      "roll",
			PlayerID:  "player-456",
			Data: map[string]interface{}{
				"dice":   "1d20",
				"result": 15,
			},
		}

		beforeRecord := time.Now()
		err := service.RecordGameEvent(event)
		afterRecord := time.Now()

		require.NoError(t, err)

		// Verify ID was generated
		require.NotEmpty(t, event.ID)

		// Verify timestamp
		require.True(t, !event.Timestamp.Before(beforeRecord))
		require.True(t, !event.Timestamp.After(afterRecord))

		// Verify event was stored
		events, exists := service.events[created.ID]
		require.True(t, exists)
		require.Len(t, events, 1)
		require.Equal(t, event, events[0])
	})

	t.Run("multiple events in order", func(t *testing.T) {
		service := NewGameService()

		// Create a session
		session := &models.GameSession{
			DMID: testDMID,
			Name: testCampaignName,
		}
		created, _ := service.CreateSession(session)

		// Record multiple events
		event1 := &models.GameEvent{
			SessionID: created.ID,
			Type:      "roll",
			PlayerID:  "player-1",
			Data:      map[string]interface{}{"result": 10},
		}

		event2 := &models.GameEvent{
			SessionID: created.ID,
			Type:      "message",
			PlayerID:  "player-2",
			Data:      map[string]interface{}{"text": "I attack!"},
		}

		event3 := &models.GameEvent{
			SessionID: created.ID,
			Type:      "combat",
			PlayerID:  "dm-123",
			Data:      map[string]interface{}{"action": "initiative"},
		}

		require.NoError(t, service.RecordGameEvent(event1))
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		require.NoError(t, service.RecordGameEvent(event2))
		time.Sleep(10 * time.Millisecond)
		require.NoError(t, service.RecordGameEvent(event3))

		// Verify all events were stored in order
		events := service.events[created.ID]
		require.Len(t, events, 3)
		require.Equal(t, "roll", events[0].Type)
		require.Equal(t, "message", events[1].Type)
		require.Equal(t, "combat", events[2].Type)

		// Verify timestamps are in order
		require.True(t, events[0].Timestamp.Before(events[1].Timestamp))
		require.True(t, events[1].Timestamp.Before(events[2].Timestamp))

		// Verify each event has unique ID
		require.NotEqual(t, events[0].ID, events[1].ID)
		require.NotEqual(t, events[1].ID, events[2].ID)
		require.NotEqual(t, events[0].ID, events[2].ID)
	})

	t.Run(testSessionNotFound, func(t *testing.T) {
		service := NewGameService()

		event := &models.GameEvent{
			SessionID: testNonExistentID,
			Type:      "roll",
			PlayerID:  testPlayerID,
		}

		err := service.RecordGameEvent(event)

		require.Error(t, err)
		require.Contains(t, err.Error(), testSessionNotFound)

		// Verify event ID and timestamp were not set
		require.Empty(t, event.ID)
		require.True(t, event.Timestamp.IsZero())
	})

	t.Run("empty session ID", func(t *testing.T) {
		service := NewGameService()

		event := &models.GameEvent{
			SessionID: "",
			Type:      "roll",
			PlayerID:  testPlayerID,
		}

		err := service.RecordGameEvent(event)

		require.Error(t, err)
		require.Contains(t, err.Error(), testSessionNotFound)
	})
}

func TestGameService_GetSessionEvents(t *testing.T) {
	t.Run("successful retrieval of events", func(t *testing.T) {
		service := NewGameService()

		// Create session and add events
		session := &models.GameSession{
			DMID: testDMID,
			Name: testCampaignName,
		}
		created, _ := service.CreateSession(session)

		// Record several events
		for i := 0; i < 5; i++ {
			event := &models.GameEvent{
				SessionID: created.ID,
				Type:      "test",
				PlayerID:  testPlayerID,
				Data:      map[string]interface{}{"index": i},
			}
			_ = service.RecordGameEvent(event)
		}

		// Retrieve events
		events, err := service.GetSessionEvents(created.ID)

		require.NoError(t, err)
		require.NotNil(t, events)
		require.Len(t, events, 5)

		// Verify order and data
		for i, event := range events {
			require.Equal(t, "test", event.Type)
			require.Equal(t, i, event.Data["index"])
		}
	})

	t.Run("empty events list", func(t *testing.T) {
		service := NewGameService()

		// Create session without events
		session := &models.GameSession{
			DMID: testDMID,
			Name: testCampaignName,
		}
		created, _ := service.CreateSession(session)

		// Retrieve events
		events, err := service.GetSessionEvents(created.ID)

		require.NoError(t, err)
		require.NotNil(t, events)
		require.Len(t, events, 0)
	})

	t.Run(testSessionNotFound, func(t *testing.T) {
		service := NewGameService()

		events, err := service.GetSessionEvents(testNonExistentID)

		require.Error(t, err)
		require.Nil(t, events)
		require.Contains(t, err.Error(), testSessionNotFound)
	})
}

// createSessionConcurrently is a helper for concurrent session creation
func createSessionConcurrently(service *GameService, id int) (string, error) {
	session := &models.GameSession{
		DMID: "dm-" + string(rune(id)),
		Name: "Campaign " + string(rune(id)),
	}
	result, err := service.CreateSession(session)
	if err != nil {
		return "", err
	}
	return result.ID, nil
}

// testConcurrentSessionCreation tests concurrent session creation
func testConcurrentSessionCreation(t *testing.T, service *GameService) {
	const numGoroutines = 10
	sessionIDs := make(chan string, numGoroutines)
	errorsChan := make(chan error, numGoroutines)

	wg := sync.WaitGroup{}
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sessionID, err := createSessionConcurrently(service, id)
			if err != nil {
				errorsChan <- err
			} else {
				sessionIDs <- sessionID
			}
		}(i)
	}

	wg.Wait()
	close(sessionIDs)
	close(errorsChan)

	// Check for errors
	for err := range errorsChan {
		t.Fatalf("Unexpected error during concurrent creation: %v", err)
	}

	// Verify all sessions were created
	createdIDs := make(map[string]bool)
	for id := range sessionIDs {
		if createdIDs[id] {
			t.Fatal("Duplicate session ID generated")
		}
		createdIDs[id] = true
	}

	require.Len(t, createdIDs, numGoroutines)
	require.Len(t, service.sessions, numGoroutines)
}

// recordEventConcurrently is a helper for concurrent event recording
func recordEventConcurrently(service *GameService, sessionID string, index int) error {
	event := &models.GameEvent{
		SessionID: sessionID,
		Type:      "concurrent",
		PlayerID:  "player-" + string(rune(index%10)),
		Data:      map[string]interface{}{"index": index},
	}
	return service.RecordGameEvent(event)
}

// testConcurrentEventRecording tests concurrent event recording
func testConcurrentEventRecording(t *testing.T, service *GameService) {
	// Create a session
	session := &models.GameSession{
		DMID: testDMID,
		Name: testCampaignName,
	}
	created, _ := service.CreateSession(session)

	const numEvents = 100
	errorsChan := make(chan error, numEvents)

	wg := sync.WaitGroup{}
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if err := recordEventConcurrently(service, created.ID, index); err != nil {
				errorsChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	for err := range errorsChan {
		t.Fatalf("Unexpected error during concurrent recording: %v", err)
	}

	// Verify all events were recorded
	events, err := service.GetSessionEvents(created.ID)
	require.NoError(t, err)
	require.Len(t, events, numEvents)

	// Verify each event has unique ID
	verifyUniqueEventIDs(t, events)
}

// verifyUniqueEventIDs verifies that all events have unique IDs
func verifyUniqueEventIDs(t *testing.T, events []*models.GameEvent) {
	eventIDs := make(map[string]bool)
	for _, event := range events {
		if eventIDs[event.ID] {
			t.Fatal("Duplicate event ID found")
		}
		eventIDs[event.ID] = true
	}
}

func TestGameService_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent session creation", func(t *testing.T) {
		service := NewGameService()
		testConcurrentSessionCreation(t, service)
	})

	t.Run("concurrent event recording", func(t *testing.T) {
		service := NewGameService()
		testConcurrentEventRecording(t, service)
	})

	t.Run("concurrent read and write", func(t *testing.T) {
		service := NewGameService()
		testConcurrentReadWrite(t, service)
	})
}

// createMultipleSessions creates multiple test sessions
func createMultipleSessions(service *GameService, count int) []string {
	sessionIDs := make([]string, count)
	for i := 0; i < count; i++ {
		session := &models.GameSession{
			DMID: "dm-" + string(rune(i)),
			Name: "Campaign " + string(rune(i)),
		}
		created, _ := service.CreateSession(session)
		sessionIDs[i] = created.ID
	}
	return sessionIDs
}

// performConcurrentWrite performs a concurrent write operation
func performConcurrentWrite(service *GameService, sessionID string, index int) error {
	event := &models.GameEvent{
		SessionID: sessionID,
		Type:      "write",
		PlayerID:  "writer-" + string(rune(index)),
	}
	return service.RecordGameEvent(event)
}

// performConcurrentRead performs concurrent read operations
func performConcurrentRead(service *GameService, sessionID string) error {
	// Read session
	if _, err := service.GetSessionByID(sessionID); err != nil {
		return err
	}
	// Read events
	if _, err := service.GetSessionEvents(sessionID); err != nil {
		return err
	}
	return nil
}

// testConcurrentReadWrite tests concurrent read and write operations
func testConcurrentReadWrite(t *testing.T, service *GameService) {
	// Create multiple sessions
	sessionIDs := createMultipleSessions(service, 5)

	// Concurrent operations
	errors := performConcurrentOperations(service, sessionIDs)

	// Check for errors
	require.Len(t, errors, 0, "Unexpected errors during concurrent operations")

	// Verify data integrity
	verifyDataIntegrity(t, service, sessionIDs)
}

// performConcurrentOperations runs concurrent read and write operations
func performConcurrentOperations(service *GameService, sessionIDs []string) []error {
	wg := sync.WaitGroup{}
	errors := make([]error, 0)
	var errorsMu sync.Mutex

	// Start writers
	startConcurrentWriters(&wg, service, sessionIDs, &errors, &errorsMu)

	// Start readers
	startConcurrentReaders(&wg, service, sessionIDs, &errors, &errorsMu)

	wg.Wait()
	return errors
}

// startConcurrentWriters starts concurrent write operations
func startConcurrentWriters(wg *sync.WaitGroup, service *GameService, sessionIDs []string, errors *[]error, errorsMu *sync.Mutex) {
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			sessionID := sessionIDs[index%len(sessionIDs)]
			if err := performConcurrentWrite(service, sessionID, index); err != nil {
				errorsMu.Lock()
				*errors = append(*errors, err)
				errorsMu.Unlock()
			}
		}(i)
	}
}

// startConcurrentReaders starts concurrent read operations
func startConcurrentReaders(wg *sync.WaitGroup, service *GameService, sessionIDs []string, errors *[]error, errorsMu *sync.Mutex) {
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			sessionID := sessionIDs[index%len(sessionIDs)]
			if err := performConcurrentRead(service, sessionID); err != nil {
				errorsMu.Lock()
				*errors = append(*errors, err)
				errorsMu.Unlock()
			}
		}(i)
	}
}

// verifyDataIntegrity verifies the integrity of sessions and events after concurrent operations
func verifyDataIntegrity(t *testing.T, service *GameService, sessionIDs []string) {
	for _, sessionID := range sessionIDs {
		session, err := service.GetSessionByID(sessionID)
		require.NoError(t, err)
		require.NotNil(t, session)

		events, err := service.GetSessionEvents(sessionID)
		require.NoError(t, err)
		require.NotNil(t, events)
	}
}

func TestGenerateID(t *testing.T) {
	t.Run("generates unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)

		for i := 0; i < 1000; i++ {
			id := generateID()

			// Verify format (UUID)
			require.Len(t, id, 36)
			require.Contains(t, id, "-")

			// Verify uniqueness
			if ids[id] {
				t.Fatal("Duplicate ID generated")
			}
			ids[id] = true
		}
	})

	t.Run("concurrent ID generation", func(t *testing.T) {
		const numGoroutines = 100
		const idsPerGoroutine = 100

		idsChan := make(chan string, numGoroutines*idsPerGoroutine)

		wg := sync.WaitGroup{}
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < idsPerGoroutine; j++ {
					idsChan <- generateID()
				}
			}()
		}

		wg.Wait()
		close(idsChan)

		// Verify uniqueness
		ids := make(map[string]bool)
		for id := range idsChan {
			if ids[id] {
				t.Fatal("Duplicate ID generated in concurrent scenario")
			}
			ids[id] = true
		}

		require.Len(t, ids, numGoroutines*idsPerGoroutine)
	})
}
