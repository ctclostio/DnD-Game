package testhelpers

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

// MockCall represents a mock method call configuration
type MockCall struct {
	Method    string
	Arguments []interface{}
	Returns   []interface{}
	RunFunc   func(args mock.Arguments)
	Times     int // 0 means any number of times
}

// SetupMockCalls configures multiple mock calls on a mock object
func SetupMockCalls(mockObj *mock.Mock, calls []MockCall) {
	for _, call := range calls {
		mockCall := mockObj.On(call.Method, call.Arguments...)
		
		if call.Returns != nil {
			mockCall.Return(call.Returns...)
		}
		
		if call.RunFunc != nil {
			mockCall.Run(call.RunFunc)
		}
		
		if call.Times > 0 {
			mockCall.Times(call.Times)
		}
	}
}

// AssertMockExpectations asserts all expectations on multiple mocks
func AssertMockExpectations(t *testing.T, mocks ...interface{}) {
	t.Helper()
	
	for _, m := range mocks {
		if mockObj, ok := m.(interface{ AssertExpectations(t *testing.T) bool }); ok {
			mockObj.AssertExpectations(t)
		}
	}
}

// MockService provides a generic base for service mocks
type MockService struct {
	mock.Mock
}

// NewMockService creates a new generic mock service
func NewMockService() *MockService {
	return &MockService{}
}

// ExpectCall sets up an expectation for a method call
func (m *MockService) ExpectCall(method string, args ...interface{}) *mock.Call {
	return m.On(method, args...)
}

// ExpectCallWithReturn sets up an expectation with return values
func (m *MockService) ExpectCallWithReturn(method string, returns []interface{}, args ...interface{}) {
	m.On(method, args...).Return(returns...)
}

// Common mock return helpers

// SuccessCall creates a MockCall that returns success (nil error)
func SuccessCall(method string, args ...interface{}) MockCall {
	return MockCall{
		Method:    method,
		Arguments: args,
		Returns:   []interface{}{nil},
	}
}

// ErrorCall creates a MockCall that returns an error
func ErrorCall(method string, err error, args ...interface{}) MockCall {
	return MockCall{
		Method:    method,
		Arguments: args,
		Returns:   []interface{}{err},
	}
}

// DataCall creates a MockCall that returns data and nil error
func DataCall(method string, data interface{}, args ...interface{}) MockCall {
	return MockCall{
		Method:    method,
		Arguments: args,
		Returns:   []interface{}{data, nil},
	}
}

// DataErrorCall creates a MockCall that returns data and an error
func DataErrorCall(method string, data interface{}, err error, args ...interface{}) MockCall {
	return MockCall{
		Method:    method,
		Arguments: args,
		Returns:   []interface{}{data, err},
	}
}

// AnyArgs returns a slice of mock.Anything for the specified count
func AnyArgs(count int) []interface{} {
	args := make([]interface{}, count)
	for i := range args {
		args[i] = mock.Anything
	}
	return args
}

// MockRepository provides common database mock functionality
type MockRepository struct {
	MockService
	ShouldFailNext bool
	FailureError   error
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// SetNextFailure configures the next call to fail
func (m *MockRepository) SetNextFailure(err error) {
	m.ShouldFailNext = true
	m.FailureError = err
}

// CheckFailure checks if the next call should fail and resets the flag
func (m *MockRepository) CheckFailure() error {
	if m.ShouldFailNext {
		m.ShouldFailNext = false
		return m.FailureError
	}
	return nil
}

// MockCache provides common cache mock functionality
type MockCache struct {
	MockService
	Store map[string]interface{}
}

// NewMockCache creates a new mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		Store: make(map[string]interface{}),
	}
}

// SetValue sets a value in the mock cache
func (m *MockCache) SetValue(key string, value interface{}) {
	m.Store[key] = value
}

// GetValue gets a value from the mock cache
func (m *MockCache) GetValue(key string) (interface{}, bool) {
	val, exists := m.Store[key]
	return val, exists
}

// Clear clears the mock cache
func (m *MockCache) Clear() {
	m.Store = make(map[string]interface{})
}