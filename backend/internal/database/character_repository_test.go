package database

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/your-org/dnd-game/internal/models"
	"github.com/your-org/dnd-game/internal/testutil"
)

func TestCharacterRepository_Create(t *testing.T) {
	cases := []testutil.DBTestCase{
		{
			Name: "successful character creation",
			Setup: func(mock sqlmock.Sqlmock) {
				char := testutil.NewCharacterBuilder().Build()
				
				mock.ExpectQuery(
					`INSERT INTO characters \(user_id, name, race, class, level, experience_points, 
					hit_points, max_hit_points, armor_class, initiative, speed, abilities, 
					skills, proficiencies, equipment, spell_slots, known_spells, prepared_spells\) 
					VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, \$11, \$12, 
					\$13, \$14, \$15, \$16, \$17, \$18\) 
					RETURNING id, created_at, updated_at`,
				).WithArgs(
					char.UserID, char.Name, char.Race, char.Class, char.Level,
					char.ExperiencePoints, char.HitPoints, char.MaxHitPoints,
					char.ArmorClass, char.Initiative, char.Speed,
					sqlmock.AnyArg(), // abilities JSON
					sqlmock.AnyArg(), // skills JSON
					sqlmock.AnyArg(), // proficiencies JSON
					sqlmock.AnyArg(), // equipment JSON
					sqlmock.AnyArg(), // spell_slots JSON
					sqlmock.AnyArg(), // known_spells JSON
					sqlmock.AnyArg(), // prepared_spells JSON
				).WillReturnRows(
					sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(1, time.Now(), time.Now()),
				)
			},
			Run: func(db *sqlx.DB) error {
				repo := NewCharacterRepository(db)
				char := testutil.NewCharacterBuilder().Build()
				return repo.Create(char)
			},
			Assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			Name: "duplicate character name for user",
			Setup: func(mock sqlmock.Sqlmock) {
				char := testutil.NewCharacterBuilder().Build()
				
				mock.ExpectQuery(
					`INSERT INTO characters.*`,
				).WithArgs(
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				).WillReturnError(sql.ErrNoRows) // Simulate constraint violation
			},
			Run: func(db *sqlx.DB) error {
				repo := NewCharacterRepository(db)
				char := testutil.NewCharacterBuilder().Build()
				return repo.Create(char)
			},
			Assert: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	testutil.RunDBTestCases(t, cases)
}

func TestCharacterRepository_GetByID(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewCharacterRepository(mockDB.DB)
	char := testutil.NewCharacterBuilder().WithID(42).Build()

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "user_id", "name", "race", "class", "level",
			"experience_points", "hit_points", "max_hit_points",
			"armor_class", "initiative", "speed", "abilities",
			"skills", "proficiencies", "equipment", "spell_slots",
			"known_spells", "prepared_spells", "created_at", "updated_at",
		}).AddRow(
			char.ID, char.UserID, char.Name, char.Race, char.Class, char.Level,
			char.ExperiencePoints, char.HitPoints, char.MaxHitPoints,
			char.ArmorClass, char.Initiative, char.Speed,
			`{"strength":10,"dexterity":14,"constitution":12,"intelligence":16,"wisdom":14,"charisma":12}`,
			`{"Arcana":5,"Investigation":5}`,
			`["Light Armor","Simple Weapons"]`,
			`["Spellbook","Component Pouch","Staff"]`,
			`{"1":{"total":2,"used":0}}`,
			`["Mage Hand","Fire Bolt","Shield","Magic Missile"]`,
			`["Shield","Magic Missile"]`,
			char.CreatedAt, char.UpdatedAt,
		)

		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM characters WHERE id = \$1`,
		).WithArgs(42).WillReturnRows(rows)

		result, err := repo.GetByID(42)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, int64(42), result.ID)
		require.Equal(t, "Gandalf", result.Name)
		testutil.RequireValidCharacter(t, result)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("character not found", func(t *testing.T) {
		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM characters WHERE id = \$1`,
		).WithArgs(999).WillReturnError(sql.ErrNoRows)

		result, err := repo.GetByID(999)
		require.Error(t, err)
		require.Nil(t, result)
		
		mockDB.AssertExpectations(t)
	})
}

