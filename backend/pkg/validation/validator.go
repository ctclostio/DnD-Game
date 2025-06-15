package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/ctclostio/DnD-Game/backend/pkg/errors"
)

// Validator wraps the go-playground validator
type Validator struct {
	validator *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Register custom tag name function
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validations
	registerCustomValidations(v)

	return &Validator{
		validator: v,
	}
}

// registerCustomValidations registers custom validation rules
func registerCustomValidations(v *validator.Validate) {
	// D&D specific validations
	_ = v.RegisterValidation("dndname", validateDnDName)
	_ = v.RegisterValidation("dicenotation", validateDiceNotation)
	_ = v.RegisterValidation("alignment", validateAlignment)
	_ = v.RegisterValidation("ability", validateAbilityScore)
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}

// ValidateRequest validates and decodes a request body
func (v *Validator) ValidateRequest(r *http.Request, dst interface{}) error {
	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		if err == io.EOF {
			return errors.NewBadRequestError("Request body is empty")
		}
		return errors.NewBadRequestError("Invalid JSON format").WithInternal(err)
	}

	// Validate struct
	return v.Validate(dst)
}

// formatValidationError formats validation errors into AppError
func (v *Validator) formatValidationError(err error) error {
	validationErrors := &errors.ValidationErrors{}

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := fe.Field()
			tag := fe.Tag()
			param := fe.Param()

			message := v.getErrorMessage(field, tag, param)
			validationErrors.Add(field, message)
		}
	}

	return validationErrors.ToAppError()
}

// getErrorMessage returns a user-friendly error message
func (v *Validator) getErrorMessage(field, tag, param string) string {
	messages := map[string]string{
		"required":     fmt.Sprintf("%s is required", field),
		"min":          fmt.Sprintf("%s must be at least %s characters long", field, param),
		"max":          fmt.Sprintf("%s must be at most %s characters long", field, param),
		"email":        fmt.Sprintf("%s must be a valid email address", field),
		"oneof":        fmt.Sprintf("%s must be one of: %s", field, param),
		"numeric":      fmt.Sprintf("%s must be a number", field),
		"alphanum":     fmt.Sprintf("%s must contain only letters and numbers", field),
		"dndname":      fmt.Sprintf("%s must be a valid character name (3-50 characters, letters, spaces, hyphens, and apostrophes only)", field),
		"dicenotation": fmt.Sprintf("%s must be valid dice notation (e.g., 2d6+3)", field),
		"alignment":    fmt.Sprintf("%s must be a valid D&D alignment", field),
		"ability":      fmt.Sprintf("%s must be between 1 and 30", field),
	}

	if msg, ok := messages[tag]; ok {
		return msg
	}

	return fmt.Sprintf("%s failed %s validation", field, tag)
}

// Custom validation functions

// validateDnDName validates D&D character names
func validateDnDName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if len(name) < 3 || len(name) > 50 {
		return false
	}

	// Allow letters, spaces, hyphens, and apostrophes
	for _, char := range name {
		valid := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			char == ' ' || char == '-' || char == '\''
		if !valid {
			return false
		}
	}

	return true
}

// validateDiceNotation validates dice notation (e.g., 2d6+3)
var diceNotationRegex = regexp.MustCompile(`^\d+d\d+(?:[+-]\d+)?$`)

func validateDiceNotation(fl validator.FieldLevel) bool {
	notation := fl.Field().String()
	return diceNotationRegex.MatchString(notation)
}

// validateAlignment validates D&D alignments
func validateAlignment(fl validator.FieldLevel) bool {
	alignment := fl.Field().String()
	validAlignments := []string{
		"Lawful Good", "Neutral Good", "Chaotic Good",
		"Lawful Neutral", "True Neutral", "Chaotic Neutral",
		"Lawful Evil", "Neutral Evil", "Chaotic Evil",
	}

	for _, valid := range validAlignments {
		if alignment == valid {
			return true
		}
	}

	return false
}

// validateAbilityScore validates ability scores (1-30)
func validateAbilityScore(fl validator.FieldLevel) bool {
	score := fl.Field().Int()
	return score >= 1 && score <= 30
}

// Request DTOs with validation tags

// CreateCharacterRequest represents a character creation request
type CreateCharacterRequest struct {
	Name       string                 `json:"name" validate:"required,dndname"`
	Race       string                 `json:"race" validate:"required"`
	Class      string                 `json:"class" validate:"required"`
	Background string                 `json:"background" validate:"required"`
	Alignment  string                 `json:"alignment" validate:"required,alignment"`
	Level      int                    `json:"level" validate:"required,min=1,max=20"`
	Abilities  map[string]int         `json:"abilities" validate:"required,dive,ability"`
	Skills     []string               `json:"skills" validate:"max=10"`
	Equipment  []string               `json:"equipment"`
	Traits     map[string]interface{} `json:"traits"`
}

// UpdateCharacterRequest represents a character update request
type UpdateCharacterRequest struct {
	Name       string                 `json:"name,omitempty" validate:"omitempty,dndname"`
	Background string                 `json:"background,omitempty"`
	Alignment  string                 `json:"alignment,omitempty" validate:"omitempty,alignment"`
	Level      int                    `json:"level,omitempty" validate:"omitempty,min=1,max=20"`
	Abilities  map[string]int         `json:"abilities,omitempty" validate:"omitempty,dive,ability"`
	Skills     []string               `json:"skills,omitempty" validate:"omitempty,max=10"`
	Equipment  []string               `json:"equipment,omitempty"`
	Traits     map[string]interface{} `json:"traits,omitempty"`
}

// DiceRollRequest represents a dice roll request
type DiceRollRequest struct {
	Notation string `json:"notation" validate:"required,dicenotation"`
	Purpose  string `json:"purpose,omitempty" validate:"max=100"`
}

// Global validator instance
var defaultValidator *Validator

// Init initializes the global validator
func Init() {
	defaultValidator = New()
}

// GetValidator returns the global validator instance
func GetValidator() *Validator {
	if defaultValidator == nil {
		Init()
	}
	return defaultValidator
}

// ValidateStruct validates a struct using the global validator
func ValidateStruct(s interface{}) error {
	return GetValidator().Validate(s)
}

// ValidateRequestBody validates and decodes a request body using the global validator
func ValidateRequestBody(r *http.Request, dst interface{}) error {
	return GetValidator().ValidateRequest(r, dst)
}
