-- Revert spell slots to simple array format
DO $$
DECLARE
    spell_slots_key CONSTANT TEXT := 'spellSlots';
BEGIN
    UPDATE characters 
    SET spells = jsonb_set(
        spells,
        ARRAY[spell_slots_key],
        CASE 
            WHEN spells->spell_slots_key IS NOT NULL AND jsonb_typeof(spells->spell_slots_key) = 'array' THEN
                (
                    SELECT jsonb_agg(
                        CASE 
                            WHEN jsonb_typeof(elem) = 'object' AND elem ? 'total' THEN
                                (elem->>'total')::int
                            ELSE elem
                        END
                    )
                    FROM jsonb_array_elements(spells->spell_slots_key) AS elem
                )
            ELSE '[]'::jsonb
        END
    )
    WHERE spells IS NOT NULL;
END $$;