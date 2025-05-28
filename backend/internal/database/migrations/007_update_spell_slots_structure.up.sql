-- Update spell slots to support current vs total tracking
-- This migration updates the spells JSONB column to support the new spell slot structure

-- First, update any existing spell slot data to the new format
UPDATE characters 
SET spells = jsonb_set(
    spells,
    '{spellSlots}',
    CASE 
        WHEN spells->'spellSlots' IS NOT NULL AND jsonb_typeof(spells->'spellSlots') = 'array' THEN
            (
                SELECT jsonb_agg(
                    CASE 
                        WHEN jsonb_typeof(elem) = 'number' THEN
                            jsonb_build_object(
                                'level', (row_number() OVER ())::int,
                                'total', elem::int,
                                'remaining', elem::int
                            )
                        ELSE elem
                    END
                )
                FROM jsonb_array_elements(spells->'spellSlots') WITH ORDINALITY AS t(elem, idx)
            )
        ELSE '[]'::jsonb
    END
)
WHERE spells IS NOT NULL;