func TestCharacterRepository_GetByUserID(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewCharacterRepository(mockDB.DB)

	t.Run("multiple characters found", func(t *testing.T) {
		char1 := testutil.NewCharacterBuilder().
			WithID(1).
			WithName("Aragorn").
			WithClass("Fighter").
			Build()
		
		char2 := testutil.NewCharacterBuilder().
			WithID(2).
			WithName("Legolas").
			WithClass("Ranger").
			Build()

		rows := sqlmock.NewRows([]string{
			"id", "user_id", "name", "race", "class", "level",
			"experience_points", "hit_points", "max_hit_points",
			"armor_class", "initiative", "speed", "abilities",
			"skills", "proficiencies", "equipment", "spell_slots",
			"known_spells", "prepared_spells", "created_at", "updated_at",
		}).AddRow(
			char1.ID, char1.UserID, char1.Name, char1.Race, char1.Class, char1.Level,
			char1.ExperiencePoints, char1.HitPoints, char1.MaxHitPoints,
			char1.ArmorClass, char1.Initiative, char1.Speed,
			`{"strength":10,"dexterity":14,"constitution":12,"intelligence":16,"wisdom":14,"charisma":12}`,
			`{}`, `[]`, `[]`, `{}`, `[]`, `[]`,
			char1.CreatedAt, char1.UpdatedAt,
		).AddRow(
			char2.ID, char2.UserID, char2.Name, char2.Race, char2.Class, char2.Level,
			char2.ExperiencePoints, char2.HitPoints, char2.MaxHitPoints,
			char2.ArmorClass, char2.Initiative, char2.Speed,
			`{"strength":10,"dexterity":14,"constitution":12,"intelligence":16,"wisdom":14,"charisma":12}`,
			`{}`, `[]`, `[]`, `{}`, `[]`, `[]`,
			char2.CreatedAt, char2.UpdatedAt,
		)

		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM characters WHERE user_id = \$1 ORDER BY created_at DESC`,
		).WithArgs(int64(1)).WillReturnRows(rows)

		results, err := repo.GetByUserID(1)
		require.NoError(t, err)
		require.Len(t, results, 2)
		require.Equal(t, "Aragorn", results[0].Name)
		require.Equal(t, "Legolas", results[1].Name)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("no characters found", func(t *testing.T) {
		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM characters WHERE user_id = \$1 ORDER BY created_at DESC`,
		).WithArgs(int64(999)).WillReturnRows(
			sqlmock.NewRows([]string{"id"}), // Empty result set
		)

		results, err := repo.GetByUserID(999)
		require.NoError(t, err)
		require.Empty(t, results)
		
		mockDB.AssertExpectations(t)
	})
}

func TestCharacterRepository_Update(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewCharacterRepository(mockDB.DB)

	t.Run("successful update", func(t *testing.T) {
		char := testutil.NewCharacterBuilder().
			WithID(1).
			WithLevel(6).
			WithHP(50, 55).
			Build()

		mockDB.Mock.ExpectExec(
			`UPDATE characters SET name = \$2, race = \$3, class = \$4, level = \$5, 
			experience_points = \$6, hit_points = \$7, max_hit_points = \$8, 
			armor_class = \$9, initiative = \$10, speed = \$11, abilities = \$12, 
			skills = \$13, proficiencies = \$14, equipment = \$15, spell_slots = \$16, 
			known_spells = \$17, prepared_spells = \$18, updated_at = CURRENT_TIMESTAMP 
			WHERE id = \$1`,
		).WithArgs(
			char.ID, char.Name, char.Race, char.Class, char.Level,
			char.ExperiencePoints, char.HitPoints, char.MaxHitPoints,
			char.ArmorClass, char.Initiative, char.Speed,
			sqlmock.AnyArg(), // abilities JSON
			sqlmock.AnyArg(), // skills JSON
			sqlmock.AnyArg(), // proficiencies JSON
			sqlmock.AnyArg(), // equipment JSON
			sqlmock.AnyArg(), // spell_slots JSON
			sqlmock.AnyArg(), // known_spells JSON
			sqlmock.AnyArg(), // prepared_spells JSON
		).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(char)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("character not found", func(t *testing.T) {
		char := testutil.NewCharacterBuilder().WithID(999).Build()

		mockDB.Mock.ExpectExec(
			`UPDATE characters SET.*WHERE id = \$1`,
		).WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(char)
		require.Error(t, err)
		
		mockDB.AssertExpectations(t)
	})
}

func TestCharacterRepository_Delete(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewCharacterRepository(mockDB.DB)

	t.Run("successful deletion", func(t *testing.T) {
		mockDB.Mock.ExpectExec(
			`DELETE FROM characters WHERE id = \$1`,
		).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(1)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("character not found", func(t *testing.T) {
		mockDB.Mock.ExpectExec(
			`DELETE FROM characters WHERE id = \$1`,
		).WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(999)
		require.Error(t, err)
		
		mockDB.AssertExpectations(t)
	})
}

func TestCharacterRepository_UpdateHP(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewCharacterRepository(mockDB.DB)

	t.Run("successful HP update", func(t *testing.T) {
		mockDB.Mock.ExpectExec(
			`UPDATE characters SET hit_points = \$2, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(int64(1), 35).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateHP(1, 35)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("HP cannot be negative", func(t *testing.T) {
		// Repository should handle this validation
		err := repo.UpdateHP(1, -10)
		require.Error(t, err)
	})
}

func TestCharacterRepository_TransactionalOperations(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewCharacterRepository(mockDB.DB)

	t.Run("create multiple characters in transaction", func(t *testing.T) {
		char1 := testutil.NewCharacterBuilder().WithName("Frodo").Build()
		char2 := testutil.NewCharacterBuilder().WithName("Sam").Build()

		// Set up transaction expectations
		mockDB.ExpectBegin()
		
		// First character
		mockDB.Mock.ExpectQuery(
			`INSERT INTO characters.*RETURNING id, created_at, updated_at`,
		).WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).WillReturnRows(
			sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(1, time.Now(), time.Now()),
		)
		
		// Second character
		mockDB.Mock.ExpectQuery(
			`INSERT INTO characters.*RETURNING id, created_at, updated_at`,
		).WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).WillReturnRows(
			sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(2, time.Now(), time.Now()),
		)
		
		mockDB.ExpectCommit()

		// Execute transaction
		tx, err := mockDB.DB.Beginx()
		require.NoError(t, err)
		
		repoTx := NewCharacterRepository(tx)
		err = repoTx.Create(char1)
		require.NoError(t, err)
		
		err = repoTx.Create(char2)
		require.NoError(t, err)
		
		err = tx.Commit()
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})
}