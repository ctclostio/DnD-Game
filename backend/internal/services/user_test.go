package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/your-username/dnd-game/backend/internal/models"
)

// MockUserRepository is a mock implementation of database.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestNewUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
}

func TestUserService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		req := models.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		// Mock expectations
		mockRepo.On("GetByUsername", ctx, req.Username).Return(nil, sql.ErrNoRows).Once()
		mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, sql.ErrNoRows).Once()
		mockRepo.On("Create", ctx, mock.MatchedBy(func(u *models.User) bool {
			// Verify password was hashed
			err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password))
			return err == nil && u.Username == req.Username && u.Email == req.Email
		})).Return(nil).Once()

		user, err := service.Register(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Username, user.Username)
		assert.Equal(t, req.Email, user.Email)
		assert.NotEmpty(t, user.ID)
		assert.NotEmpty(t, user.PasswordHash)

		mockRepo.AssertExpectations(t)
	})

	t.Run("username already exists", func(t *testing.T) {
		req := models.RegisterRequest{
			Username: "existinguser",
			Email:    "new@example.com",
			Password: "password123",
		}

		existingUser := &models.User{
			ID:       "existing-id",
			Username: req.Username,
		}

		mockRepo.On("GetByUsername", ctx, req.Username).Return(existingUser, nil).Once()

		user, err := service.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "username already taken")

		mockRepo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		req := models.RegisterRequest{
			Username: "newuser",
			Email:    "existing@example.com",
			Password: "password123",
		}

		existingUser := &models.User{
			ID:    "existing-id",
			Email: req.Email,
		}

		mockRepo.On("GetByUsername", ctx, req.Username).Return(nil, sql.ErrNoRows).Once()
		mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil).Once()

		user, err := service.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "email already in use")

		mockRepo.AssertExpectations(t)
	})

	t.Run("missing username", func(t *testing.T) {
		req := models.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		user, err := service.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "username is required")
	})

	t.Run("missing email", func(t *testing.T) {
		req := models.RegisterRequest{
			Username: "testuser",
			Password: "password123",
		}

		user, err := service.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("missing password", func(t *testing.T) {
		req := models.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
		}

		user, err := service.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "password is required")
	})
}

func TestUserService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		user := &models.User{
			ID:           "user-123",
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Role:         "player",
		}

		req := models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}

		mockRepo.On("GetByUsername", ctx, req.Username).Return(user, nil).Once()

		result, err := service.Login(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.ID, result.User.ID)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		req := models.LoginRequest{
			Username: "nonexistent",
			Password: "password123",
		}

		mockRepo.On("GetByUsername", ctx, req.Username).Return(nil, sql.ErrNoRows).Once()

		result, err := service.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid username or password")

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		user := &models.User{
			ID:           "user-123",
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Role:         "player",
		}

		req := models.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		mockRepo.On("GetByUsername", ctx, req.Username).Return(user, nil).Once()

		result, err := service.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid username or password")

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	t.Run("user found", func(t *testing.T) {
		userID := "user-123"
		expectedUser := &models.User{
			ID:       userID,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "player",
		}

		mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil).Once()

		user, err := service.GetByID(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		userID := "nonexistent"

		mockRepo.On("GetByID", ctx, userID).Return(nil, sql.ErrNoRows).Once()

		user, err := service.GetByID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

