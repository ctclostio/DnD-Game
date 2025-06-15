# Phase 1: Critical Performance Fixes

## Overview
Identified 145 gocritic performance issues that need fixing:
- **30+ hugeParam**: Large structs passed by value (80-488 bytes)
- **20+ rangeValCopy**: Range loops copying large values (136-488 bytes per iteration)
- **40+ paramTypeCombine**: Multiple parameters of same type

## Priority Order (by impact)

### 1. Largest Memory Impact (Fix First)
These involve the biggest structs and most frequent operations:

#### Combat-Related (488 bytes per copy!)
- `models.Combatant` - 488 bytes
- Fix in: `combat_analytics.go`, `handlers/combat.go`
- Impact: High - combat happens frequently

#### Enemy/Encounter (360 bytes per copy)
- `models.Enemy` - 360 bytes  
- Fix in: `encounter_repository.go`, `encounter.go`, `ai_encounter_builder.go`
- Impact: High - encounters are core gameplay

#### Events (256 bytes per copy)
- `models.NarrativeEvent` - 256 bytes
- Fix in: `ai_narrative_engine.go`, `narrative.go`, `faction_personality.go`
- Impact: Medium-High - events trigger frequently

### 2. Hot Path Optimizations
These are in frequently called code paths:

#### Request Structs (80-152 bytes)
- Various request types in AI services
- Impact: Medium - called on every AI interaction

#### Database Filters (88 bytes)
- `NPCSearchFilter` and similar
- Impact: Medium - database queries are frequent

### 3. Loop Performance Killers
Range loops that copy large structs every iteration:

#### Worst Offenders:
- `combat_analytics.go:139,176,484` - Copying 488-byte Combatants
- `encounter.go:237,283` - Copying 360-byte Enemies
- `ai_encounter_builder.go:647,704` - Copying 360-byte Enemies

## Implementation Strategy

### Step 1: Fix Largest Structs First
Convert these to pointers with careful consideration:
- Add nil checks
- Consider ownership (who allocates/frees)
- Update all callers
- Test for any mutations

### Step 2: Fix Range Loops
Three approaches:
1. Use index-based loops: `for i := 0; i < len(slice); i++ { item := &slice[i] }`
2. Range over pointers: `for _, item := range []*Type{...}`
3. Store pointers in slice: `[]* Type` instead of `[]Type`

### Step 3: Combine Parameters
Simple refactoring:
```go
// Before
func Process(ctx context.Context, id1 string, id2 string) error

// After  
func Process(ctx context.Context, id1, id2 string) error
```

## Measurement Strategy
- Benchmark before/after key functions
- Monitor memory allocations
- Check GC pressure reduction