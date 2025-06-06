package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/your-username/dnd-game/backend/pkg/errors"
	"github.com/your-username/dnd-game/backend/pkg/response"
)

// ValidationMiddlewareV2 provides enhanced request validation
type ValidationMiddlewareV2 struct {
	validator *validator.Validate
}

// NewValidationMiddlewareV2 creates a new validation middleware
func NewValidationMiddlewareV2() *ValidationMiddlewareV2 {
	v := validator.New()
	
	// Register custom tag name function
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	// Register custom validators
	registerCustomValidators(v)
	
	return &ValidationMiddlewareV2{
		validator: v,
	}
}

// ValidateRequest validates a request against a struct type
func (vm *ValidationMiddlewareV2) ValidateRequest(structType interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate for methods with body
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}
			
			// Create new instance of the struct type
			reqType := reflect.TypeOf(structType)
			if reqType.Kind() == reflect.Ptr {
				reqType = reqType.Elem()
			}
			reqValue := reflect.New(reqType)
			req := reqValue.Interface()
			
			// Decode JSON body
			if err := json.NewDecoder(r.Body).Decode(req); err != nil {
				response.BadRequest(w, r, "Invalid JSON in request body")
				return
			}
			
			// Validate struct
			if err := vm.validator.Struct(req); err != nil {
				// Convert validation errors to our format
				validationErrors := vm.formatValidationErrors(err)
				response.ValidationError(w, r, validationErrors)
				return
			}
			
			// Store validated request in context for handler use
			ctx := context.WithValue(r.Context(), "validated_request", req)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}
}

// formatValidationErrors converts validator errors to our ValidationErrors format
func (vm *ValidationMiddlewareV2) formatValidationErrors(err error) *errors.ValidationErrors {
	validationErrors := &errors.ValidationErrors{}
	
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()
			
			// Generate user-friendly error message
			message := vm.getErrorMessage(field, tag, param, e.Value())
			validationErrors.Add(field, message)
		}
	}
	
	return validationErrors
}

// getErrorMessage generates user-friendly validation error messages
func (vm *ValidationMiddlewareV2) getErrorMessage(field, tag, param string, value interface{}) string {
	switch tag {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + param
	case "max":
		return field + " must be at most " + param
	case "len":
		return field + " must be exactly " + param + " characters"
	case "email":
		return field + " must be a valid email address"
	case "url":
		return field + " must be a valid URL"
	case "oneof":
		return field + " must be one of: " + param
	case "gt":
		return field + " must be greater than " + param
	case "gte":
		return field + " must be greater than or equal to " + param
	case "lt":
		return field + " must be less than " + param
	case "lte":
		return field + " must be less than or equal to " + param
	case "eqfield":
		return field + " must equal " + param
	case "nefield":
		return field + " must not equal " + param
	case "alpha":
		return field + " must contain only letters"
	case "alphanum":
		return field + " must contain only letters and numbers"
	case "numeric":
		return field + " must be a valid number"
	case "hexadecimal":
		return field + " must be a valid hexadecimal"
	case "uuid":
		return field + " must be a valid UUID"
	case "uuid4":
		return field + " must be a valid UUID v4"
	case "rgb":
		return field + " must be a valid RGB color"
	case "rgba":
		return field + " must be a valid RGBA color"
	case "hsl":
		return field + " must be a valid HSL color"
	case "hsla":
		return field + " must be a valid HSLA color"
	case "json":
		return field + " must be valid JSON"
	case "lowercase":
		return field + " must be lowercase"
	case "uppercase":
		return field + " must be uppercase"
	case "datetime":
		return field + " must be a valid datetime"
	
	// D&D specific validators
	case "dnd_ability_score":
		return field + " must be between 3 and 20"
	case "dnd_level":
		return field + " must be between 1 and 20"
	case "dnd_alignment":
		return field + " must be a valid D&D alignment"
	case "dnd_dice_notation":
		return field + " must be valid dice notation (e.g., 2d6+3)"
	case "dnd_skill":
		return field + " must be a valid D&D skill"
	case "dnd_ability":
		return field + " must be a valid D&D ability (str, dex, con, int, wis, cha)"
		
	default:
		return field + " failed " + tag + " validation"
	}
}

