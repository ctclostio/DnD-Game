package handlers

// Game Session API documentation.
// CreateGameSession godoc.
// @Summary Create game session.
// @Description Create a new game session.
// @Tags game.
// @Accept json.
// @Produce json.
// @Security Bearer.
// @Param request body models.CreateGameRequest true "Game session details"
// @Success 201 {object} models.GameSession "Created game session"
// @Failure 400 {object} map[string]string "Invalid request".
// @Failure 401 {object} map[string]string "Unauthorized".
// @Failure 500 {object} map[string]string "Internal server error".
// @Router /game/session [post].
// GetGameSession godoc.
// @Summary Get game session.
// @Description Get details of a specific game session.
// @Tags game.
// @Accept json.
// @Produce json.
// @Security Bearer.
// @Param id path string true "Session ID".
// @Success 200 {object} models.GameSession "Game session details"
// @Failure 401 {object} map[string]string "Unauthorized".
// @Failure 404 {object} map[string]string "Session not found".
// @Failure 500 {object} map[string]string "Internal server error".
// @Router /game/session/{id} [get].
// JoinGameSession godoc.
// @Summary Join game session.
// @Description Join an existing game session.
// @Tags game.
// @Accept json.
// @Produce json.
// @Security Bearer.
// @Param id path string true "Session ID".
// @Param request body models.JoinGameRequest true "Join request"
// @Success 200 {object} map[string]string "Successfully joined".
// @Failure 400 {object} map[string]string "Invalid request".
// @Failure 401 {object} map[string]string "Unauthorized".
// @Failure 404 {object} map[string]string "Session not found".
// @Failure 409 {object} map[string]string "Session full".
// @Router /game/session/{id}/join [post].
// LeaveGameSession godoc.
// @Summary Leave game session.
// @Description Leave a game session.
// @Tags game.
// @Accept json.
// @Produce json.
// @Security Bearer.
// @Param id path string true "Session ID".
// @Success 200 {object} map[string]string "Successfully left".
// @Failure 401 {object} map[string]string "Unauthorized".
// @Failure 404 {object} map[string]string "Session not found".
// @Router /game/session/{id}/leave [post].
// ListGameSessions godoc.
// @Summary List game sessions.
// @Description Get a list of active game sessions.
// @Tags game.
// @Accept json.
// @Produce json.
// @Security Bearer.
// @Success 200 {array} models.GameSession "List of game sessions"
// @Failure 401 {object} map[string]string "Unauthorized".
// @Failure 500 {object} map[string]string "Internal server error".
// @Router /game/sessions [get].
