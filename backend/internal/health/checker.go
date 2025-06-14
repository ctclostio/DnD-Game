package health

import "context"

// Checker defines a dependency health checker.
// Name returns the component name, Check verifies health.
// returning an error if unhealthy.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// Result represents a health check result.
// Status is "healthy" or "unhealthy" with an optional message.
// when an error occurs.
type Result struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// RunChecks executes all checkers and returns a map keyed by.
// component name with their health results.
func RunChecks(ctx context.Context, checkers ...Checker) map[string]Result {
	results := make(map[string]Result)
	for _, c := range checkers {
		err := c.Check(ctx)
		if err != nil {
			results[c.Name()] = Result{Status: "unhealthy", Message: err.Error()}
		} else {
			results[c.Name()] = Result{Status: "healthy"}
		}
	}
	return results
}
