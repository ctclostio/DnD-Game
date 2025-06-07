package database

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestInventoryRepository_CreateItem(t *testing.T) {
	cases := []testutil.DBTestCase{
		{
			Name: "successful item creation",
			Setup: func(mock sqlmock.Sqlmock) {
				item := testutil.NewInventoryItemBuilder().Build()
				propertiesJSON, _ := json.Marshal(item.Properties)
				
				mock.ExpectQuery(
					`INSERT INTO inventory_items \(character_id, name, type, quantity, weight, value, properties, equipped, description\) 
					VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9\) 
					RETURNING id, created_at, updated_at`,
				).WithArgs(
					item.CharacterID, item.Name, item.Type, item.Quantity,
					item.Weight, item.Value, propertiesJSON, item.Equipped,
					item.Description,
				).WillReturnRows(
					sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(1, time.Now(), time.Now()),
				)
			},
			Run: func(db *sqlx.DB) error {
				repo := NewInventoryRepository(db)
				item := testutil.NewInventoryItemBuilder().Build()
				return repo.CreateItem(item)
			},
			Assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			Name: "create magic item with complex properties",
			Setup: func(mock sqlmock.Sqlmock) {
				item := testutil.NewInventoryItemBuilder().
					WithName("Flametongue Longsword").
					AsMagicItem("rare").
					Build()
				
				item.Properties["damage"] = "1d8+2d6"
				item.Properties["damageType"] = "slashing + fire"
				item.Properties["bonus"] = "+1"
				
				propertiesJSON, _ := json.Marshal(item.Properties)
				
				mock.ExpectQuery(
					`INSERT INTO inventory_items.*RETURNING id, created_at, updated_at`,
				).WithArgs(
					item.CharacterID, item.Name, item.Type, item.Quantity,
					item.Weight, item.Value, propertiesJSON, item.Equipped,
					item.Description,
				).WillReturnRows(
					sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(1, time.Now(), time.Now()),
				)
			},
			Run: func(db *sqlx.DB) error {
				repo := NewInventoryRepository(db)
				item := testutil.NewInventoryItemBuilder().
					WithName("Flametongue Longsword").
					AsMagicItem("rare").
					Build()
				
				item.Properties["damage"] = "1d8+2d6"
				item.Properties["damageType"] = "slashing + fire"
				item.Properties["bonus"] = "+1"
				
				return repo.CreateItem(item)
			},
			Assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	testutil.RunDBTestCases(t, cases)
}

func TestInventoryRepository_GetItemByID(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewInventoryRepository(mockDB.DB)

	t.Run("successful retrieval", func(t *testing.T) {
		item := testutil.NewInventoryItemBuilder().WithName("Bag of Holding").Build()
		propertiesJSON, _ := json.Marshal(item.Properties)
		
		rows := sqlmock.NewRows([]string{
			"id", "character_id", "name", "type", "quantity", "weight",
			"value", "properties", "equipped", "description",
			"created_at", "updated_at",
		}).AddRow(
			item.ID, item.CharacterID, item.Name, item.Type, item.Quantity,
			item.Weight, item.Value, propertiesJSON, item.Equipped,
			item.Description, item.CreatedAt, item.UpdatedAt,
		)

		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM inventory_items WHERE id = \$1`,
		).WithArgs(int64(1)).WillReturnRows(rows)

		result, err := repo.GetItemByID(1)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Bag of Holding", result.Name)
		testutil.RequireValidInventoryItem(t, result)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("item not found", func(t *testing.T) {
		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM inventory_items WHERE id = \$1`,
		).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)

		result, err := repo.GetItemByID(999)
		require.Error(t, err)
		require.Nil(t, result)
		
		mockDB.AssertExpectations(t)
	})
}