// registerCustomValidators registers D&D-specific validators
func registerCustomValidators(v *validator.Validate) {
	// D&D ability score (3-20)
	v.RegisterValidation("dnd_ability_score", func(fl validator.FieldLevel) bool {
		if value, ok := fl.Field().Interface().(int); ok {
			return value >= 3 && value <= 20
		}
		return false
	})
	
	// D&D level (1-20)
	v.RegisterValidation("dnd_level", func(fl validator.FieldLevel) bool {
		if value, ok := fl.Field().Interface().(int); ok {
			return value >= 1 && value <= 20
		}
		return false
	})
	
	// D&D alignment
	v.RegisterValidation("dnd_alignment", func(fl validator.FieldLevel) bool {
		validAlignments := map[string]bool{
			"lawful good":     true,
			"neutral good":    true,
			"chaotic good":    true,
			"lawful neutral":  true,
			"true neutral":    true,
			"chaotic neutral": true,
			"lawful evil":     true,
			"neutral evil":    true,
			"chaotic evil":    true,
		}
		
		if value, ok := fl.Field().Interface().(string); ok {
			return validAlignments[strings.ToLower(value)]
		}
		return false
	})
	
	// D&D dice notation (e.g., 2d6+3)
	v.RegisterValidation("dnd_dice_notation", func(fl validator.FieldLevel) bool {
		if value, ok := fl.Field().Interface().(string); ok {
			// Simple regex for dice notation
			pattern := `^\d+d\d+([+-]\d+)?$`
			return regexp.MustCompile(pattern).MatchString(value)
		}
		return false
	})
	
	// D&D skill
	v.RegisterValidation("dnd_skill", func(fl validator.FieldLevel) bool {
		validSkills := map[string]bool{
			"acrobatics":      true,
			"animal handling": true,
			"arcana":          true,
			"athletics":       true,
			"deception":       true,
			"history":         true,
			"insight":         true,
			"intimidation":    true,
			"investigation":   true,
			"medicine":        true,
			"nature":          true,
			"perception":      true,
			"performance":     true,
			"persuasion":      true,
			"religion":        true,
			"sleight of hand": true,
			"stealth":         true,
			"survival":        true,
		}
		
		if value, ok := fl.Field().Interface().(string); ok {
			return validSkills[strings.ToLower(value)]
		}
		return false
	})
	
	// D&D ability
	v.RegisterValidation("dnd_ability", func(fl validator.FieldLevel) bool {
		validAbilities := map[string]bool{
			"strength":     true,
			"str":          true,
			"dexterity":    true,
			"dex":          true,
			"constitution": true,
			"con":          true,
			"intelligence": true,
			"int":          true,
			"wisdom":       true,
			"wis":          true,
			"charisma":     true,
			"cha":          true,
		}
		
		if value, ok := fl.Field().Interface().(string); ok {
			return validAbilities[strings.ToLower(value)]
		}
		return false
	})
}

// GetValidatedRequest retrieves the validated request from context
func GetValidatedRequest[T any](r *http.Request) (T, error) {
	var zero T
	
	val := r.Context().Value("validated_request")
	if val == nil {
		return zero, errors.NewInternalError("No validated request in context", nil)
	}
	
	typed, ok := val.(T)
	if !ok {
		// Try pointer type
		if ptr, ok := val.(*T); ok {
			return *ptr, nil
		}
		return zero, errors.NewInternalError("Invalid request type in context", nil)
	}
	
	return typed, nil
}

// Example usage with struct tags:
/*
type CreateCharacterRequest struct {
    Name       string `json:"name" validate:"required,min=1,max=50"`
    Race       string `json:"race" validate:"required"`
    Class      string `json:"class" validate:"required"`
    Level      int    `json:"level" validate:"required,dnd_level"`
    Alignment  string `json:"alignment" validate:"required,dnd_alignment"`
    Attributes struct {
        Strength     int `json:"strength" validate:"required,dnd_ability_score"`
        Dexterity    int `json:"dexterity" validate:"required,dnd_ability_score"`
        Constitution int `json:"constitution" validate:"required,dnd_ability_score"`
        Intelligence int `json:"intelligence" validate:"required,dnd_ability_score"`
        Wisdom       int `json:"wisdom" validate:"required,dnd_ability_score"`
        Charisma     int `json:"charisma" validate:"required,dnd_ability_score"`
    } `json:"attributes" validate:"required"`
}
*/