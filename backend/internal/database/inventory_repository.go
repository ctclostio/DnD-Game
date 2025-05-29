package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/your-username/dnd-game/backend/internal/models"
)

type inventoryRepository struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) CreateItem(item *models.Item) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	query := `
		INSERT INTO items (id, name, type, rarity, weight, value, properties, 
			requires_attunement, attunement_requirements, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(query, item.ID, item.Name, item.Type, item.Rarity, item.Weight,
		item.Value, item.Properties, item.RequiresAttunement, item.AttunementRequirements,
		item.Description, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *inventoryRepository) GetItem(itemID string) (*models.Item, error) {
	var item models.Item
	query := `SELECT * FROM items WHERE id = $1`
	err := r.db.Get(&item, query, itemID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &item, err
}

func (r *inventoryRepository) GetItemsByType(itemType models.ItemType) ([]*models.Item, error) {
	var items []*models.Item
	query := `SELECT * FROM items WHERE type = $1 ORDER BY name`
	err := r.db.Select(&items, query, itemType)
	return items, err
}

func (r *inventoryRepository) AddItemToInventory(characterID, itemID string, quantity int) error {
	id := uuid.New().String()
	now := time.Now()
	
	query := `
		INSERT INTO character_inventory (id, character_id, item_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (character_id, item_id) 
		DO UPDATE SET quantity = character_inventory.quantity + EXCLUDED.quantity,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query, id, characterID, itemID, quantity, now, now)
	if err != nil {
		return err
	}
	
	return r.updateCharacterWeight(characterID)
}

func (r *inventoryRepository) RemoveItemFromInventory(characterID, itemID string, quantity int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentQty int
	err = tx.Get(&currentQty, `SELECT quantity FROM character_inventory WHERE character_id = $1 AND item_id = $2`, characterID, itemID)
	if err != nil {
		return err
	}

	if currentQty <= quantity {
		_, err = tx.Exec(`DELETE FROM character_inventory WHERE character_id = $1 AND item_id = $2`, characterID, itemID)
	} else {
		_, err = tx.Exec(`UPDATE character_inventory SET quantity = quantity - $3, updated_at = $4 WHERE character_id = $1 AND item_id = $2`,
			characterID, itemID, quantity, time.Now())
	}
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return r.updateCharacterWeight(characterID)
}

func (r *inventoryRepository) GetCharacterInventory(characterID string) ([]*models.InventoryItem, error) {
	query := `
		SELECT ci.*, i.* 
		FROM character_inventory ci
		JOIN items i ON ci.item_id = i.id
		WHERE ci.character_id = $1
		ORDER BY i.name
	`
	
	rows, err := r.db.Query(query, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.InventoryItem
	for rows.Next() {
		var inv models.InventoryItem
		var item models.Item
		
		err := rows.Scan(
			&inv.ID, &inv.CharacterID, &inv.ItemID, &inv.Quantity,
			&inv.Equipped, &inv.Attuned, &inv.CustomProperties, &inv.Notes,
			&inv.CreatedAt, &inv.UpdatedAt,
			&item.ID, &item.Name, &item.Type, &item.Rarity, &item.Weight,
			&item.Value, &item.Properties, &item.RequiresAttunement,
			&item.AttunementRequirements, &item.Description,
			&item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		inv.Item = &item
		items = append(items, &inv)
	}
	
	return items, nil
}

func (r *inventoryRepository) EquipItem(characterID, itemID string, equip bool) error {
	query := `UPDATE character_inventory SET equipped = $3, updated_at = $4 WHERE character_id = $1 AND item_id = $2`
	_, err := r.db.Exec(query, characterID, itemID, equip, time.Now())
	return err
}

func (r *inventoryRepository) AttuneItem(characterID, itemID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var slotsUsed, maxSlots int
	err = tx.Get(&slotsUsed, `SELECT attunement_slots_used FROM characters WHERE id = $1`, characterID)
	if err != nil {
		return err
	}
	
	err = tx.Get(&maxSlots, `SELECT attunement_slots_max FROM characters WHERE id = $1`, characterID)
	if err != nil {
		return err
	}

	if slotsUsed >= maxSlots {
		return fmt.Errorf("maximum attunement slots (%d) already in use", maxSlots)
	}

	var requiresAttunement bool
	err = tx.Get(&requiresAttunement, `SELECT requires_attunement FROM items WHERE id = $1`, itemID)
	if err != nil {
		return err
	}

	if !requiresAttunement {
		return fmt.Errorf("item does not require attunement")
	}

	_, err = tx.Exec(`UPDATE character_inventory SET attuned = true, updated_at = $3 WHERE character_id = $1 AND item_id = $2`,
		characterID, itemID, time.Now())
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE characters SET attunement_slots_used = attunement_slots_used + 1 WHERE id = $1`, characterID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *inventoryRepository) UnattuneItem(characterID, itemID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE character_inventory SET attuned = false, updated_at = $3 WHERE character_id = $1 AND item_id = $2`,
		characterID, itemID, time.Now())
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE characters SET attunement_slots_used = attunement_slots_used - 1 WHERE id = $1`, characterID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *inventoryRepository) GetCharacterCurrency(characterID string) (*models.Currency, error) {
	var currency models.Currency
	query := `SELECT * FROM character_currency WHERE character_id = $1`
	err := r.db.Get(&currency, query, characterID)
	if err == sql.ErrNoRows {
		currency = models.Currency{
			CharacterID: characterID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err = r.CreateCharacterCurrency(&currency)
		if err != nil {
			return nil, err
		}
	}
	return &currency, err
}

func (r *inventoryRepository) CreateCharacterCurrency(currency *models.Currency) error {
	query := `
		INSERT INTO character_currency (character_id, copper, silver, electrum, gold, platinum, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query, currency.CharacterID, currency.Copper, currency.Silver,
		currency.Electrum, currency.Gold, currency.Platinum, currency.CreatedAt, currency.UpdatedAt)
	return err
}

func (r *inventoryRepository) UpdateCharacterCurrency(currency *models.Currency) error {
	currency.UpdatedAt = time.Now()
	query := `
		UPDATE character_currency 
		SET copper = $2, silver = $3, electrum = $4, gold = $5, platinum = $6, updated_at = $7
		WHERE character_id = $1
	`
	_, err := r.db.Exec(query, currency.CharacterID, currency.Copper, currency.Silver,
		currency.Electrum, currency.Gold, currency.Platinum, currency.UpdatedAt)
	return err
}

func (r *inventoryRepository) updateCharacterWeight(characterID string) error {
	query := `
		UPDATE characters 
		SET current_weight = (
			SELECT COALESCE(SUM(i.weight * ci.quantity), 0)
			FROM character_inventory ci
			JOIN items i ON ci.item_id = i.id
			WHERE ci.character_id = $1
		)
		WHERE id = $1
	`
	_, err := r.db.Exec(query, characterID)
	return err
}

func (r *inventoryRepository) GetCharacterWeight(characterID string) (*models.InventoryWeight, error) {
	var weight models.InventoryWeight
	query := `
		SELECT current_weight, carry_capacity 
		FROM characters 
		WHERE id = $1
	`
	err := r.db.QueryRow(query, characterID).Scan(&weight.CurrentWeight, &weight.CarryCapacity)
	if err != nil {
		return nil, err
	}
	
	weight.UpdateEncumbrance()
	return &weight, nil
}