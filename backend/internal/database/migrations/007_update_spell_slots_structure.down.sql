-- Revert spell slots to simple array format
UPDATE characters 
SET spells = jsonb_set(
    spells,
    '{spellSlots}',
    CASE 
        WHEN spells->'spellSlots' IS NOT NULL AND jsonb_typeof(spells->'spellSlots') = 'array' THEN
            (
                SELECT jsonb_agg(
                    CASE 
                        WHEN jsonb_typeof(elem) = 'object' AND elem ? 'total' THEN
                            (elem->>'total')::int
                        ELSE elem
                    END
                )
                FROM jsonb_array_elements(spells->'spellSlots') AS elem
            )
        ELSE '[]'::jsonb
    END
)
WHERE spells IS NOT NULL;