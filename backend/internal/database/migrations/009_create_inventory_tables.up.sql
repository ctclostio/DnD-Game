-- Create items table
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- weapon, armor, consumable, tool, etc.
    rarity TEXT DEFAULT 'common', -- common, uncommon, rare, very_rare, legendary
    weight REAL DEFAULT 0,
    value INTEGER DEFAULT 0, -- in copper pieces
    properties JSONB DEFAULT '{}', -- damage, AC bonus, special properties
    requires_attunement BOOLEAN DEFAULT FALSE,
    attunement_requirements TEXT, -- class/race requirements
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create character_inventory table
CREATE TABLE IF NOT EXISTS character_inventory (
    id TEXT PRIMARY KEY,
    character_id TEXT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    item_id TEXT NOT NULL REFERENCES items(id),
    quantity INTEGER DEFAULT 1,
    equipped BOOLEAN DEFAULT FALSE,
    attuned BOOLEAN DEFAULT FALSE,
    custom_properties JSONB DEFAULT '{}', -- for magical +1, +2 etc modifications
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_character_item UNIQUE(character_id, item_id)
);

-- Create character_currency table
CREATE TABLE IF NOT EXISTS character_currency (
    character_id TEXT PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    copper INTEGER DEFAULT 0,
    silver INTEGER DEFAULT 0,
    electrum INTEGER DEFAULT 0,
    gold INTEGER DEFAULT 0,
    platinum INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add inventory-related fields to characters table
ALTER TABLE characters 
ADD COLUMN IF NOT EXISTS carry_capacity REAL DEFAULT 0,
ADD COLUMN IF NOT EXISTS current_weight REAL DEFAULT 0,
ADD COLUMN IF NOT EXISTS attunement_slots_used INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS attunement_slots_max INTEGER DEFAULT 3;

-- Create indexes
CREATE INDEX idx_items_type ON items(type);
CREATE INDEX idx_items_rarity ON items(rarity);
CREATE INDEX idx_character_inventory_character ON character_inventory(character_id);
CREATE INDEX idx_character_inventory_equipped ON character_inventory(equipped);
CREATE INDEX idx_character_inventory_attuned ON character_inventory(attuned);