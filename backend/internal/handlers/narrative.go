package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
)

// Error messages
const (
	errUnauthorized = "unauthorized"
	errFailedToGetParticipants = "Failed to get participants"
)

// NarrativeHandlers manages narrative-related HTTP endpoints
type NarrativeHandlers struct {
	narrativeEngine *services.NarrativeEngine
	narrativeRepo   *database.NarrativeRepository
	characterRepo   database.CharacterRepository
	gameRepo        database.GameSessionRepository
}

// NewNarrativeHandlers creates a new narrative handlers instance
func NewNarrativeHandlers(
	narrativeEngine *services.NarrativeEngine,
	narrativeRepo *database.NarrativeRepository,
	characterRepo database.CharacterRepository,
	gameRepo database.GameSessionRepository,
) *NarrativeHandlers {
	return &NarrativeHandlers{
		narrativeEngine: narrativeEngine,
		narrativeRepo:   narrativeRepo,
		characterRepo:   characterRepo,
		gameRepo:        gameRepo,
	}
}

// RegisterRoutes registers all narrative-related routes
func (h *NarrativeHandlers) RegisterRoutes(router *mux.Router, authMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// Profile management
	router.HandleFunc("/api/narrative/profile/{characterId}", authMiddleware(h.GetNarrativeProfile)).Methods("GET")
	router.HandleFunc("/api/narrative/profile", authMiddleware(h.CreateNarrativeProfile)).Methods("POST")
	router.HandleFunc("/api/narrative/profile/{characterId}", authMiddleware(h.UpdateNarrativeProfile)).Methods("PUT")

	// Backstory management
	router.HandleFunc("/api/narrative/backstory/{characterId}", authMiddleware(h.GetBackstoryElements)).Methods("GET")
	router.HandleFunc("/api/narrative/backstory", authMiddleware(h.CreateBackstoryElement)).Methods("POST")

	// Player actions and consequences
	router.HandleFunc("/api/narrative/action", authMiddleware(h.RecordPlayerAction)).Methods("POST")
	router.HandleFunc("/api/narrative/consequences/{sessionId}", authMiddleware(h.GetPendingConsequences)).Methods("GET")
	router.HandleFunc("/api/narrative/consequences/{consequenceId}/trigger", authMiddleware(h.TriggerConsequence)).Methods("POST")

	// World events and perspectives
	router.HandleFunc("/api/narrative/event", authMiddleware(h.CreateWorldEvent)).Methods("POST")
	router.HandleFunc("/api/narrative/event/{eventId}", authMiddleware(h.GetWorldEvent)).Methods("GET")
	router.HandleFunc("/api/narrative/event/{eventId}/perspectives", authMiddleware(h.GetEventPerspectives)).Methods("GET")
	router.HandleFunc("/api/narrative/event/{eventId}/personalize/{characterId}", authMiddleware(h.PersonalizeEvent)).Methods("POST")

	// Narrative generation
	router.HandleFunc("/api/narrative/generate/story", authMiddleware(h.GeneratePersonalizedStory)).Methods("POST")
	router.HandleFunc("/api/narrative/generate/perspectives", authMiddleware(h.GenerateMultiplePerspectives)).Methods("POST")

	// Memory management
	router.HandleFunc("/api/narrative/memory/{characterId}", authMiddleware(h.GetCharacterMemories)).Methods("GET")
	router.HandleFunc("/api/narrative/memory", authMiddleware(h.CreateMemory)).Methods("POST")

	// Narrative threads
	router.HandleFunc("/api/narrative/threads", authMiddleware(h.GetActiveThreads)).Methods("GET")
	router.HandleFunc("/api/narrative/threads", authMiddleware(h.CreateThread)).Methods("POST")
}

