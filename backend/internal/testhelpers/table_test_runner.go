package testhelpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TableTestRunner is a helper for running table-driven tests with common patterns
type TableTestRunner[T any] struct {
	t *testing.T
}

// NewTableTestRunner creates a new table test runner
func NewTableTestRunner[T any](t *testing.T) *TableTestRunner[T] {
	return &TableTestRunner[T]{t: t}
}

// TestCase represents a single test case in a table-driven test
type TestCase[T any] struct {
	Name          string
	SetupMock     func(mock T)
	Execute       func() error
	ExpectedError string
}

// Run executes all test cases
func (r *TableTestRunner[T]) Run(tests []TestCase[T], mockFactory func() T) {
	for _, tt := range tests {
		r.t.Run(tt.Name, func(t *testing.T) {
			mock := mockFactory()
			if tt.SetupMock != nil {
				tt.SetupMock(mock)
			}

			err := tt.Execute()

			if tt.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.ExpectedError)
			} else {
				require.NoError(t, err)
			}

			// Assert expectations if the mock has an AssertExpectations method
			if asserter, ok := any(mock).(interface{ AssertExpectations(*testing.T) }); ok {
				asserter.AssertExpectations(t)
			}
		})
	}
}

// ServiceTestHelper provides common service test functionality
type ServiceTestHelper struct {
	Ctx context.Context
}

// NewServiceTestHelper creates a new service test helper
func NewServiceTestHelper() *ServiceTestHelper {
	return &ServiceTestHelper{
		Ctx: context.Background(),
	}
}
