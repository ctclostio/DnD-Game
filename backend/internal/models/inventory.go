package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type ItemType string

const (
	ItemTypeWeapon     ItemType = "weapon"
	ItemTypeArmor      ItemType = "armor"
	ItemTypeConsumable ItemType = "consumable"
	ItemTypeTool       ItemType = "tool"
	ItemTypeMagic      ItemType = "magic"
	ItemTypeOther      ItemType = "other"
)

type ItemRarity string

const (
	ItemRarityCommon    ItemRarity = "common"
	ItemRarityUncommon  ItemRarity = "uncommon"
	ItemRarityRare      ItemRarity = "rare"
	ItemRarityVeryRare  ItemRarity = "very_rare"
	ItemRarityLegendary ItemRarity = "legendary"
	ItemRarityArtifact  ItemRarity = "artifact"
)

type ItemProperties map[string]interface{}

func (p ItemProperties) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ItemProperties) Scan(value interface{}) error {
	if value == nil {
		*p = make(ItemProperties)
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, p)
}

type Item struct {
	ID                     string         `json:"id" db:"id"`
	Name                   string         `json:"name" db:"name"`
	Type                   ItemType       `json:"type" db:"type"`
	Rarity                 ItemRarity     `json:"rarity" db:"rarity"`
	Weight                 float64        `json:"weight" db:"weight"`
	Value                  int            `json:"value" db:"value"`
	Properties             ItemProperties `json:"properties" db:"properties"`
	RequiresAttunement     bool           `json:"requires_attunement" db:"requires_attunement"`
	AttunementRequirements string         `json:"attunement_requirements,omitempty" db:"attunement_requirements"`
	Description            string         `json:"description,omitempty" db:"description"`
	CreatedAt              time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at" db:"updated_at"`
}

type InventoryItem struct {
	ID               string         `json:"id" db:"id"`
	CharacterID      string         `json:"character_id" db:"character_id"`
	ItemID           string         `json:"item_id" db:"item_id"`
	Quantity         int            `json:"quantity" db:"quantity"`
	Equipped         bool           `json:"equipped" db:"equipped"`
	Attuned          bool           `json:"attuned" db:"attuned"`
	CustomProperties ItemProperties `json:"custom_properties" db:"custom_properties"`
	Notes            string         `json:"notes,omitempty" db:"notes"`
	Item             *Item          `json:"item,omitempty"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
}

type Currency struct {
	CharacterID string    `json:"character_id" db:"character_id"`
	Copper      int       `json:"copper" db:"copper"`
	Silver      int       `json:"silver" db:"silver"`
	Electrum    int       `json:"electrum" db:"electrum"`
	Gold        int       `json:"gold" db:"gold"`
	Platinum    int       `json:"platinum" db:"platinum"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

func (c *Currency) TotalInCopper() int {
	return c.Copper + (c.Silver * 10) + (c.Electrum * 50) + (c.Gold * 100) + (c.Platinum * 1000)
}

func (c *Currency) CanAfford(copperValue int) bool {
	return c.TotalInCopper() >= copperValue
}

func (c *Currency) Subtract(copperValue int) bool {
	if !c.CanAfford(copperValue) {
		return false
	}

	total := c.TotalInCopper() - copperValue

	c.Platinum = total / 1000
	total %= 1000

	c.Gold = total / 100
	total %= 100

	c.Electrum = total / 50
	total %= 50

	c.Silver = total / 10
	c.Copper = total % 10

	return true
}

type InventoryWeight struct {
	CurrentWeight     float64 `json:"current_weight"`
	CarryCapacity     float64 `json:"carry_capacity"`
	Encumbered        bool    `json:"encumbered"`
	HeavilyEncumbered bool    `json:"heavily_encumbered"`
}

func CalculateCarryCapacity(strength int) float64 {
	return float64(strength * 15)
}

func (w *InventoryWeight) UpdateEncumbrance() {
	w.Encumbered = w.CurrentWeight > w.CarryCapacity
	w.HeavilyEncumbered = w.CurrentWeight > (w.CarryCapacity * 2)
}
