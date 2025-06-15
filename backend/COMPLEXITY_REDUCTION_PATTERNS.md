# Complexity Reduction Patterns

## Overview
This guide provides patterns and techniques for reducing cyclomatic complexity in the D&D Game backend. Based on analysis of the codebase, we've identified functions with complexity scores ranging from 12-35 that need refactoring.

## Common Complexity Issues

### 1. **Large Switch Statements**
**Problem**: Switch statements with 10+ cases increase cyclomatic complexity linearly.

**Solution**: Replace with strategy pattern or function maps.

**Before**:
```go
func processAction(actionType string, data interface{}) error {
    switch actionType {
    case "attack":
        // 20 lines of attack logic
    case "spell":
        // 25 lines of spell logic
    case "move":
        // 15 lines of move logic
    // ... 10 more cases
    }
}
```

**After**:
```go
type ActionHandler func(data interface{}) error

var actionHandlers = map[string]ActionHandler{
    "attack": handleAttack,
    "spell":  handleSpell,
    "move":   handleMove,
}

func processAction(actionType string, data interface{}) error {
    handler, exists := actionHandlers[actionType]
    if !exists {
        return fmt.Errorf("unknown action type: %s", actionType)
    }
    return handler(data)
}
```

### 2. **Deeply Nested Conditions**
**Problem**: Multiple levels of if statements create complexity and readability issues.

**Solution**: Use early returns, guard clauses, and extract to methods.

**Before**:
```go
func validateCharacter(char *Character) error {
    if char != nil {
        if char.Level > 0 {
            if char.Level <= 20 {
                if char.HitPoints > 0 {
                    if char.Stats.Strength >= 1 && char.Stats.Strength <= 30 {
                        // More nested conditions...
                        return nil
                    }
                }
            }
        }
    }
    return errors.New("invalid character")
}
```

**After**:
```go
func validateCharacter(char *Character) error {
    if char == nil {
        return errors.New("character is nil")
    }
    
    if err := validateLevel(char.Level); err != nil {
        return err
    }
    
    if err := validateHitPoints(char.HitPoints); err != nil {
        return err
    }
    
    return validateStats(&char.Stats)
}

func validateLevel(level int) error {
    if level < 1 || level > 20 {
        return fmt.Errorf("invalid level: %d", level)
    }
    return nil
}
```

### 3. **Long Functions with Multiple Responsibilities**
**Problem**: Functions doing too many things become hard to understand and test.

**Solution**: Extract cohesive groups of operations into separate functions.

**Before**:
```go
func processEncounter(encounter *Encounter) error {
    // 30 lines of validation logic
    // 40 lines of enemy initialization
    // 25 lines of reward calculation
    // 35 lines of objective setup
    // 20 lines of notification sending
}
```

**After**:
```go
func processEncounter(encounter *Encounter) error {
    if err := validateEncounter(encounter); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    enemies, err := initializeEnemies(encounter)
    if err != nil {
        return fmt.Errorf("enemy initialization failed: %w", err)
    }
    
    encounter.Enemies = enemies
    encounter.Rewards = calculateRewards(encounter)
    encounter.Objectives = setupObjectives(encounter)
    
    return notifyParticipants(encounter)
}
```

### 4. **Embedded Data Tables**
**Problem**: Large data tables in functions increase line count and complexity.

**Solution**: Extract to configuration or constant structures.

**Before**:
```go
func getSpellSlots(class string, level int) []int {
    switch class {
    case "wizard":
        switch level {
        case 1: return []int{2, 0, 0, 0, 0, 0, 0, 0, 0}
        case 2: return []int{3, 0, 0, 0, 0, 0, 0, 0, 0}
        // ... 18 more cases
        }
    case "cleric":
        // ... another 20 cases
    }
}
```

**After**:
```go
var spellSlotProgression = map[string]map[int][]int{
    "wizard": wizardSpellSlots,
    "cleric": clericSpellSlots,
}

func getSpellSlots(class string, level int) []int {
    if progression, ok := spellSlotProgression[class]; ok {
        if slots, ok := progression[level]; ok {
            return slots
        }
    }
    return defaultSpellSlots
}
```

### 5. **Type Assertion Chains**
**Problem**: Multiple type assertions create branching complexity.

**Solution**: Use type switches or visitor pattern.

**Before**:
```go
func processValue(v interface{}) (int, error) {
    if intVal, ok := v.(int); ok {
        return intVal, nil
    }
    if floatVal, ok := v.(float64); ok {
        return int(floatVal), nil
    }
    if strVal, ok := v.(string); ok {
        if intVal, err := strconv.Atoi(strVal); err == nil {
            return intVal, nil
        }
    }
    // More type assertions...
}
```

