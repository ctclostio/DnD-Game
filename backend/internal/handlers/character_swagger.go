package handlers

// Character API documentation

// GetCharacters godoc
// @Summary List all characters
// @Description Get a list of all characters for the authenticated user
// @Tags characters
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} models.Character "List of characters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /characters [get]

// GetCharacter godoc
// @Summary Get character by ID
// @Description Get detailed information about a specific character
// @Tags characters
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Character ID"
// @Success 200 {object} models.Character "Character details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Character not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /characters/{id} [get]

// CreateCharacter godoc
// @Summary Create a new character
// @Description Create a new D&D character
// @Tags characters
// @Accept json
// @Produce json
// @Security Bearer
// @Param character body models.Character true "Character details"
// @Success 201 {object} models.Character "Created character"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /characters [post]

// UpdateCharacter godoc
// @Summary Update character
// @Description Update an existing character's information
// @Tags characters
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Character ID"
// @Param character body models.Character true "Updated character details"
// @Success 200 {object} models.Character "Updated character"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Character not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /characters/{id} [put]

// DeleteCharacter godoc
// @Summary Delete character
// @Description Delete a character
// @Tags characters
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Character ID"
// @Success 204 "Character deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Character not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /characters/{id} [delete]

// GetCharacterOptions godoc
// @Summary Get character creation options
// @Description Get available races, classes, backgrounds, and skills for character creation
// @Tags characters
// @Accept json
// @Produce json
// @Success 200 {object} models.CharacterOptions "Character creation options"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /characters/options [get]