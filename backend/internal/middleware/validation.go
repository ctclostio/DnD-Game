package middleware

import (
	"net/http"

	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
	"github.com/ctclostio/DnD-Game/backend/pkg/validation"
)

// ValidationMiddleware provides request validation.
type ValidationMiddleware struct {
	validator *validation.Validator
}

// NewValidationMiddleware creates a new validation middleware.
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validation.New(),
	}
}

// Validate returns a middleware that validates request bodies.
func (vm *ValidationMiddleware) Validate(targetStruct interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate for methods with body
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				if err := vm.validator.ValidateRequest(r, targetStruct); err != nil {
					SendError(w, err, logger.GetLogger().WithContext(r.Context()))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
