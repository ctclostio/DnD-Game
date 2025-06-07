package database

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestUserRepository_Create(t *testing.T) {
	cases := []testutil.DBTestCase{
		{
			Name: "successful user creation",
			Setup: func(mock sqlmock.Sqlmock) {
				user := testutil.NewUserBuilder().Build()
				
				mock.ExpectQuery(
					`INSERT INTO users \(username, email, password_hash\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`,
				).WithArgs(
					user.Username, user.Email, user.PasswordHash,
				).WillReturnRows(
					sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(1, time.Now(), time.Now()),
				)
			},
			Run: func(db *sqlx.DB) error {
				repo := NewUserRepository(db)
				user := testutil.NewUserBuilder().Build()
				return repo.Create(user)
			},
			Assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			Name: "duplicate username",
			Setup: func(mock sqlmock.Sqlmock) {
				user := testutil.NewUserBuilder().Build()
				
				mock.ExpectQuery(
					`INSERT INTO users \(username, email, password_hash\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`,
				).WithArgs(
					user.Username, user.Email, user.PasswordHash,
				).WillReturnError(sql.ErrNoRows) // Simulate unique constraint violation
			},
			Run: func(db *sqlx.DB) error {
				repo := NewUserRepository(db)
				user := testutil.NewUserBuilder().Build()
				return repo.Create(user)
			},
			Assert: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			Name: "duplicate email",
			Setup: func(mock sqlmock.Sqlmock) {
				user := testutil.NewUserBuilder().
					WithUsername("newuser").
					WithEmail("existing@example.com").
					Build()
				
				mock.ExpectQuery(
					`INSERT INTO users \(username, email, password_hash\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`,
				).WithArgs(
					user.Username, user.Email, user.PasswordHash,
				).WillReturnError(sql.ErrNoRows) // Simulate unique constraint violation
			},
			Run: func(db *sqlx.DB) error {
				repo := NewUserRepository(db)
				user := testutil.NewUserBuilder().
					WithUsername("newuser").
					WithEmail("existing@example.com").
					Build()
				return repo.Create(user)
			},
			Assert: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	testutil.RunDBTestCases(t, cases)
}

func TestUserRepository_GetByID(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewUserRepository(mockDB.DB)

	t.Run("successful retrieval", func(t *testing.T) {
		user := testutil.NewUserBuilder().WithID(42).Build()
		
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "created_at", "updated_at",
		}).AddRow(
			user.ID, user.Username, user.Email, user.PasswordHash,
			user.CreatedAt, user.UpdatedAt,
		)

		mockDB.Mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = \$1`,
		).WithArgs(int64(42)).WillReturnRows(rows)

		result, err := repo.GetByID(42)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, int64(42), result.ID)
		require.Equal(t, "testuser", result.Username)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockDB.Mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = \$1`,
		).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)

		result, err := repo.GetByID(999)
		require.Error(t, err)
		require.Nil(t, result)
		
		mockDB.AssertExpectations(t)
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewUserRepository(mockDB.DB)

	t.Run("successful retrieval", func(t *testing.T) {
		user := testutil.NewUserBuilder().
			WithUsername("aragorn").
			Build()
		
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "created_at", "updated_at",
		}).AddRow(
			user.ID, user.Username, user.Email, user.PasswordHash,
			user.CreatedAt, user.UpdatedAt,
		)

		mockDB.Mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = \$1`,
		).WithArgs("aragorn").WillReturnRows(rows)

		result, err := repo.GetByUsername("aragorn")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "aragorn", result.Username)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("case insensitive search", func(t *testing.T) {
		user := testutil.NewUserBuilder().
			WithUsername("aragorn").
			Build()
		
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "created_at", "updated_at",
		}).AddRow(
			user.ID, user.Username, user.Email, user.PasswordHash,
			user.CreatedAt, user.UpdatedAt,
		)

		// Should use LOWER() for case-insensitive search
		mockDB.Mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE LOWER\(username\) = LOWER\(\$1\)`,
		).WithArgs("ARAGORN").WillReturnRows(rows)

		result, err := repo.GetByUsername("ARAGORN")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "aragorn", result.Username)
		
		mockDB.AssertExpectations(t)
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewUserRepository(mockDB.DB)

	t.Run("successful retrieval", func(t *testing.T) {
		user := testutil.NewUserBuilder().
			WithEmail("aragorn@gondor.com").
			Build()
		
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "created_at", "updated_at",
		}).AddRow(
			user.ID, user.Username, user.Email, user.PasswordHash,
			user.CreatedAt, user.UpdatedAt,
		)

		mockDB.Mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = \$1`,
		).WithArgs("aragorn@gondor.com").WillReturnRows(rows)

		result, err := repo.GetByEmail("aragorn@gondor.com")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "aragorn@gondor.com", result.Email)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("email normalization", func(t *testing.T) {
		user := testutil.NewUserBuilder().
			WithEmail("test@example.com").
			Build()
		
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "created_at", "updated_at",
		}).AddRow(
			user.ID, user.Username, user.Email, user.PasswordHash,
			user.CreatedAt, user.UpdatedAt,
		)

		// Should normalize email to lowercase
		mockDB.Mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE LOWER\(email\) = LOWER\(\$1\)`,
		).WithArgs("TEST@EXAMPLE.COM").WillReturnRows(rows)

		result, err := repo.GetByEmail("TEST@EXAMPLE.COM")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "test@example.com", result.Email)
		
		mockDB.AssertExpectations(t)
	})
}

