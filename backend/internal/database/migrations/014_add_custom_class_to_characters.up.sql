-- Add custom_class_id to characters table
ALTER TABLE characters
ADD COLUMN custom_class_id UUID REFERENCES custom_classes(id) ON DELETE SET NULL;

-- Add index for performance
CREATE INDEX idx_characters_custom_class_id ON characters(custom_class_id);