func TestInventoryRepository_GetItemsByCharacterID(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewInventoryRepository(mockDB.DB)

	t.Run("multiple items with different types", func(t *testing.T) {
		weapon := testutil.NewInventoryItemBuilder().
			WithName("Longsword").
			WithType("weapon").
			Build()
		
		armor := testutil.NewInventoryItemBuilder().
			WithName("Chain Mail").
			AsArmor(16).
			Build()
		
		potion := testutil.NewInventoryItemBuilder().
			WithName("Healing Potion").
			WithType("consumable").
			WithQuantity(3).
			Build()

		weaponProps, _ := json.Marshal(weapon.Properties)
		armorProps, _ := json.Marshal(armor.Properties)
		potionProps, _ := json.Marshal(potion.Properties)

		rows := sqlmock.NewRows([]string{
			"id", "character_id", "name", "type", "quantity", "weight",
			"value", "properties", "equipped", "description",
			"created_at", "updated_at",
		}).
			AddRow(1, 1, weapon.Name, weapon.Type, weapon.Quantity,
				weapon.Weight, weapon.Value, weaponProps, weapon.Equipped,
				weapon.Description, weapon.CreatedAt, weapon.UpdatedAt).
			AddRow(2, 1, armor.Name, armor.Type, armor.Quantity,
				armor.Weight, armor.Value, armorProps, armor.Equipped,
				armor.Description, armor.CreatedAt, armor.UpdatedAt).
			AddRow(3, 1, potion.Name, potion.Type, potion.Quantity,
				potion.Weight, potion.Value, potionProps, potion.Equipped,
				potion.Description, potion.CreatedAt, potion.UpdatedAt)

		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM inventory_items WHERE character_id = \$1 ORDER BY name`,
		).WithArgs(int64(1)).WillReturnRows(rows)

		results, err := repo.GetItemsByCharacterID(1)
		require.NoError(t, err)
		require.Len(t, results, 3)
		
		// Verify items are sorted by name
		require.Equal(t, "Chain Mail", results[0].Name)
		require.Equal(t, "Healing Potion", results[1].Name)
		require.Equal(t, "Longsword", results[2].Name)
		
		// Validate all items
		for _, item := range results {
			testutil.RequireValidInventoryItem(t, item)
		}
		
		mockDB.AssertExpectations(t)
	})

	t.Run("empty inventory", func(t *testing.T) {
		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM inventory_items WHERE character_id = \$1 ORDER BY name`,
		).WithArgs(int64(999)).WillReturnRows(
			sqlmock.NewRows([]string{"id"}), // Empty result set
		)

		results, err := repo.GetItemsByCharacterID(999)
		require.NoError(t, err)
		require.Empty(t, results)
		
		mockDB.AssertExpectations(t)
	})
}

