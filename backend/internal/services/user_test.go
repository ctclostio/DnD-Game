package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// MockUserRepository is already defined in user_test.go
// Ensure it implements all methods from database.UserRepository

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.RegisterRequest
		setupMock     func(*MockUserRepository)
		expectedError string
		validate      func(*testing.T, *models.User)
	}{
		{
			name: "successful registration",
			request: models.RegisterRequest{
				Username: "newuser",
				Email:    "newuser@example.com",
				Password: "SecurePass123!",
			},
			setupMock: func(m *MockUserRepository) {
				// Check username doesn't exist
				m.On("GetByUsername", ctx, "newuser").Return(nil, errors.New("not found"))
				// Check email doesn't exist
				m.On("GetByEmail", ctx, "newuser@example.com").Return(nil, errors.New("not found"))
				// Create user
				m.On("Create", ctx, mock.MatchedBy(func(u *models.User) bool {
					return u.Username == "newuser" &&
						u.Email == "newuser@example.com" &&
						u.PasswordHash != "" &&
						u.PasswordHash != "SecurePass123!" // Should be hashed
				})).Return(nil)
			},
			validate: func(t *testing.T, u *models.User) {
				assert.Equal(t, "newuser", u.Username)
				assert.Equal(t, "newuser@example.com", u.Email)
				assert.NotEmpty(t, u.PasswordHash)
				// Verify password was hashed correctly
				err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("SecurePass123!"))
				assert.NoError(t, err)
			},
		},
		{
			name: "empty username",
			request: models.RegisterRequest{
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			expectedError: "username is required",
		},
		{
			name: "empty email",
			request: models.RegisterRequest{
				Username: "testuser",
				Password: "SecurePass123!",
			},
			expectedError: "email is required",
		},
		{
			name: "empty password",
			request: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
			},
			expectedError: "password is required",
		},
		{
			name: "password too short",
			request: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "short",
			},
			expectedError: "password must be at least 8 characters long",
		},
		{
			name: "username already taken",
			request: models.RegisterRequest{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: "SecurePass123!",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "existinguser").Return(&models.User{
					ID:       "user-123",
					Username: "existinguser",
				}, nil)
			},
			expectedError: "username already taken",
		},
		{
			name: "email already registered",
			request: models.RegisterRequest{
				Username: "newuser",
				Email:    "existing@example.com",
				Password: "SecurePass123!",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "newuser").Return(nil, errors.New("not found"))
				m.On("GetByEmail", ctx, "existing@example.com").Return(&models.User{
					ID:    "user-456",
					Email: "existing@example.com",
				}, nil)
			},
			expectedError: "email already registered",
		},
		{
			name: "repository create error",
			request: models.RegisterRequest{
				Username: "newuser",
				Email:    "new@example.com",
				Password: "SecurePass123!",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "newuser").Return(nil, errors.New("not found"))
				m.On("GetByEmail", ctx, "new@example.com").Return(nil, errors.New("not found"))
				m.On("Create", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewUserService(mockRepo)
			user, err := service.Register(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				if tt.validate != nil {
					tt.validate(t, user)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()

	// Create a test password and hash
	testPassword := "SecurePass123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		request       models.LoginRequest
		setupMock     func(*MockUserRepository)
		expectedError string
		validate      func(*testing.T, *models.AuthResponse)
	}{
		{
			name: "successful login",
			request: models.LoginRequest{
				Username: "testuser",
				Password: testPassword,
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "testuser").Return(&models.User{
					ID:           "user-123",
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "player",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil)
			},
			validate: func(t *testing.T, resp *models.AuthResponse) {
				assert.Equal(t, "Bearer", resp.TokenType)
				assert.Equal(t, "user-123", resp.User.ID)
				assert.Equal(t, "testuser", resp.User.Username)
				assert.Equal(t, "test@example.com", resp.User.Email)
				assert.Empty(t, resp.User.PasswordHash) // Should not expose password hash
			},
		},
		{
			name: "user not found",
			request: models.LoginRequest{
				Username: "nonexistent",
				Password: "password",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "invalid username or password",
		},
		{
			name: "incorrect password",
			request: models.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "testuser").Return(&models.User{
					ID:           "user-123",
					Username:     "testuser",
					PasswordHash: string(hashedPassword),
				}, nil)
			},
			expectedError: "invalid username or password",
		},
		{
			name: "empty username",
			request: models.LoginRequest{
				Password: "password",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "").Return(nil, errors.New("not found"))
			},
			expectedError: "invalid username or password",
		},
		{
			name: "empty password",
			request: models.LoginRequest{
				Username: "testuser",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "testuser").Return(&models.User{
					ID:           "user-123",
					Username:     "testuser",
					PasswordHash: string(hashedPassword),
				}, nil)
			},
			expectedError: "invalid username or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewUserService(mockRepo)
			resp, err := service.Login(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        string
		setupMock     func(*MockUserRepository)
		expected      *models.User
		expectedError string
	}{
		{
			name:   "successful retrieval",
			userID: "user-123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", ctx, "user-123").Return(&models.User{
					ID:        "user-123",
					Username:  "testuser",
					Email:     "test@example.com",
					CreatedAt: time.Now(),
				}, nil)
			},
			expected: &models.User{
				ID:       "user-123",
				Username: "testuser",
				Email:    "test@example.com",
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:   "empty user ID",
			userID: "",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", ctx, "").Return(nil, errors.New("invalid ID"))
			},
			expectedError: "invalid ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewUserService(mockRepo)
			user, err := service.GetUserByID(ctx, tt.userID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.expected.ID, user.ID)
				assert.Equal(t, tt.expected.Username, user.Username)
				assert.Equal(t, tt.expected.Email, user.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name          string
		user          *models.User
		setupMock     func(*MockUserRepository)
		expectedError string
		validate      func(*testing.T, *models.User)
	}{
		{
			name: "successful update",
			user: &models.User{
				ID:       "user-123",
				Username: "updateduser",
				Email:    "updated@example.com",
			},
			setupMock: func(m *MockUserRepository) {
				// GetByID returns existing user
				m.On("GetByID", ctx, "user-123").Return(&models.User{
					ID:           "user-123",
					Username:     "olduser",
					Email:        "old@example.com",
					PasswordHash: "existinghash",
					CreatedAt:    now,
				}, nil)
				// Update with preserved fields
				m.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
					return u.ID == "user-123" &&
						u.Username == "updateduser" &&
						u.Email == "updated@example.com" &&
						u.PasswordHash == "existinghash" && // Should preserve
						u.CreatedAt.Equal(now) // Should preserve
				})).Return(nil)
			},
			validate: func(t *testing.T, u *models.User) {
				assert.Equal(t, "existinghash", u.PasswordHash)
				assert.Equal(t, now, u.CreatedAt)
			},
		},
		{
			name: "missing user ID",
			user: &models.User{
				Username: "test",
				Email:    "test@example.com",
			},
			expectedError: "user ID is required",
		},
		{
			name: "user not found",
			user: &models.User{
				ID: "nonexistent",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "user not found",
		},
		{
			name: "repository update error",
			user: &models.User{
				ID:       "user-123",
				Username: "updated",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", ctx, "user-123").Return(&models.User{
					ID:           "user-123",
					PasswordHash: "hash",
					CreatedAt:    now,
				}, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewUserService(mockRepo)
			err := service.UpdateUser(ctx, tt.user)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.user)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	ctx := context.Background()

	// Create test passwords and hashes
	oldPassword := "OldPass123!"
	newPassword := "NewPass456!"
	oldHash, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		userID        string
		oldPassword   string
		newPassword   string
		setupMock     func(*MockUserRepository)
		expectedError string
	}{
		{
			name:        "successful password change",
			userID:      "user-123",
			oldPassword: oldPassword,
			newPassword: newPassword,
			setupMock: func(m *MockUserRepository) {
				// GetByID returns user with old password
				m.On("GetByID", ctx, "user-123").Return(&models.User{
					ID:           "user-123",
					Username:     "testuser",
					PasswordHash: string(oldHash),
				}, nil)
				// Update with new password hash
				m.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
					// Verify new password was hashed
					err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(newPassword))
					return u.ID == "user-123" && err == nil
				})).Return(nil)
			},
		},
		{
			name:        "user not found",
			userID:      "nonexistent",
			oldPassword: oldPassword,
			newPassword: newPassword,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "user not found",
		},
		{
			name:        "incorrect old password",
			userID:      "user-123",
			oldPassword: "wrongpassword",
			newPassword: newPassword,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", ctx, "user-123").Return(&models.User{
					ID:           "user-123",
					PasswordHash: string(oldHash),
				}, nil)
			},
			expectedError: "invalid password",
		},
		{
			name:        "new password too short",
			userID:      "user-123",
			oldPassword: oldPassword,
			newPassword: "short",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", ctx, "user-123").Return(&models.User{
					ID:           "user-123",
					PasswordHash: string(oldHash),
				}, nil)
			},
			expectedError: "password must be at least 8 characters long",
		},
		{
			name:        "update error",
			userID:      "user-123",
			oldPassword: oldPassword,
			newPassword: newPassword,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByID", ctx, "user-123").Return(&models.User{
					ID:           "user-123",
					PasswordHash: string(oldHash),
				}, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewUserService(mockRepo)
			err := service.ChangePassword(ctx, tt.userID, tt.oldPassword, tt.newPassword)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        string
		setupMock     func(*MockUserRepository)
		expectedError string
	}{
		{
			name:   "successful deletion",
			userID: "user-123",
			setupMock: func(m *MockUserRepository) {
				m.On("Delete", ctx, "user-123").Return(nil)
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(m *MockUserRepository) {
				m.On("Delete", ctx, "nonexistent").Return(errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:   "repository error",
			userID: "user-123",
			setupMock: func(m *MockUserRepository) {
				m.On("Delete", ctx, "user-123").Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:   "empty user ID",
			userID: "",
			setupMock: func(m *MockUserRepository) {
				m.On("Delete", ctx, "").Return(errors.New("invalid ID"))
			},
			expectedError: "invalid ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewUserService(mockRepo)
			err := service.DeleteUser(ctx, tt.userID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetByUsername(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		username      string
		setupMock     func(*MockUserRepository)
		expected      *models.User
		expectedError string
	}{
		{
			name:     "successful retrieval",
			username: "testuser",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "testuser").Return(&models.User{
					ID:       "user-123",
					Username: "testuser",
					Email:    "test@example.com",
				}, nil)
			},
			expected: &models.User{
				ID:       "user-123",
				Username: "testuser",
				Email:    "test@example.com",
			},
		},
		{
			name:     "user not found",
			username: "nonexistent",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:     "empty username",
			username: "",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "").Return(nil, errors.New("invalid username"))
			},
			expectedError: "invalid username",
		},
		{
			name:     "case sensitivity check",
			username: "TestUser",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByUsername", ctx, "TestUser").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewUserService(mockRepo)
			user, err := service.GetByUsername(ctx, tt.username)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.expected.ID, user.ID)
				assert.Equal(t, tt.expected.Username, user.Username)
				assert.Equal(t, tt.expected.Email, user.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test concurrent operations
func TestUserService_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Set up expectations for concurrent calls
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("user-%d", i)
		mockRepo.On("GetByID", ctx, userID).Return(&models.User{
			ID:       userID,
			Username: fmt.Sprintf("user%d", i),
		}, nil).Maybe()
	}

	// Run concurrent GetByID operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			userID := fmt.Sprintf("user-%d", id)
			user, err := service.GetByID(ctx, userID)
			assert.NoError(t, err)
			assert.Equal(t, userID, user.ID)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	mockRepo.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkUserService_Register(b *testing.B) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Set up mock to always succeed
	mockRepo.On("GetByUsername", ctx, mock.Anything).Return(nil, errors.New("not found"))
	mockRepo.On("GetByEmail", ctx, mock.Anything).Return(nil, errors.New("not found"))
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := models.RegisterRequest{
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "SecurePass123!",
		}
		_, _ = service.Register(ctx, req)
	}
}

func BenchmarkUserService_ChangePassword(b *testing.B) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	oldHash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123!"), bcrypt.DefaultCost)

	// Set up mock
	mockRepo.On("GetByID", ctx, "user-123").Return(&models.User{
		ID:           "user-123",
		PasswordHash: string(oldHash),
	}, nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.ChangePassword(ctx, "user-123", "OldPass123!", "NewPass456!")
	}
}