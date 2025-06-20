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

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
)

const (
	testNameSuccessfulRetrieval = "successful retrieval"
	testNameUserNotFound        = "user not found"
)

func TestUserRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run("successful user creation", func(t *testing.T) {
		user := &models.User{
			Username:     "testuser",
			Email:        testutil.TestEmail,
			PasswordHash: testutil.TestPasswordHash,
		}

		mock.ExpectQuery(
			`INSERT INTO users \(username, email, password_hash\) VALUES \(\?, \?, \?\) RETURNING id, created_at, updated_at`,
		).WithArgs(
			user.Username, user.Email, user.PasswordHash,
		).WillReturnRows(
			sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(testutil.TestUserID3, time.Now(), time.Now()),
		)

		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
		assert.Equal(t, testutil.TestUserID3, user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate username", func(t *testing.T) {
		user := &models.User{
			Username:     "existing",
			Email:        "new@example.com",
			PasswordHash: testutil.TestPasswordHash,
		}

		mock.ExpectQuery(
			`INSERT INTO users \(username, email, password_hash\) VALUES \(\?, \?, \?\) RETURNING id, created_at, updated_at`,
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
			PasswordHash: testutil.TestPasswordHash,
		}

		mock.ExpectQuery(
			`INSERT INTO users \(username, email, password_hash\) VALUES \(\?, \?, \?\) RETURNING id, created_at, updated_at`,
		).WithArgs(
			user.Username, user.Email, user.PasswordHash,
		).WillReturnError(sql.ErrNoRows) // Simulate unique constraint violation

		err := repo.Create(context.Background(), user)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepositoryGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run(testNameSuccessfulRetrieval, func(t *testing.T) {
		expectedUser := &models.User{
			ID:           testutil.TestUserID2,
			Username:     "testuser",
			Email:        testutil.TestEmail,
			PasswordHash: testutil.TestPasswordHash,
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
			`SELECT id, username, email, password_hash, COALESCE\(role, 'player'\) as role, created_at, updated_at FROM users WHERE id = \?`,
		).WithArgs(testutil.TestUserID2).WillReturnRows(rows)

		user, err := repo.GetByID(context.Background(), testutil.TestUserID2)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testutil.TestUserID2, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run(testNameUserNotFound, func(t *testing.T) {
		mock.ExpectQuery(
			`SELECT id, username, email, password_hash, COALESCE\(role, 'player'\) as role, created_at, updated_at FROM users WHERE id = \?`,
		).WithArgs(testutil.TestNonexistent).WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByID(context.Background(), testutil.TestNonexistent)
		assert.Error(t, err)
		assert.Equal(t, models.ErrUserNotFound, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepositoryGetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run(testNameSuccessfulRetrieval, func(t *testing.T) {
		expectedUser := &models.User{
			ID:           testutil.TestUserID3,
			Username:     "aragorn",
			Email:        "aragorn@gondor.com",
			PasswordHash: testutil.TestPasswordHash,
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
			`SELECT id, username, email, password_hash, COALESCE\(role, 'player'\) as role, created_at, updated_at FROM users WHERE username = \?`,
		).WithArgs("aragorn").WillReturnRows(rows)

		user, err := repo.GetByUsername(context.Background(), "aragorn")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "aragorn", user.Username)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run(testNameUserNotFound, func(t *testing.T) {
		mock.ExpectQuery(
			`SELECT id, username, email, password_hash, COALESCE\(role, 'player'\) as role, created_at, updated_at FROM users WHERE username = \?`,
		).WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByUsername(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, models.ErrUserNotFound, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepositoryGetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run(testNameSuccessfulRetrieval, func(t *testing.T) {
		expectedUser := &models.User{
			ID:           "user-456",
			Username:     "testuser",
			Email:        testutil.TestEmail,
			PasswordHash: testutil.TestPasswordHash,
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
			`SELECT id, username, email, password_hash, COALESCE\(role, 'player'\) as role, created_at, updated_at FROM users WHERE email = \?`,
		).WithArgs(testutil.TestEmail).WillReturnRows(rows)

		user, err := repo.GetByEmail(context.Background(), testutil.TestEmail)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testutil.TestEmail, user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepositoryUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run("successful update", func(t *testing.T) {
		user := &models.User{
			ID:           testutil.TestUserID3,
			Username:     "updateduser",
			Email:        "updated@example.com",
			PasswordHash: "$2a$10$newhashedpassword",
		}

		// The Update method uses QueryRowContext with RETURNING
		mock.ExpectQuery(
			`UPDATE users SET username = \?, email = \?, password_hash = \?, updated_at = CURRENT_TIMESTAMP WHERE id = \? RETURNING updated_at`,
		).WithArgs(
			user.Username, user.Email, user.PasswordHash, user.ID,
		).WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(time.Now()))

		err := repo.Update(context.Background(), user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run(testNameUserNotFound, func(t *testing.T) {
		user := &models.User{
			ID:           testutil.TestNonexistent,
			Username:     "updateduser",
			Email:        "updated@example.com",
			PasswordHash: "$2a$10$newhashedpassword",
		}

		// The Update method uses QueryRowContext with RETURNING
		// For user not found, it should return sql.ErrNoRows
		mock.ExpectQuery(
			`UPDATE users SET username = \?, email = \?, password_hash = \?, updated_at = CURRENT_TIMESTAMP WHERE id = \? RETURNING updated_at`,
		).WithArgs(
			user.Username, user.Email, user.PasswordHash, user.ID,
		).WillReturnError(sql.ErrNoRows)

		err := repo.Update(context.Background(), user)
		assert.Error(t, err)
		assert.Equal(t, models.ErrUserNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepositoryDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewUserRepository(dbWrapper)

	t.Run("successful delete", func(t *testing.T) {
		mock.ExpectExec(
			`DELETE FROM users WHERE id = \?`,
		).WithArgs(testutil.TestUserID3).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), testutil.TestUserID3)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run(testNameUserNotFound, func(t *testing.T) {
		mock.ExpectExec(
			`DELETE FROM users WHERE id = \?`,
		).WithArgs(testutil.TestNonexistent).WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(context.Background(), testutil.TestNonexistent)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