// GetNarrativeProfile retrieves a character's narrative profile
func (h *NarrativeHandlers) GetNarrativeProfile(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	// Verify character ownership
	character, err := h.characterRepo.GetByID(r.Context(), characterID)
	if err != nil {
		http.Error(w, constants.ErrCharacterNotFound, http.StatusNotFound)
		return
	}

	if character.UserID != claims.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	profile, err := h.narrativeRepo.GetNarrativeProfile(characterID)
	if err != nil {
		// Create default profile if none exists
		profile = &models.NarrativeProfile{
			UserID:      claims.UserID,
			CharacterID: characterID,
			Preferences: models.StoryPreferences{
				Themes:           []string{},
				Tone:             []string{},
				Complexity:       3,
				MoralAlignment:   "neutral",
				PacingPreference: "moderate",
				CombatNarrative:  0.5,
			},
			DecisionHistory: []models.DecisionRecord{},
			PlayStyle:       "balanced",
			Analytics:       make(map[string]interface{}),
		}

		if err := h.narrativeRepo.CreateNarrativeProfile(profile); err != nil {
			http.Error(w, "Failed to create profile", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// CreateNarrativeProfile creates a new narrative profile
func (h *NarrativeHandlers) CreateNarrativeProfile(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())

	var profile models.NarrativeProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, ErrInvalidRequestBody, http.StatusBadRequest)
		return
	}

	// Verify character ownership
	character, err := h.characterRepo.GetByID(r.Context(), profile.CharacterID)
	if err != nil {
		http.Error(w, constants.ErrCharacterNotFound, http.StatusNotFound)
		return
	}

	if character.UserID != claims.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	profile.UserID = claims.UserID
	if err := h.narrativeRepo.CreateNarrativeProfile(&profile); err != nil {
		http.Error(w, "Failed to create profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// UpdateNarrativeProfile updates an existing narrative profile
func (h *NarrativeHandlers) UpdateNarrativeProfile(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	// Verify character ownership
	character, err := h.characterRepo.GetByID(r.Context(), characterID)
	if err != nil {
		http.Error(w, constants.ErrCharacterNotFound, http.StatusNotFound)
		return
	}

	if character.UserID != claims.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, ErrInvalidRequestBody, http.StatusBadRequest)
		return
	}

	profile, err := h.narrativeRepo.GetNarrativeProfile(characterID)
	if err != nil {
		http.Error(w, constants.ErrProfileNotFound, http.StatusNotFound)
		return
	}

	// Apply updates
	h.applyProfileUpdates(profile, updates)

	if err := h.narrativeRepo.UpdateNarrativeProfile(profile); err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// GetBackstoryElements retrieves backstory elements for a character
func (h *NarrativeHandlers) GetBackstoryElements(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	// Verify character ownership
	character, err := h.characterRepo.GetByID(r.Context(), characterID)
	if err != nil {
		http.Error(w, constants.ErrCharacterNotFound, http.StatusNotFound)
		return
	}

	if character.UserID != claims.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	elements, err := h.narrativeRepo.GetBackstoryElements(characterID)
	if err != nil {
		http.Error(w, "Failed to retrieve backstory elements", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(elements); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// CreateBackstoryElement creates a new backstory element
func (h *NarrativeHandlers) CreateBackstoryElement(w http.ResponseWriter, r *http.Request) {
	HandleCharacterOwnedCreation(w, r, h.characterRepo,
		func(element *models.BackstoryElement) string { return element.CharacterID },
		func(element *models.BackstoryElement) error { return h.narrativeRepo.CreateBackstoryElement(element) })
}

// RecordPlayerAction records a significant player action
func (h *NarrativeHandlers) RecordPlayerAction(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())

	var action models.PlayerAction
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		http.Error(w, ErrInvalidRequestBody, http.StatusBadRequest)
		return
	}

	// Verify permissions
	if err := h.verifyActionPermissions(r.Context(), claims, &action); err != nil {
		h.handleActionPermissionError(w, err)
		return
	}

	// Record the action
	if err := h.narrativeRepo.CreatePlayerAction(&action); err != nil {
		http.Error(w, "Failed to record action", http.StatusInternalServerError)
		return
	}

	// Calculate consequences asynchronously
	go h.processActionConsequences(r.Context(), action)

	// Update narrative profile with decision
	h.updateNarrativeProfileWithAction(r.Context(), action)

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(action); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// GetPendingConsequences retrieves consequences ready to trigger
func (h *NarrativeHandlers) GetPendingConsequences(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Verify session participation
	participants, err := h.gameRepo.GetParticipants(r.Context(), sessionID)
	if err != nil {
		http.Error(w, errFailedToGetParticipants, http.StatusInternalServerError)
		return
	}

	isParticipant := false
	for _, p := range participants {
		if p.UserID == claims.UserID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		http.Error(w, "Not a participant in this session", http.StatusForbidden)
		return
	}

	consequences, err := h.narrativeRepo.GetPendingConsequences(sessionID, time.Now())
	if err != nil {
		http.Error(w, "Failed to retrieve consequences", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(consequences); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// TriggerConsequence triggers a specific consequence
func (h *NarrativeHandlers) TriggerConsequence(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	vars := mux.Vars(r)
	consequenceID := vars["consequenceId"]

	var triggerData struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&triggerData); err != nil {
		http.Error(w, ErrInvalidRequestBody, http.StatusBadRequest)
		return
	}

	// Verify DM role
	participants, err := h.gameRepo.GetParticipants(r.Context(), triggerData.SessionID)
	if err != nil {
		http.Error(w, errFailedToGetParticipants, http.StatusInternalServerError)
		return
	}

	isDM := false
	for _, p := range participants {
		if p.UserID == claims.UserID && p.Role == models.ParticipantRoleDM {
			isDM = true
			break
		}
	}

	if !isDM {
		http.Error(w, "Only DM can trigger consequences", http.StatusForbidden)
		return
	}

	now := time.Now()
	if err := h.narrativeRepo.UpdateConsequenceStatus(consequenceID, "triggered", &now); err != nil {
		http.Error(w, "Failed to trigger consequence", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "triggered"}); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// CreateWorldEvent creates a new world event
func (h *NarrativeHandlers) CreateWorldEvent(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())

	var event models.NarrativeEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, ErrInvalidRequestBody, http.StatusBadRequest)
		return
	}

	// Verify DM role for the session
	if err := h.verifyDMRoleForEvent(r.Context(), claims, &event); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if err := h.narrativeRepo.CreateWorldEvent(&event); err != nil {
		http.Error(w, "Failed to create world event", http.StatusInternalServerError)
		return
	}

	// Generate perspectives asynchronously
	go h.generateEventPerspectives(r.Context(), &event)

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// GetWorldEvent retrieves a specific world event
func (h *NarrativeHandlers) GetWorldEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["eventId"]

	event, err := h.narrativeRepo.GetWorldEvent(eventID)
	if err != nil {
		http.Error(w, constants.ErrEventNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// GetEventPerspectives retrieves all perspectives for an event
func (h *NarrativeHandlers) GetEventPerspectives(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["eventId"]

	perspectives, err := h.narrativeRepo.GetEventPerspectives(eventID)
	if err != nil {
		http.Error(w, "Failed to retrieve perspectives", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(perspectives); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// PersonalizeEvent creates a personalized version of an event for a character
func (h *NarrativeHandlers) PersonalizeEvent(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	vars := mux.Vars(r)
	eventID := vars["eventId"]
	characterID := vars["characterId"]

	// Verify ownership and gather data
	data, err := h.gatherPersonalizationData(r.Context(), claims, characterID, eventID)
	if err != nil {
		h.handlePersonalizationError(w, err)
		return
	}

	// Generate and save narrative
	narrative, err := h.createPersonalizedNarrative(r.Context(), data)
	if err != nil {
		http.Error(w, "Failed to create narrative", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(narrative); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// GeneratePersonalizedStory generates a new personalized story
func (h *NarrativeHandlers) GeneratePersonalizedStory(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())

	var request struct {
		CharacterID string                 `json:"character_id"`
		EventType   string                 `json:"event_type"`
		Context     map[string]interface{} `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, ErrInvalidRequestBody, http.StatusBadRequest)
		return
	}

	// Verify character ownership
	character, err := h.characterRepo.GetByID(r.Context(), request.CharacterID)
	if err != nil {
		http.Error(w, constants.ErrCharacterNotFound, http.StatusNotFound)
		return
	}

	if character.UserID != claims.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Create base event
	event := models.NarrativeEvent{
		Type:        request.EventType,
		Name:        "Generated Story Event",
		Description: "A personalized story event",
		Status:      "active",
		Metadata:    request.Context,
	}

	// Get profile and backstory
	profile, err := h.narrativeRepo.GetNarrativeProfile(request.CharacterID)
	if err != nil {
		http.Error(w, constants.ErrProfileNotFound, http.StatusNotFound)
		return
	}

	backstory, err := h.narrativeRepo.GetBackstoryElements(request.CharacterID)
	if err != nil {
		backstory = []models.BackstoryElement{}
	}

	// Generate narrative
	narrative, err := h.narrativeEngine.GeneratePersonalizedNarrative(r.Context(), &event, profile, backstory)
	if err != nil {
		http.Error(w, "Failed to generate story", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(narrative); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// GenerateMultiplePerspectives generates perspectives for an event
func (h *NarrativeHandlers) GenerateMultiplePerspectives(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())

	var request struct {
		EventID   string                     `json:"event_id"`
		Sources   []models.PerspectiveSource `json:"sources"`
		SessionID string                     `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, ErrInvalidRequestBody, http.StatusBadRequest)
		return
	}

	// Verify DM role
	participants, err := h.gameRepo.GetParticipants(r.Context(), request.SessionID)
	if err != nil {
		http.Error(w, errFailedToGetParticipants, http.StatusInternalServerError)
		return
	}

	isDM := false
	for _, p := range participants {
		if p.UserID == claims.UserID && p.Role == models.ParticipantRoleDM {
			isDM = true
			break
		}
	}

	if !isDM {
		http.Error(w, "Only DM can generate perspectives", http.StatusForbidden)
		return
	}

	// Get event
	event, err := h.narrativeRepo.GetWorldEvent(request.EventID)
	if err != nil {
		http.Error(w, constants.ErrEventNotFound, http.StatusNotFound)
		return
	}

	// Generate perspectives
	perspectives, err := h.narrativeEngine.PerspectiveGen.GenerateMultiplePerspectives(r.Context(), event, request.Sources)
	if err != nil {
		http.Error(w, "Failed to generate perspectives", http.StatusInternalServerError)
		return
	}

	// Save perspectives
	for i := range perspectives {
		if err := h.narrativeRepo.CreatePerspectiveNarrative(&perspectives[i]); err != nil {
			continue
		}
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(perspectives); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// GetCharacterMemories retrieves narrative memories for a character
func (h *NarrativeHandlers) GetCharacterMemories(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	vars := mux.Vars(r)
	characterID := vars["characterId"]

	// Verify character ownership
	character, err := h.characterRepo.GetByID(r.Context(), characterID)
	if err != nil {
		http.Error(w, constants.ErrCharacterNotFound, http.StatusNotFound)
		return
	}

	if character.UserID != claims.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	memories, err := h.narrativeRepo.GetActiveMemories(characterID, 20)
	if err != nil {
		http.Error(w, "Failed to retrieve memories", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(memories); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// CreateMemory creates a new narrative memory
func (h *NarrativeHandlers) CreateMemory(w http.ResponseWriter, r *http.Request) {
	HandleCharacterOwnedCreation(w, r, h.characterRepo,
		func(memory *models.NarrativeMemory) string { return memory.CharacterID },
		func(memory *models.NarrativeMemory) error { return h.narrativeRepo.CreateNarrativeMemory(memory) })
}

// GetActiveThreads retrieves active narrative threads
func (h *NarrativeHandlers) GetActiveThreads(w http.ResponseWriter, _ *http.Request) {
	threads, err := h.narrativeRepo.GetActiveNarrativeThreads()
	if err != nil {
		http.Error(w, "Failed to retrieve threads", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	if err := json.NewEncoder(w).Encode(threads); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// CreateThread creates a new narrative thread
func (h *NarrativeHandlers) CreateThread(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())

	var thread models.NarrativeThread
	if err := json.NewDecoder(r.Body).Decode(&thread); err != nil {
		http.Error(w, ErrInvalidRequestBody, http.StatusBadRequest)
		return
	}

	// Verify DM role if session is specified
	if sessionID, ok := thread.Metadata["session_id"].(string); ok {
		participants, err := h.gameRepo.GetParticipants(r.Context(), sessionID)
		if err != nil {
			http.Error(w, errFailedToGetParticipants, http.StatusInternalServerError)
			return
		}

		isDM := false
		for _, p := range participants {
			if p.UserID == claims.UserID && p.Role == models.ParticipantRoleDM {
				isDM = true
				break
			}
		}

		if !isDM {
			http.Error(w, "Only DM can create narrative threads", http.StatusForbidden)
			return
		}
	}

	if err := h.narrativeRepo.CreateNarrativeThread(&thread); err != nil {
		http.Error(w, "Failed to create thread", http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(thread); err != nil {
		http.Error(w, constants.ErrFailedToEncode, http.StatusInternalServerError)
		return
	}
}

// Helper functions

func (h *NarrativeHandlers) buildWorldState(ctx context.Context, sessionID string) map[string]interface{} {
	// Build world state from various sources
	worldState := make(map[string]interface{})

	// Add session info
	session, err := h.gameRepo.GetByID(ctx, sessionID)
	if err == nil {
		worldState["session"] = session
	}

	// Add active threads
	threads, err := h.narrativeRepo.GetActiveNarrativeThreads()
	if err == nil {
		worldState["active_threads"] = threads
	}

	// Additional world state data can be added here
	// - Faction relationships
	// - Economic data
	// - Environmental conditions
	// - etc.

	return worldState
}

func (h *NarrativeHandlers) getRelevantPerspectiveSources(event *models.NarrativeEvent) []models.PerspectiveSource {
	if event == nil {
		return []models.PerspectiveSource{}
	}
	// Get relevant NPCs, factions, etc. that might have perspectives
	sources := []models.PerspectiveSource{}

	// Add participants as sources
	for _, participantID := range event.Participants {
		sources = append(sources, models.PerspectiveSource{
			ID:   participantID,
			Type: "participant",
			Name: participantID, // Would be replaced with actual name from DB
		})
	}

	// Add witnesses
	for _, witnessID := range event.Witnesses {
		sources = append(sources, models.PerspectiveSource{
			ID:   witnessID,
			Type: "witness",
			Name: witnessID,
		})
	}

	// Add a neutral historian perspective
	sources = append(sources, models.PerspectiveSource{
		ID:          "historian",
		Type:        "historical",
		Name:        "Scholar of the Realm",
		Background:  "An impartial observer recording events for posterity",
		Motivations: []string{"accuracy", "completeness", "context"},
	})

	return sources
}

func interfaceSliceToStringSlice(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		if str, ok := v.(string); ok {
			result[i] = str
		}
	}
	return result
}

// verifyActionPermissions checks if the user can perform the action
func (h *NarrativeHandlers) verifyActionPermissions(ctx context.Context, claims *auth.Claims, action *models.PlayerAction) error {
	// Verify character ownership
	character, err := h.characterRepo.GetByID(ctx, action.CharacterID)
	if err != nil {
		return fmt.Errorf(constants.ErrCharacterNotFound)
	}

	if character.UserID != claims.UserID {
		return fmt.Errorf(errUnauthorized)
	}

	// Verify session participation
	participants, err := h.gameRepo.GetParticipants(ctx, action.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get participants")
	}

	for _, p := range participants {
		if p.UserID == claims.UserID {
			return nil
		}
	}

	return fmt.Errorf("not a participant in this session")
}

// handleActionPermissionError sends the appropriate error response
func (h *NarrativeHandlers) handleActionPermissionError(w http.ResponseWriter, err error) {
	switch err.Error() {
	case constants.ErrCharacterNotFound:
		http.Error(w, constants.ErrCharacterNotFound, http.StatusNotFound)
	case errUnauthorized:
		http.Error(w, "Unauthorized", http.StatusForbidden)
	case "not a participant in this session":
		http.Error(w, "Not a participant in this session", http.StatusForbidden)
	default:
		http.Error(w, errFailedToGetParticipants, http.StatusInternalServerError)
	}
}

// processActionConsequences calculates and stores action consequences
func (h *NarrativeHandlers) processActionConsequences(ctx context.Context, action models.PlayerAction) {
	worldState := h.buildWorldState(ctx, action.SessionID)
	consequences, err := h.narrativeEngine.ConsequenceEngine.CalculateConsequences(ctx, action, worldState)
	if err != nil {
		return
	}
	
	for i := range consequences {
		_ = h.narrativeRepo.CreateConsequenceEvent(&consequences[i])
	}
	action.PotentialConsequences = len(consequences)
	_ = h.narrativeRepo.UpdatePlayerAction(&action)
}

// personalizationData holds data needed for narrative personalization
type personalizationData struct {
	event     *models.NarrativeEvent
	profile   *models.NarrativeProfile
	backstory []models.BackstoryElement
}

// gatherPersonalizationData collects all required data for personalization
func (h *NarrativeHandlers) gatherPersonalizationData(ctx context.Context, claims *auth.Claims, characterID, eventID string) (*personalizationData, error) {
	// Verify character ownership
	character, err := h.characterRepo.GetByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrCharacterNotFound)
	}

	if character.UserID != claims.UserID {
		return nil, fmt.Errorf(errUnauthorized)
	}

	// Get event
	event, err := h.narrativeRepo.GetWorldEvent(eventID)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrEventNotFound)
	}

	// Get profile
	profile, err := h.narrativeRepo.GetNarrativeProfile(characterID)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrProfileNotFound)
	}

	// Get backstory (optional)
	backstory, err := h.narrativeRepo.GetBackstoryElements(characterID)
	if err != nil {
		backstory = []models.BackstoryElement{}
	}

	return &personalizationData{
		event:     event,
		profile:   profile,
		backstory: backstory,
	}, nil
}

// handlePersonalizationError sends appropriate error response
func (h *NarrativeHandlers) handlePersonalizationError(w http.ResponseWriter, err error) {
	switch err.Error() {
	case constants.ErrCharacterNotFound:
		http.Error(w, constants.ErrCharacterNotFound, http.StatusNotFound)
	case errUnauthorized:
		http.Error(w, "Unauthorized", http.StatusForbidden)
	case constants.ErrEventNotFound:
		http.Error(w, constants.ErrEventNotFound, http.StatusNotFound)
	case constants.ErrProfileNotFound:
		http.Error(w, constants.ErrProfileNotFound, http.StatusNotFound)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// createPersonalizedNarrative generates and saves a personalized narrative
func (h *NarrativeHandlers) createPersonalizedNarrative(ctx context.Context, data *personalizationData) (*models.PersonalizedNarrative, error) {
	// Generate narrative
	narrative, err := h.narrativeEngine.GeneratePersonalizedNarrative(ctx, data.event, data.profile, data.backstory)
	if err != nil {
		return nil, err
	}

	// Save narrative
	if err := h.narrativeRepo.CreatePersonalizedNarrative(narrative); err != nil {
		return nil, err
	}

	// Update backstory usage
	for _, callback := range narrative.BackstoryCallbacks {
		_ = h.narrativeRepo.IncrementBackstoryUsage(callback.BackstoryElementID)
	}

	return narrative, nil
}

// updateNarrativeProfileWithAction updates the narrative profile based on a player action
func (h *NarrativeHandlers) updateNarrativeProfileWithAction(ctx context.Context, action models.PlayerAction) {
	profile, err := h.narrativeRepo.GetNarrativeProfile(action.CharacterID)
	if err != nil {
		return
	}

	decision := models.DecisionRecord{
		Timestamp:       action.Timestamp,
		Context:         action.ActionDescription,
		Decision:        action.ActionType,
		Consequences:    []string{action.ImmediateResult},
		EmotionalWeight: 0.5, // Default weight
		Tags:            []string{action.ActionType, action.TargetType},
	}

	updatedProfile, err := h.narrativeEngine.ProfileService.AnalyzePlayerDecision(ctx, profile, decision)
	if err == nil {
		_ = h.narrativeRepo.UpdateNarrativeProfile(updatedProfile)
	}
}

// verifyDMRoleForEvent checks if the user has DM role for the event's session
func (h *NarrativeHandlers) verifyDMRoleForEvent(ctx context.Context, claims *auth.Claims, event *models.NarrativeEvent) error {
	sessionID, ok := event.Metadata["session_id"].(string)
	if !ok {
		return nil // No session specified, allow creation
	}

	participants, err := h.gameRepo.GetParticipants(ctx, sessionID)
	if err != nil {
		return fmt.Errorf(errFailedToGetParticipants)
	}

	for _, p := range participants {
		if p.UserID == claims.UserID && p.Role == models.ParticipantRoleDM {
			return nil
		}
	}

	return fmt.Errorf("only DM can create world events")
}

// generateEventPerspectives generates perspectives for an event asynchronously
func (h *NarrativeHandlers) generateEventPerspectives(ctx context.Context, event *models.NarrativeEvent) {
	sources := h.getRelevantPerspectiveSources(event)
	perspectives, err := h.narrativeEngine.PerspectiveGen.GenerateMultiplePerspectives(ctx, event, sources)
	if err != nil {
		return
	}
	
	for i := range perspectives {
		_ = h.narrativeRepo.CreatePerspectiveNarrative(&perspectives[i])
	}
}

// applyProfileUpdates applies updates to a narrative profile
func (h *NarrativeHandlers) applyProfileUpdates(profile *models.NarrativeProfile, updates map[string]interface{}) {
	if preferences, ok := updates["preferences"].(map[string]interface{}); ok {
		h.updatePreferences(&profile.Preferences, preferences)
	}

	if playStyle, ok := updates["play_style"].(string); ok {
		profile.PlayStyle = playStyle
	}
}

// updatePreferences updates story preferences from a map
func (h *NarrativeHandlers) updatePreferences(prefs *models.StoryPreferences, updates map[string]interface{}) {
	if themes, ok := updates["themes"].([]interface{}); ok {
		prefs.Themes = interfaceSliceToStringSlice(themes)
	}
	if tone, ok := updates["tone"].([]interface{}); ok {
		prefs.Tone = interfaceSliceToStringSlice(tone)
	}
	if complexity, ok := updates["complexity"].(float64); ok {
		prefs.Complexity = int(complexity)
	}
	if moral, ok := updates["moral_alignment"].(string); ok {
		prefs.MoralAlignment = moral
	}
	if pacing, ok := updates["pacing"].(string); ok {
		prefs.PacingPreference = pacing
	}
	if combat, ok := updates["combat_narrative"].(float64); ok {
		prefs.CombatNarrative = combat
	}
}
