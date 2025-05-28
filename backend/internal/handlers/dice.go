package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/your-username/dnd-game/backend/pkg/dice"
)

type DiceRollRequest struct {
	DiceType string `json:"diceType"` // e.g., "d20", "2d6", "1d8+3"
	Purpose  string `json:"purpose"`  // attack, damage, skill check, etc.
}

type DiceRollResponse struct {
	Request  DiceRollRequest `json:"request"`
	Results  []int          `json:"results"`
	Modifier int            `json:"modifier"`
	Total    int            `json:"total"`
}

func RollDice(w http.ResponseWriter, r *http.Request) {
	var req DiceRollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	roller := dice.NewRoller()
	result, err := roller.Roll(req.DiceType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := DiceRollResponse{
		Request:  req,
		Results:  result.Dice,
		Modifier: result.Modifier,
		Total:    result.Total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}