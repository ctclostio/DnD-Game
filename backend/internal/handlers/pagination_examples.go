package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/pagination"
	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
)

// GetCharactersPaginated handles paginated character list requests
// @Summary Get paginated characters
// @Description Get a paginated list of characters for the authenticated user
// @Tags characters
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param sort_by query string false "Sort field (name, level, class, race, created_at)"
// @Param sort_dir query string false "Sort direction (asc, desc)"
// @Param filter_class query string false "Filter by class"
// @Param filter_race query string false "Filter by race"
// @Param filter_min_level query int false "Filter by minimum level"
// @Success 200 {object} pagination.PageResult{data=[]models.Character}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/characters [get]
func (h *Handlers) GetCharactersPaginated(w http.ResponseWriter, r *http.Request) {
	// In actual implementation, we would get context and userID
	// ctx := r.Context()
	// userID, _ := auth.GetUserIDFromContext(ctx)

	// Parse pagination parameters
	params := pagination.FromRequest(r)
	
	// Validate parameters
	if err := params.Validate(); err != nil {
		response.Error(w, r, errors.NewValidationError(err.Error()))
		return
	}

	// Example: Get paginated characters
	// This would typically call a repository method like:
	// result, err := h.characterService.GetCharactersPaginated(ctx, userID, params)
	// For this example, we'll return a placeholder
	result := pagination.NewPageResult([]interface{}{}, params, 0)

	// Generate pagination links
	baseURL := r.URL.Path
	if result.Pagination.TotalPages > 0 {
		links := pagination.GenerateLinks(baseURL, params, result.Pagination.TotalPages)
		
		// Add links to response headers
		if links.First != "" {
			w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="first"`, links.First))
		}
		if links.Last != "" {
			w.Header().Add("Link", fmt.Sprintf(`<%s>; rel="last"`, links.Last))
		}
		if links.Prev != "" {
			w.Header().Add("Link", fmt.Sprintf(`<%s>; rel="prev"`, links.Prev))
		}
		if links.Next != "" {
			w.Header().Add("Link", fmt.Sprintf(`<%s>; rel="next"`, links.Next))
		}
	}

	// Write pagination headers
	pagination.WritePaginationHeaders(w, result.Pagination)

	// Set cache control
	w.Header().Set(constants.CacheControl, "private, max-age=60")

	response.JSON(w, r, http.StatusOK, result)
}

// GetGameSessionsPaginated handles paginated game session requests
// @Summary Get paginated game sessions
// @Description Get a paginated list of game sessions
// @Tags sessions
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param sort_by query string false "Sort field"
// @Param sort_dir query string false "Sort direction"
// @Param filter_status query string false "Filter by status (active, pending, ended)"
// @Param filter_dm_user_id query string false "Filter by DM user ID"
// @Success 200 {object} pagination.PageResult{data=[]models.GameSession}
// @Router /api/game-sessions [get]
func (h *Handlers) GetGameSessionsPaginated(w http.ResponseWriter, r *http.Request) {
	// ctx would be used in actual implementation
	// ctx := r.Context()

	// Parse pagination parameters
	params := pagination.FromRequest(r)
	
	// Validate parameters
	if err := params.Validate(); err != nil {
		response.Error(w, r, errors.NewValidationError(err.Error()))
		return
	}

	// Example: Get paginated sessions
	// This would typically call a service method like:
	// result, err := h.gameService.GetGameSessionsPaginated(ctx, params)
	// For this example, we'll return a placeholder
	result := pagination.NewPageResult([]interface{}{}, params, 0)

	// Write pagination headers
	pagination.WritePaginationHeaders(w, result.Pagination)

	// Set cache control for active sessions (shorter cache)
	if status, ok := params.Filters["status"].(string); ok && status == "active" {
		w.Header().Set(constants.CacheControl, "private, max-age=30")
	} else {
		w.Header().Set(constants.CacheControl, "private, max-age=300")
	}

	response.JSON(w, r, http.StatusOK, result)
}

// GetCharactersCursor handles cursor-based pagination for characters
// @Summary Get characters with cursor pagination
// @Description Get characters using cursor-based pagination for efficient scrolling
// @Tags characters
// @Accept json
// @Produce json
// @Security Bearer
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Items per page (default: 20)"
// @Param sort_dir query string false "Sort direction (asc, desc)"
// @Success 200 {object} pagination.CursorResult{data=[]models.Character}
// @Router /api/characters/cursor [get]
func (h *Handlers) GetCharactersCursor(w http.ResponseWriter, r *http.Request) {
	// In actual implementation, we would get context and userID
	// ctx := r.Context()
	// userID, _ := auth.GetUserIDFromContext(ctx)

	// Parse pagination parameters
	params := pagination.FromRequest(r)
	
	// Example: Get cursor-paginated characters
	// This would typically use a cursor pagination helper:
	// result, err := h.characterService.GetCharactersCursor(ctx, userID, params)
	// For this example, we'll return a placeholder
	result := &pagination.CursorResult{
		Data: []interface{}{},
		Pagination: pagination.CursorInfo{
			Limit:      params.Limit,
			HasMore:    false,
			NextCursor: "",
			PrevCursor: "",
		},
	}

	// Set cache control
	w.Header().Set(constants.CacheControl, "private, max-age=60")

	response.JSON(w, r, http.StatusOK, result)
}

// SearchCharacters handles paginated character search
// @Summary Search characters
// @Description Search characters with full-text search and pagination
// @Tags characters
// @Accept json
// @Produce json
// @Security Bearer
// @Param q query string true "Search query"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} pagination.PageResult{data=[]models.Character}
// @Router /api/characters/search [get]
func (h *Handlers) SearchCharacters(w http.ResponseWriter, r *http.Request) {
	// In actual implementation, we would get context and userID
	// ctx := r.Context()
	// userID, _ := auth.GetUserIDFromContext(ctx)

	// Get search query
	query := r.URL.Query().Get("q")
	if query == "" {
		response.Error(w, r, errors.NewValidationError("search query is required"))
		return
	}

	// Parse pagination parameters
	params := pagination.FromRequest(r)
	
	// Example: Perform search with pagination
	// This would typically call a search method:
	// result, err := h.characterService.SearchCharacters(ctx, userID, query, params)
	// For this example, we'll return a placeholder
	result := pagination.NewPageResult([]interface{}{}, params, 0)

	// No cache for search results
	w.Header().Set(constants.CacheControl, "no-cache, no-store, must-revalidate")

	// Write pagination headers
	pagination.WritePaginationHeaders(w, result.Pagination)

	response.JSON(w, r, http.StatusOK, result)
}

// GetCampaignsPaginated handles paginated campaign requests
// @Summary Get paginated campaigns
// @Description Get campaigns the user owns or participates in
// @Tags campaigns
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param filter_status query string false "Filter by status"
// @Success 200 {object} pagination.PageResult{data=[]models.Campaign}
// @Router /api/campaigns [get]
func (h *Handlers) GetCampaignsPaginated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.GetUserIDFromContext(ctx)

	// Parse pagination parameters
	params := pagination.FromRequest(r)
	
	// Example: Get paginated campaigns
	// This would typically call a service method:
	// result, err := h.campaignService.GetCampaignsPaginated(ctx, userID, params)
	// For this example, we'll return a placeholder
	result := pagination.NewPageResult([]interface{}{}, params, 0)

	// Add metadata to response
	enrichedResult := map[string]interface{}{
		"data":       result.Data,
		"pagination": result.Pagination,
		"meta": map[string]interface{}{
			"user_id": userID,
			"filters": params.Filters,
		},
	}

	// Write pagination headers
	pagination.WritePaginationHeaders(w, result.Pagination)

	response.JSON(w, r, http.StatusOK, enrichedResult)
}

// Middleware example for automatic pagination

// PaginationMiddleware automatically adds pagination to list endpoints
func PaginationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply to GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Parse and validate pagination params
		params := pagination.FromRequest(r)
		if err := params.Validate(); err != nil {
			// Use defaults if validation fails
			params = pagination.DefaultPaginationParams()
		}

		// Store params in context for handlers to use
		ctx := context.WithValue(r.Context(), "pagination", params)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Helper function for handlers to get pagination params from context
func GetPaginationParams(ctx context.Context) *pagination.PaginationParams {
	if params, ok := ctx.Value("pagination").(*pagination.PaginationParams); ok {
		return params
	}
	return pagination.DefaultPaginationParams()
}

// Example of using pagination in a service method
// This would be implemented in the services package, not handlers:
//
// func (s *CharacterService) GetUserCharactersWithStats(ctx context.Context, userID string, params *pagination.PaginationParams) (*pagination.PageResult, error) {
// 	// Get paginated characters
// 	result, err := s.repo.GetCharactersPaginated(ctx, userID, params)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Enhance with additional stats
// 	characters := result.Data.([]*models.Character)
// 	for _, char := range characters {
// 		// Add computed stats
// 		char.TotalPlayTime = s.calculatePlayTime(ctx, char.ID)
// 		char.RecentActivity = s.getRecentActivity(ctx, char.ID)
// 	}
//
// 	return result, nil
// }

// Batch processing example using pagination

// ExportAllCharacters demonstrates batch processing with pagination
func (h *Handlers) ExportAllCharacters(w http.ResponseWriter, r *http.Request) {
	// In actual implementation, we would get context and userID
	// ctx := r.Context()
	// userID, _ := auth.GetUserIDFromContext(ctx)

	// Example: Batch export using pagination
	// In a real implementation, you would:
	// 1. Use pagination to fetch characters in batches
	// 2. Stream the results to avoid memory issues
	// 3. Handle errors gracefully

	w.Header().Set(constants.ContentType, constants.ApplicationJSON)
	w.Header().Set("Content-Disposition", "attachment; filename=characters.json")

	// For this example, we'll just return an empty array
	// In production, you'd paginate through all characters
	characters := []*models.Character{}
	
	// Example pagination loop (commented out for simplicity):
	// page := 1
	// for {
	//     params := &pagination.PaginationParams{Page: page, Limit: 100}
	//     result, err := h.characterService.GetCharactersPaginated(ctx, userID, params)
	//     if err != nil {
	//         // Handle error
	//         break
	//     }
	//     
	//     characters = append(characters, result.Data.([]*models.Character)...)
	//     
	//     if !result.Pagination.HasMore {
	//         break
	//     }
	//     page++
	// }

	// Encode and write response
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(characters); err != nil {
		// In production, handle this error appropriately
		response.Error(w, r, errors.Wrap(err, "failed to encode characters"))
		return
	}
}