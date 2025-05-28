package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
)

var characterService = services.NewCharacterService()

func GetCharacters(w http.ResponseWriter, r *http.Request) {
	characters, err := characterService.GetAllCharacters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(characters)
}

func GetCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	character, err := characterService.GetCharacterByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(character)
}

func CreateCharacter(w http.ResponseWriter, r *http.Request) {
	var character models.Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := characterService.CreateCharacter(&character)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func UpdateCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var character models.Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	character.ID = id
	updated, err := characterService.UpdateCharacter(&character)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}