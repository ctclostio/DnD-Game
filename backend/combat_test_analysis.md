# Combat Test Analysis

## Summary

The `combat_test.go.skip` file cannot be restored without significant changes because it tests methods that don't exist in the current `CombatService` implementation.

## Issues Found

### 1. Model Field Changes
- ✅ Fixed: `Type` → `ActionType` (for CombatAction)
- ✅ Fixed: `Turn` → `CurrentTurn` (for Combat)
- ✅ Fixed: `Initiative` → `InitiativeRoll` (for tracking roll vs modifier)
- ✅ Fixed: `AbilityScores` → `Attributes`
- ✅ Fixed: `Stats` → `Abilities` (map structure)
- ✅ Fixed: `Status` → `Condition`
- ✅ Fixed: Resistance/Immunity/Vulnerability field names

### 2. Missing Methods
The test expects these methods on `CombatService`, but they don't exist:
- `RollInitiative(combatants []models.Combatant) []models.Combatant`
- `Attack(attacker, target *models.Combatant, weapon *Weapon) (bool, int, error)`
- `SavingThrow(character *models.Character, saveType string, dc int) (bool, int, error)`
- `SkillCheck(character *models.Character, skill string, dc int) (bool, int, error)`
- `ApplyDamage(target *models.Combatant, damage int, damageType string) error`
- `CalculateAC(character *models.Character, armor *Armor, shield bool) int`

### 3. Current Implementation
The current `CombatService` has these methods instead:
- `StartCombat(ctx context.Context, gameSessionID string, combatants []models.Combatant) (*models.Combat, error)`
- `GetCombat(ctx context.Context, combatID string) (*models.Combat, error)`
- `NextTurn(ctx context.Context, combatID string) (*models.Combatant, error)`
- `ProcessAction(ctx context.Context, combatID string, request models.CombatRequest) (*models.CombatAction, error)`

### 4. Missing Types
The test defines its own types that don't exist in models:
- `Weapon` struct
- `Armor` struct
- `DieRollResult` struct (different from models.RollDetails)

## Recommendation

This test file appears to be from an older version of the codebase with a different combat system design. To properly restore it, we would need to either:

1. **Option A**: Implement the missing methods in `CombatService` or create a new service (e.g., `CombatMechanicsService`) that provides these D&D-specific combat mechanics.

2. **Option B**: Rewrite the tests to match the current implementation's approach, which seems more event-driven and session-based.

3. **Option C**: Keep it as `.skip` until the combat system is refactored to include these mechanics.

The test itself is well-written and covers important D&D mechanics like initiative, attacks, saving throws, skill checks, damage application, and AC calculation. These would be valuable to have in the system, but they need to be implemented first.