package database

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/models"
)

func TestInventoryRepository_CreateItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewInventoryRepository(dbWrapper)

	t.Run("successful item creation", func(t *testing.T) {
		item := &models.Item{
			ID:          "test-item-id",
			Name:        "Longsword",
			Type:        models.ItemTypeWeapon,
			Rarity:      models.ItemRarityCommon,
			Weight:      3.0,
			Value:       15,
			Properties:  models.ItemProperties{"damage": "1d8", "type": "slashing"},
			Description: "A standard longsword",
		}
		
		mock.ExpectExec(
			`INSERT INTO items \(id, name, type, rarity, weight, value, properties, 
				requires_attunement, attunement_requirements, description, created_at, updated_at\)`,
		).WithArgs(
			item.ID, item.Name, item.Type, item.Rarity, item.Weight,
			item.Value, sqlmock.AnyArg(), item.RequiresAttunement, 
			item.AttunementRequirements, item.Description, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateItem(item)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create magic item with complex properties", func(t *testing.T) {
		item := &models.Item{
			ID:                 "magic-item-id",
			Name:               "Flametongue Longsword",
			Type:               models.ItemTypeWeapon,
			Rarity:             models.ItemRarityRare,
			Weight:             3.0,
			Value:              5000,
			RequiresAttunement: true,
			Properties: models.ItemProperties{
				"damage":     "1d8+2d6",
				"damageType": "slashing + fire",
				"bonus":      "+1",
			},
			Description: "A magical sword wreathed in flames",
		}
		
		mock.ExpectExec(
			`INSERT INTO items`,
		).WithArgs(
			item.ID, item.Name, item.Type, item.Rarity, item.Weight,
			item.Value, sqlmock.AnyArg(), item.RequiresAttunement,
			item.AttunementRequirements, item.Description, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateItem(item)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepository_GetItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewInventoryRepository(dbWrapper)

	t.Run("successful retrieval", func(t *testing.T) {
		expectedItem := &models.Item{
			ID:          "test-item-id",
			Name:        "Bag of Holding",
			Type:        models.ItemTypeMagic,
			Rarity:      models.ItemRarityUncommon,
			Weight:      15.0,
			Value:       500,
			Properties:  models.ItemProperties{"capacity": "500 lbs"},
			Description: "This bag has an interior space considerably larger than its outside dimensions",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		propertiesJSON, _ := json.Marshal(expectedItem.Properties)
		
		mock.ExpectQuery(`SELECT id, name, type, rarity, weight, value, properties, requires_attunement, attunement_requirements, description, created_at, updated_at FROM items WHERE id = \?`).
			WithArgs("test-item-id").
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "type", "rarity", "weight", "value", "properties",
				"requires_attunement", "attunement_requirements", "description",
				"created_at", "updated_at",
			}).AddRow(
				expectedItem.ID, expectedItem.Name, expectedItem.Type, expectedItem.Rarity,
				expectedItem.Weight, expectedItem.Value, propertiesJSON,
				expectedItem.RequiresAttunement, expectedItem.AttunementRequirements,
				expectedItem.Description, expectedItem.CreatedAt, expectedItem.UpdatedAt,
			))

		item, err := repo.GetItem("test-item-id")
		assert.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, expectedItem.Name, item.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("item not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, type, rarity, weight, value, properties, requires_attunement, attunement_requirements, description, created_at, updated_at FROM items WHERE id = \?`).
			WithArgs("non-existent").
			WillReturnError(sql.ErrNoRows)

		item, err := repo.GetItem("non-existent")
		assert.NoError(t, err)
		assert.Nil(t, item)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepository_AddItemToInventory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewInventoryRepository(dbWrapper)

	t.Run("add new item to inventory", func(t *testing.T) {
		characterID := "char-123"
		itemID := "item-456"
		quantity := 1

		// Expect the insert/update query
		mock.ExpectExec(
			`INSERT INTO character_inventory \(id, character_id, item_id, quantity, created_at, updated_at\)`,
		).WithArgs(
			sqlmock.AnyArg(), characterID, itemID, quantity, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect the weight update query from updateCharacterWeight
		// Need to match the full query including WHERE clause
		mock.ExpectExec(
			`UPDATE characters SET current_weight = \(\s*SELECT COALESCE\(SUM\(i\.weight \* ci\.quantity\), 0\)\s*FROM character_inventory ci\s*JOIN items i ON ci\.item_id = i\.id\s*WHERE ci\.character_id = \?\s*\)\s*WHERE id = \?`,
		).WithArgs(characterID, characterID).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.AddItemToInventory(characterID, itemID, quantity)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepository_GetCharacterInventory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewInventoryRepository(dbWrapper)

	t.Run("get character inventory with items", func(t *testing.T) {
		characterID := "char-123"
		
		// Mock the query that joins character_inventory with items
		rows := sqlmock.NewRows([]string{
			"id", "character_id", "item_id", "quantity", "equipped", "attuned",
			"custom_properties", "notes", "created_at", "updated_at",
			"item_id", "name", "type", "rarity", "weight", "value", "properties",
			"requires_attunement", "attunement_requirements", "description",
			"item_created_at", "item_updated_at",
		}).AddRow(
			"inv-1", characterID, "item-1", 1, false, false,
			"{}", "", time.Now(), time.Now(),
			"item-1", "Longsword", "weapon", "common", 3.0, 15, `{"damage":"1d8"}`,
			false, "", "A standard longsword",
			time.Now(), time.Now(),
		)

		mock.ExpectQuery(`SELECT ci\.id, ci\.character_id, ci\.item_id, ci\.quantity, ci\.equipped, ci\.attuned, ci\.custom_properties, ci\.notes, ci\.created_at, ci\.updated_at, i\.id, i\.name, i\.type, i\.rarity, i\.weight, i\.value, i\.properties, i\.requires_attunement, i\.attunement_requirements, i\.description, i\.created_at, i\.updated_at FROM character_inventory ci JOIN items i ON ci\.item_id = i\.id WHERE ci\.character_id = \? ORDER BY i\.name`).
			WithArgs(characterID).
			WillReturnRows(rows)

		items, err := repo.GetCharacterInventory(characterID)
		assert.NoError(t, err)
		assert.Len(t, items, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInventoryRepository_GetCharacterCurrency(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	dbWrapper := &DB{DB: sqlxDB}
	repo := NewInventoryRepository(dbWrapper)

	t.Run("get existing currency", func(t *testing.T) {
		characterID := "char-123"
		expectedCurrency := &models.Currency{
			CharacterID: characterID,
			Copper:      50,
			Silver:      25,
			Electrum:    0,
			Gold:        100,
			Platinum:    5,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mock.ExpectQuery(`SELECT \* FROM character_currency WHERE character_id = \?`).
			WithArgs(characterID).
			WillReturnRows(sqlmock.NewRows([]string{
				"character_id", "copper", "silver", "electrum", "gold", "platinum",
				"created_at", "updated_at",
			}).AddRow(
				expectedCurrency.CharacterID, expectedCurrency.Copper,
				expectedCurrency.Silver, expectedCurrency.Electrum,
				expectedCurrency.Gold, expectedCurrency.Platinum,
				expectedCurrency.CreatedAt, expectedCurrency.UpdatedAt,
			))

		currency, err := repo.GetCharacterCurrency(characterID)
		assert.NoError(t, err)
		assert.NotNil(t, currency)
		assert.Equal(t, expectedCurrency.Gold, currency.Gold)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no currency found - create default", func(t *testing.T) {
		characterID := "char-456"

		mock.ExpectQuery(`SELECT \* FROM character_currency WHERE character_id = \?`).
			WithArgs(characterID).
			WillReturnError(sql.ErrNoRows)

		// Expect insert of default currency
		mock.ExpectExec(`INSERT INTO character_currency`).
			WithArgs(characterID, 0, 0, 0, 0, 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		currency, err := repo.GetCharacterCurrency(characterID)
		assert.NoError(t, err)
		assert.NotNil(t, currency)
		assert.Equal(t, 0, currency.Gold)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}