func TestUserRepository_Update(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewUserRepository(mockDB.DB)

	t.Run("successful update", func(t *testing.T) {
		user := testutil.NewUserBuilder().
			WithID(1).
			WithEmail("newemail@example.com").
			Build()

		mockDB.Mock.ExpectExec(
			`UPDATE users SET username = \$2, email = \$3, password_hash = \$4, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(
			user.ID, user.Username, user.Email, user.PasswordHash,
		).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(user)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		user := testutil.NewUserBuilder().WithID(999).Build()

		mockDB.Mock.ExpectExec(
			`UPDATE users SET username = \$2, email = \$3, password_hash = \$4, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(
			user.ID, user.Username, user.Email, user.PasswordHash,
		).WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(user)
		require.Error(t, err)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("update violates unique constraint", func(t *testing.T) {
		user := testutil.NewUserBuilder().
			WithID(1).
			WithUsername("existing_user").
			Build()

		mockDB.Mock.ExpectExec(
			`UPDATE users SET username = \$2, email = \$3, password_hash = \$4, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(
			user.ID, user.Username, user.Email, user.PasswordHash,
		).WillReturnError(sql.ErrNoRows) // Simulate constraint violation

		err := repo.Update(user)
		require.Error(t, err)
		
		mockDB.AssertExpectations(t)
	})
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewUserRepository(mockDB.DB)

	t.Run("successful password update", func(t *testing.T) {
		newPasswordHash := "$2a$10$newhashedpassword"

		mockDB.Mock.ExpectExec(
			`UPDATE users SET password_hash = \$2, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(int64(1), newPasswordHash).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdatePassword(1, newPasswordHash)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("empty password hash rejected", func(t *testing.T) {
		err := repo.UpdatePassword(1, "")
		require.Error(t, err)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewUserRepository(mockDB.DB)

	t.Run("successful soft delete", func(t *testing.T) {
		// Assuming soft delete implementation
		mockDB.Mock.ExpectExec(
			`UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(1)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("hard delete", func(t *testing.T) {
		mockDB.Mock.ExpectExec(
			`DELETE FROM users WHERE id = \$1`,
		).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.HardDelete(1)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})
}

func TestUserRepository_ExistsByUsername(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewUserRepository(mockDB.DB)

	t.Run("username exists", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mockDB.Mock.ExpectQuery(
			`SELECT EXISTS\(SELECT 1 FROM users WHERE LOWER\(username\) = LOWER\(\$1\)\)`,
		).WithArgs("testuser").WillReturnRows(rows)

		exists, err := repo.ExistsByUsername("testuser")
		require.NoError(t, err)
		require.True(t, exists)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("username does not exist", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

		mockDB.Mock.ExpectQuery(
			`SELECT EXISTS\(SELECT 1 FROM users WHERE LOWER\(username\) = LOWER\(\$1\)\)`,
		).WithArgs("nonexistent").WillReturnRows(rows)

		exists, err := repo.ExistsByUsername("nonexistent")
		require.NoError(t, err)
		require.False(t, exists)
		
		mockDB.AssertExpectations(t)
	})
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewUserRepository(mockDB.DB)

	t.Run("email exists", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mockDB.Mock.ExpectQuery(
			`SELECT EXISTS\(SELECT 1 FROM users WHERE LOWER\(email\) = LOWER\(\$1\)\)`,
		).WithArgs("test@example.com").WillReturnRows(rows)

		exists, err := repo.ExistsByEmail("test@example.com")
		require.NoError(t, err)
		require.True(t, exists)
		
		mockDB.AssertExpectations(t)
	})
}

func TestUserRepository_BulkOperations(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewUserRepository(mockDB.DB)

	t.Run("get multiple users by IDs", func(t *testing.T) {
		userIDs := []int64{1, 2, 3}
		
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "created_at", "updated_at",
		}).
			AddRow(1, "user1", "user1@example.com", "hash1", time.Now(), time.Now()).
			AddRow(2, "user2", "user2@example.com", "hash2", time.Now(), time.Now()).
			AddRow(3, "user3", "user3@example.com", "hash3", time.Now(), time.Now())

		mockDB.Mock.ExpectQuery(
			`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id IN \(\$1, \$2, \$3\)`,
		).WithArgs(int64(1), int64(2), int64(3)).WillReturnRows(rows)

		users, err := repo.GetByIDs(userIDs)
		require.NoError(t, err)
		require.Len(t, users, 3)
		
		mockDB.AssertExpectations(t)
	})
}