**After**:
```go
func processValue(v interface{}) (int, error) {
    switch val := v.(type) {
    case int:
        return val, nil
    case float64:
        return int(val), nil
    case string:
        return strconv.Atoi(val)
    default:
        return 0, fmt.Errorf("unsupported type: %T", v)
    }
}
```

## Refactoring Strategies

### 1. **Extract Method**
Break down large functions into smaller, focused methods.

### 2. **Replace Conditional with Polymorphism**
Use interfaces and different implementations instead of if/switch statements.

### 3. **Introduce Parameter Object**
Group related parameters into a struct.

### 4. **Replace Nested Conditional with Guard Clauses**
Use early returns to reduce nesting.

### 5. **Extract Class/Service**
Move cohesive groups of functions to a dedicated service.

### 6. **Use Table-Driven Methods**
Replace complex conditionals with data structures.

### 7. **Apply Chain of Responsibility**
For sequential processing with multiple handlers.

### 8. **Introduce Null Object**
Eliminate null checks by using a null object pattern.

## Complexity Metrics Guidelines

### Target Complexity Scores
- **Ideal**: < 10
- **Acceptable**: 10-15
- **Needs Refactoring**: 15-25
- **Critical**: > 25

### Function Length Guidelines
- **Ideal**: < 20 lines
- **Acceptable**: 20-50 lines
- **Needs Refactoring**: 50-100 lines
- **Critical**: > 100 lines

### Nesting Depth Guidelines
- **Ideal**: 1-2 levels
- **Acceptable**: 3 levels
- **Needs Refactoring**: 4 levels
- **Critical**: > 4 levels

## Example Refactorings

### 1. Strategy Pattern for Complex Actions
```go
// Define strategy interface
type CombatActionStrategy interface {
    Execute(context *CombatContext) (*CombatResult, error)
    Validate(context *CombatContext) error
}

// Implement strategies
type AttackStrategy struct{}
type SpellStrategy struct{}
type MoveStrategy struct{}

// Use strategy in combat service
type CombatService struct {
    strategies map[string]CombatActionStrategy
}

func (s *CombatService) ExecuteAction(actionType string, context *CombatContext) (*CombatResult, error) {
    strategy, exists := s.strategies[actionType]
    if !exists {
        return nil, fmt.Errorf("unknown action type: %s", actionType)
    }
    
    if err := strategy.Validate(context); err != nil {
        return nil, err
    }
    
    return strategy.Execute(context)
}
```

### 2. Builder Pattern for Complex Objects
```go
type CharacterBuilder struct {
    character *Character
}

func NewCharacterBuilder() *CharacterBuilder {
    return &CharacterBuilder{
        character: &Character{
            Stats: CharacterStats{},
        },
    }
}

func (b *CharacterBuilder) WithName(name string) *CharacterBuilder {
    b.character.Name = name
    return b
}

func (b *CharacterBuilder) WithClass(class string) *CharacterBuilder {
    b.character.Class = class
    b.initializeClassFeatures()
    return b
}

func (b *CharacterBuilder) WithLevel(level int) *CharacterBuilder {
    b.character.Level = level
    b.calculateDerivedStats()
    return b
}

func (b *CharacterBuilder) Build() (*Character, error) {
    if err := b.validate(); err != nil {
        return nil, err
    }
    return b.character, nil
}
```

### 3. Pipeline Pattern for Sequential Processing
```go
type EncounterProcessor func(*Encounter) error

type EncounterPipeline struct {
    processors []EncounterProcessor
}

func (p *EncounterPipeline) Process(encounter *Encounter) error {
    for _, processor := range p.processors {
        if err := processor(encounter); err != nil {
            return fmt.Errorf("pipeline failed: %w", err)
        }
    }
    return nil
}

// Usage
pipeline := &EncounterPipeline{
    processors: []EncounterProcessor{
        validateEncounter,
        initializeEnemies,
        calculateDifficulty,
        setupObjectives,
        assignRewards,
    },
}
```

## Testing Strategies for Reduced Complexity

### 1. **Unit Test Each Extracted Method**
Smaller functions are easier to test in isolation.

### 2. **Use Table-Driven Tests**
Test multiple scenarios without duplicating code.

### 3. **Mock Dependencies**
Focused interfaces make mocking simpler.

### 4. **Test Edge Cases**
Simpler functions make edge cases more obvious.

## Conclusion

Reducing complexity is not just about lowering numbers—it's about creating code that is:
- Easier to understand
- Simpler to test
- More maintainable
- Less prone to bugs
- More flexible for future changes

Apply these patterns incrementally, always ensuring tests pass after each refactoring step.