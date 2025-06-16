package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/services/mocks"
)

// Test constants to avoid hardcoded credentials
const (
	testUserPassword    = "SecurePass123!"
	testPasswordValid   = "TestPass123!"
	testPasswordWeak    = "weak"
	testPasswordOld     = "OldPass123!"
	testPasswordNew     = "NewPass456!"
)

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       models.RegisterRequest
		setupMock     func(*mocks.MockUserRepository)
		expectedError string
		validate      func(*testing.T, *models.User)
	}{
		{
			name: "successful registration",
			request: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: testPasswordValid,
			},
			setupMock: func(m *mocks.MockUserRepository) {
				// Check that username doesn't exist
				m.On("GetByUsername", ctx, "testuser").Return(nil, nil)
				// Check that email doesn't exist
				m.On("GetByEmail", ctx, "test@example.com").Return(nil, nil)
				// Create user
				m.On("Create", ctx, mock.MatchedBy(func(u *models.User) bool {
					return u.Username == "testuser" &&
						u.Email == "test@example.com" &&
						u.PasswordHash != "" &&
						u.ID == "" && // ID should be empty, repository will set it
						u.Role == "" // Role should be empty, repository will set it
				})).Return(nil).Run(func(args mock.Arguments) {
					// Simulate repository setting the ID and role
					user := args.Get(1).(*models.User)
					user.ID = "generated-id"
					user.Role = "player"
				})
			},
			validate: func(t *testing.T, user *models.User) {
				assert.NotEmpty(t, user.ID)
				assert.Equal(t, "testuser", user.Username)
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "player", user.Role)
				assert.NotEmpty(t, user.PasswordHash)
				assert.NotEqual(t, testUserPassword, user.PasswordHash) // Should be hashed

				// Verify password was hashed correctly
				err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(testUserPassword))
				assert.NoError(t, err)
			},
		},
		{
			name: "username already exists",
			request: models.RegisterRequest{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: testUserPassword,
			},
			setupMock: func(m *mocks.MockUserRepository) {
				existingUser := &models.User{
					ID:       "123",
					Username: "existinguser",
					Email:    "existing@example.com",
				}
				m.On("GetByUsername", ctx, "existinguser").Return(existingUser, nil)
			},
			expectedError: "username already taken",
		},
		{
			name: "email already exists",
			request: models.RegisterRequest{
				Username: "newuser",
				Email:    "existing@example.com",
				Password: testPasswordValid,
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByUsername", ctx, "newuser").Return(nil, nil)
				existingUser := &models.User{
					ID:    "123",
					Email: "existing@example.com",
				}
				m.On("GetByEmail", ctx, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: "email already registered",
		},
		{
			name: "weak password",
			request: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "weak",
			},
			expectedError: "password must be at least 8 characters",
		},
		{
			name: "empty username",
			request: models.RegisterRequest{
				Username: "",
				Email:    "test@example.com",
				Password: testPasswordValid,
			},
			expectedError: "username is required",
		},
		{
			name: "empty email",
			request: models.RegisterRequest{
				Username: "testuser",
				Email:    "",
				Password: testPasswordValid,
			},
			expectedError: "email is required",
		},
		{
			name: "repository error on create",
			request: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: testPasswordValid,
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByUsername", ctx, "testuser").Return(nil, nil)
				m.On("GetByEmail", ctx, "test@example.com").Return(nil, nil)
				m.On("Create", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewUserService(mockRepo)
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
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPasswordValid), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		request       models.LoginRequest
		setupMock     func(*mocks.MockUserRepository)
		expectedError string
		validate      func(*testing.T, *models.AuthResponse)
	}{
		{
			name: "successful login with username",
			request: models.LoginRequest{
				Username: "testuser",
				Password: testPasswordValid,
			},
			setupMock: func(m *mocks.MockUserRepository) {
				user := &models.User{
					ID:           "user-123",
					Username:     "testuser",
					Email:        "test@example.com",
					Role:         "player",
					PasswordHash: string(hashedPassword),
				}
				m.On("GetByUsername", ctx, "testuser").Return(user, nil)
			},
			validate: func(t *testing.T, auth *models.AuthResponse) {
				assert.Equal(t, "user-123", auth.User.ID)
				assert.Equal(t, "testuser", auth.User.Username)
			},
		},
		{
			name: "incorrect password",
			request: models.LoginRequest{
				Username: "testuser",
				Password: testPasswordWeak,
			},
			setupMock: func(m *mocks.MockUserRepository) {
				user := &models.User{
					ID:           "user-123",
					Username:     "testuser",
					PasswordHash: string(hashedPassword),
				}
				m.On("GetByUsername", ctx, "testuser").Return(user, nil)
			},
			expectedError: "invalid username or password",
		},
		{
			name: "user not found",
			request: models.LoginRequest{
				Username: "nonexistent",
				Password: testPasswordValid,
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByUsername", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "invalid username or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewUserService(mockRepo)
			auth, err := service.Login(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, auth)
			} else {
				require.NoError(t, err)
				require.NotNil(t, auth)
				if tt.validate != nil {
					tt.validate(t, auth)
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
		setupMock     func(*mocks.MockUserRepository)
		expectedError string
		validate      func(*testing.T, *models.User)
	}{
		{
			name:   "successful get user",
			userID: "user-123",
			setupMock: func(m *mocks.MockUserRepository) {
				user := &models.User{
					ID:       "user-123",
					Username: "testuser",
					Email:    "test@example.com",
				}
				m.On("GetByID", ctx, "user-123").Return(user, nil)
			},
			validate: func(t *testing.T, user *models.User) {
				assert.Equal(t, "user-123", user.ID)
				assert.Equal(t, "testuser", user.Username)
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:          "empty user ID",
			userID:        "",
			expectedError: "user ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewUserService(mockRepo)
			user, err := service.GetUserByID(ctx, tt.userID)

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

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		user          *models.User
		setupMock     func(*mocks.MockUserRepository)
		expectedError string
	}{
		{
			name: "successful update",
			user: &models.User{
				ID:       "user-123",
				Username: "newusername",
				Email:    "newemail@example.com",
			},
			setupMock: func(m *mocks.MockUserRepository) {
				existingUser := &models.User{
					ID:           "user-123",
					Username:     "oldusername",
					Email:        "oldemail@example.com",
					PasswordHash: "hashedpassword",
					CreatedAt:    time.Now(),
				}
				m.On("GetByID", ctx, "user-123").Return(existingUser, nil)
				m.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
					return u.ID == "user-123" &&
						u.Username == "newusername" &&
						u.Email == "newemail@example.com" &&
						u.PasswordHash == "hashedpassword" // Password preserved
				})).Return(nil)
			},
		},
		{
			name: "user not found",
			user: &models.User{
				ID:       "nonexistent",
				Username: "testuser",
			},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "user not found",
		},
		{
			name: "empty user ID",
			user: &models.User{
				ID:       "",
				Username: "testuser",
			},
			expectedError: "user ID is required",
		},
		{
			name: "repository error on update",
			user: &models.User{
				ID:       "user-123",
				Username: "newusername",
			},
			setupMock: func(m *mocks.MockUserRepository) {
				existingUser := &models.User{
					ID:           "user-123",
					PasswordHash: "hashedpassword",
					CreatedAt:    time.Now(),
				}
				m.On("GetByID", ctx, "user-123").Return(existingUser, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewUserService(mockRepo)
			err := service.UpdateUser(ctx, tt.user)

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

func TestUserService_ChangePassword(t *testing.T) {
	ctx := context.Background()

	oldHash, _ := bcrypt.GenerateFromPassword([]byte(testPasswordOld), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		userID        string
		oldPassword   string
		newPassword   string
		setupMock     func(*mocks.MockUserRepository)
		expectedError string
	}{
		{
			name:        "successful password change",
			userID:      "user-123",
			oldPassword: testPasswordOld,
			newPassword: testPasswordNew,
			setupMock: func(m *mocks.MockUserRepository) {
				user := &models.User{
					ID:           "user-123",
					Username:     "testuser",
					PasswordHash: string(oldHash),
				}
				m.On("GetByID", ctx, "user-123").Return(user, nil)
				m.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
					// Verify new password hash is different
					return u.ID == "user-123" && u.PasswordHash != string(oldHash)
				})).Return(nil)
			},
		},
		{
			name:        "incorrect current password",
			userID:      "user-123",
			oldPassword: testPasswordWeak,
			newPassword: testPasswordNew,
			setupMock: func(m *mocks.MockUserRepository) {
				user := &models.User{
					ID:           "user-123",
					Username:     "testuser",
					PasswordHash: string(oldHash),
				}
				m.On("GetByID", ctx, "user-123").Return(user, nil)
			},
			expectedError: "invalid password",
		},
		{
			name:        "weak new password",
			userID:      "user-123",
			oldPassword: testPasswordOld,
			newPassword: "weak",
			setupMock: func(m *mocks.MockUserRepository) {
				user := &models.User{
					ID:           "user-123",
					Username:     "testuser",
					PasswordHash: string(oldHash),
				}
				m.On("GetByID", ctx, "user-123").Return(user, nil)
			},
			expectedError: "password must be at least 8 characters",
		},
		{
			name:        "user not found",
			userID:      "nonexistent",
			oldPassword: testPasswordOld,
			newPassword: testPasswordNew,
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewUserService(mockRepo)
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
		setupMock     func(*mocks.MockUserRepository)
		expectedError string
	}{
		{
			name:   "successful deletion",
			userID: "user-123",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Delete", ctx, "user-123").Return(nil)
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("Delete", ctx, "nonexistent").Return(errors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:          "empty user ID",
			userID:        "",
			expectedError: "user ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewUserService(mockRepo)
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