func TestInventoryRepository_UpdateItem(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewInventoryRepository(mockDB.DB)

	t.Run("successful update", func(t *testing.T) {
		item := testutil.NewInventoryItemBuilder().
			WithID(1).
			WithQuantity(5).
			Build()
		
		propertiesJSON, _ := json.Marshal(item.Properties)

		mockDB.Mock.ExpectExec(
			`UPDATE inventory_items SET name = \$2, type = \$3, quantity = \$4, 
			weight = \$5, value = \$6, properties = \$7, equipped = \$8, 
			description = \$9, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(
			item.ID, item.Name, item.Type, item.Quantity,
			item.Weight, item.Value, propertiesJSON, item.Equipped,
			item.Description,
		).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateItem(item)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("item not found", func(t *testing.T) {
		item := testutil.NewInventoryItemBuilder().WithID(999).Build()
		propertiesJSON, _ := json.Marshal(item.Properties)

		mockDB.Mock.ExpectExec(
			`UPDATE inventory_items SET.*WHERE id = \$1`,
		).WithArgs(
			item.ID, item.Name, item.Type, item.Quantity,
			item.Weight, item.Value, propertiesJSON, item.Equipped,
			item.Description,
		).WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateItem(item)
		require.Error(t, err)
		
		mockDB.AssertExpectations(t)
	})
}

func TestInventoryRepository_DeleteItem(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewInventoryRepository(mockDB.DB)

	t.Run("successful deletion", func(t *testing.T) {
		mockDB.Mock.ExpectExec(
			`DELETE FROM inventory_items WHERE id = \$1`,
		).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteItem(1)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})
}

func TestInventoryRepository_SpecializedQueries(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewInventoryRepository(mockDB.DB)

	t.Run("get equipped items", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "character_id", "name", "type", "quantity", "weight",
			"value", "properties", "equipped", "description",
			"created_at", "updated_at",
		}).
			AddRow(1, 1, "Longsword", "weapon", 1, 3.0, 15, `{}`, true, "", time.Now(), time.Now()).
			AddRow(2, 1, "Chain Mail", "armor", 1, 55.0, 75, `{}`, true, "", time.Now(), time.Now())

		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM inventory_items WHERE character_id = \$1 AND equipped = true`,
		).WithArgs(int64(1)).WillReturnRows(rows)

		items, err := repo.GetEquippedItems(1)
		require.NoError(t, err)
		require.Len(t, items, 2)
		
		for _, item := range items {
			require.True(t, item.Equipped)
		}
		
		mockDB.AssertExpectations(t)
	})

	t.Run("get items by type", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "character_id", "name", "type", "quantity", "weight",
			"value", "properties", "equipped", "description",
			"created_at", "updated_at",
		}).
			AddRow(1, 1, "Healing Potion", "consumable", 3, 0.5, 50, `{}`, false, "", time.Now(), time.Now()).
			AddRow(2, 1, "Mana Potion", "consumable", 2, 0.5, 75, `{}`, false, "", time.Now(), time.Now())

		mockDB.Mock.ExpectQuery(
			`SELECT \* FROM inventory_items WHERE character_id = \$1 AND type = \$2`,
		).WithArgs(int64(1), "consumable").WillReturnRows(rows)

		items, err := repo.GetItemsByType(1, "consumable")
		require.NoError(t, err)
		require.Len(t, items, 2)
		
		for _, item := range items {
			require.Equal(t, "consumable", item.Type)
		}
		
		mockDB.AssertExpectations(t)
	})

	t.Run("calculate total weight", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"total_weight"}).AddRow(125.5)

		mockDB.Mock.ExpectQuery(
			`SELECT COALESCE\(SUM\(weight \* quantity\), 0\) as total_weight FROM inventory_items WHERE character_id = \$1`,
		).WithArgs(int64(1)).WillReturnRows(rows)

		totalWeight, err := repo.GetTotalWeight(1)
		require.NoError(t, err)
		require.Equal(t, 125.5, totalWeight)
		
		mockDB.AssertExpectations(t)
	})

	t.Run("transfer item between characters", func(t *testing.T) {
		// Begin transaction
		mockDB.ExpectBegin()
		
		// Check item exists and belongs to source character
		rows := sqlmock.NewRows([]string{"character_id"}).AddRow(int64(1))
		mockDB.Mock.ExpectQuery(
			`SELECT character_id FROM inventory_items WHERE id = \$1 FOR UPDATE`,
		).WithArgs(int64(10)).WillReturnRows(rows)
		
		// Update item to new character
		mockDB.Mock.ExpectExec(
			`UPDATE inventory_items SET character_id = \$2, updated_at = CURRENT_TIMESTAMP WHERE id = \$1`,
		).WithArgs(int64(10), int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
		
		// Commit transaction
		mockDB.ExpectCommit()

		err := repo.TransferItem(10, 1, 2)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})
}

func TestInventoryRepository_BulkOperations(t *testing.T) {
	mockDB := testutil.NewMockDB(t)
	defer mockDB.Close()

	repo := NewInventoryRepository(mockDB.DB)

	t.Run("bulk create items", func(t *testing.T) {
		items := []*models.InventoryItem{
			testutil.NewInventoryItemBuilder().WithName("Sword").Build(),
			testutil.NewInventoryItemBuilder().WithName("Shield").Build(),
			testutil.NewInventoryItemBuilder().WithName("Potion").Build(),
		}

		// Expect a transaction
		mockDB.ExpectBegin()
		
		for _, item := range items {
			propertiesJSON, _ := json.Marshal(item.Properties)
			mockDB.Mock.ExpectQuery(
				`INSERT INTO inventory_items.*RETURNING id, created_at, updated_at`,
			).WithArgs(
				item.CharacterID, item.Name, item.Type, item.Quantity,
				item.Weight, item.Value, propertiesJSON, item.Equipped,
				item.Description,
			).WillReturnRows(
				sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
					AddRow(1, time.Now(), time.Now()),
			)
		}
		
		mockDB.ExpectCommit()

		err := repo.BulkCreateItems(items)
		require.NoError(t, err)
		
		mockDB.AssertExpectations(t)
	})
}