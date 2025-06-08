package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run("successful user creation", func(t *testing.T) {
		user := &models.User{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$10$hashedpassword",
		}

		mock.ExpectQuery(
			`INSERT INTO users \(username, email, password_hash\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`,
		).WithArgs(
			user.Username, user.Email, user.PasswordHash,
		).WillReturnRows(
			sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow("user-123", time.Now(), time.Now()),
		)

		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
		assert.Equal(t, "user-123", user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate username", func(t *testing.T) {
		user := &models.User{
			Username:     "existing",
			Email:        "new@example.com",
			PasswordHash: "$2a$10$hashedpassword",
		}

		mock.ExpectQuery(
			`INSERT INTO users \(username, email, password_hash\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`,
		).WithArgs(
			user.Username, user.Email, user.PasswordHash,
		).WillReturnError(sql.ErrNoRows) // Simulate unique constraint violation

		err := repo.Create(context.Background(), user)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate email", func(t *testing.T) {
		user := &models.User{
			Username:     "newuser",
			Email:        "existing@example.com",
			PasswordHash: "$2a$10$hashedpassword",
		}

		mock.ExpectQuery(
			`INSERT INTO users \(username, email, password_hash\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`,
		).WithArgs(
			user.Username, user.Email, user.PasswordHash,
		).WillReturnError(sql.ErrNoRows) // Simulate unique constraint violation

		err := repo.Create(context.Background(), user)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run("successful retrieval", func(t *testing.T) {
		expectedUser := &models.User{
			ID:           "user-42",
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$10$hashedpassword",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "created_at", "updated_at",
		}).AddRow(
			expectedUser.ID, expectedUser.Username, expectedUser.Email, 
			expectedUser.PasswordHash, expectedUser.CreatedAt, expectedUser.UpdatedAt,
		)

		mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = \$1`,
		).WithArgs("user-42").WillReturnRows(rows)

		user, err := repo.GetByID(context.Background(), "user-42")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user-42", user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = \$1`,
		).WithArgs("non-existent").WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByID(context.Background(), "non-existent")
		assert.Error(t, err)
		assert.Equal(t, models.ErrUserNotFound, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run("successful retrieval", func(t *testing.T) {
		expectedUser := &models.User{
			ID:           "user-123",
			Username:     "aragorn",
			Email:        "aragorn@gondor.com",
			PasswordHash: "$2a$10$hashedpassword",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "created_at", "updated_at",
		}).AddRow(
			expectedUser.ID, expectedUser.Username, expectedUser.Email,
			expectedUser.PasswordHash, expectedUser.CreatedAt, expectedUser.UpdatedAt,
		)

		mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = \$1`,
		).WithArgs("aragorn").WillReturnRows(rows)

		user, err := repo.GetByUsername(context.Background(), "aragorn")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "aragorn", user.Username)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = \$1`,
		).WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByUsername(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, models.ErrUserNotFound, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run("successful retrieval", func(t *testing.T) {
		expectedUser := &models.User{
			ID:           "user-456",
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$10$hashedpassword",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "created_at", "updated_at",
		}).AddRow(
			expectedUser.ID, expectedUser.Username, expectedUser.Email,
			expectedUser.PasswordHash, expectedUser.CreatedAt, expectedUser.UpdatedAt,
		)

		mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = \$1`,
		).WithArgs("test@example.com").WillReturnRows(rows)

		user, err := repo.GetByEmail(context.Background(), "test@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run("successful update", func(t *testing.T) {
		user := &models.User{
			ID:           "user-123",
			Username:     "updateduser",
			Email:        "updated@example.com",
			PasswordHash: "$2a$10$newhashedpassword",
		}

		mock.ExpectExec(
			`UPDATE users SET username = \$2, email = \$3, password_hash = \$4, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(
			user.ID, user.Username, user.Email, user.PasswordHash,
		).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		user := &models.User{
			ID:           "non-existent",
			Username:     "updateduser",
			Email:        "updated@example.com",
			PasswordHash: "$2a$10$newhashedpassword",
		}

		mock.ExpectExec(
			`UPDATE users SET username = \$2, email = \$3, password_hash = \$4, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(
			user.ID, user.Username, user.Email, user.PasswordHash,
		).WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(context.Background(), user)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run("successful delete", func(t *testing.T) {
		mock.ExpectExec(
			`DELETE FROM users WHERE id = \$1`,
		).WithArgs("user-123").WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), "user-123")
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectExec(
			`DELETE FROM users WHERE id = \$1`,
		).WithArgs("non-existent").WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(context.Background(), "non-existent")
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}