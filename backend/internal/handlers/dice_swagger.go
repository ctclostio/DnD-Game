package handlers

// Dice API documentation.
// RollDice godoc.
// @Summary Roll dice.
// @Description Roll dice using standard D&D notation (e.g., "2d6+3")
// @Tags dice.
// @Accept json.
// @Produce json.
// @Security Bearer.
// @Param request body models.DiceRollRequest true "Dice notation"
// @Success 200 {object} models.DiceRollResponse "Roll results"
// @Failure 400 {object} map[string]string "Invalid dice notation".
// @Failure 401 {object} map[string]string "Unauthorized".
// @Router /dice/roll [post].
// Example request:.
// {.
//   "notation": "2d20+5",.
//   "purpose": "Attack roll".
// }.
//
// Example response:.
// {.
//   "notation": "2d20+5",.
//   "rolls": [15, 8],.
//   "modifier": 5,.
//   "total": 28,.
//   "purpose": "Attack roll".
// }.
