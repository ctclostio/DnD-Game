# PRNG False Positives Documentation

This document tracks all pseudorandom number generator (PRNG) uses in the codebase that have been marked as false positives for security scanning.

## Summary

All uses of `math/rand` in this codebase are for game mechanics, not security-sensitive operations. Security-sensitive operations correctly use `crypto/rand` through the `pkg/security/random.go` abstraction.

## NOSONAR Suppressions in living_ecosystem.go

The following uses of `math/rand` have been marked with NOSONAR comments:

### 1. NPC Goal Generation (Line 225)
```go
return len(goals) < 3 && rand.Float64() < 0.3 // NOSONAR: math/rand is appropriate for game mechanics (NPC goal probability)
```
- **Purpose**: Determines if an NPC should create a new goal (30% chance)
- **Security Risk**: None - purely game simulation

### 2. Political Opportunity Generation (Line 663)
```go
return rand.Float64() < 0.15*(timeDelta.Hours()/168.0) // NOSONAR: math/rand is appropriate for game mechanics (political opportunity generation)
```
- **Purpose**: Generates political opportunities for factions (15% chance per week)
- **Security Risk**: None - world event simulation

### 3. Faction Interaction Type Selection (Line 939)
```go
return interactionTypes[rand.Intn(len(interactionTypes))] // NOSONAR: math/rand is appropriate for game mechanics (faction interaction selection)
```
- **Purpose**: Randomly selects interaction types between factions
- **Security Risk**: None - gameplay variety

### 4-7. Faction Relationship Changes (Lines 964, 966, 968, 970)
```go
// Negative relationship changes for conflicts
return -(rand.Float64()*10 + 5) // NOSONAR: math/rand is appropriate for game mechanics (faction relationship changes)

// Positive relationship changes for cooperation
return rand.Float64()*10 + 5 // NOSONAR: math/rand is appropriate for game mechanics (faction relationship changes)

// Moderate changes for cultural exchanges
return rand.Float64()*5 + 2 // NOSONAR: math/rand is appropriate for game mechanics (faction relationship changes)

// Variable changes for default interactions
return (rand.Float64() - 0.5) * 10 // NOSONAR: math/rand is appropriate for game mechanics (faction relationship changes)
```
- **Purpose**: Calculates how faction relationships change based on different interactions
- **Security Risk**: None - diplomatic simulation

## Other Non-Flagged PRNG Uses

The following uses of `math/rand` exist but haven't been flagged by SonarCloud:
- Progress rate modifiers
- Event probabilities
- Random economic impacts
- Combat dice rolls
- Treasure generation

These are all similarly used for game mechanics and pose no security risk.

## Security-Sensitive Operations

All security-sensitive operations correctly use `crypto/rand`:
- JWT token generation (`internal/auth/jwt.go`)
- CSRF token generation (`internal/auth/csrf.go`)
- Session ID generation (`internal/services/game_session.go`)
- Secure random utilities (`pkg/security/random.go`)

## Guidelines for Developers

1. **Use `math/rand`** for:
   - Game mechanics (dice rolls, loot drops, NPC behavior)
   - Non-security randomness
   - Reproducible randomness (with seeds)

2. **Use `crypto/rand`** for:
   - Token generation
   - Session IDs
   - Cryptographic nonces
   - Any security-sensitive randomness

3. **When adding new PRNG uses**:
   - If it's for game mechanics, use `math/rand` and add a NOSONAR comment if flagged
   - If it's for security, use the utilities in `pkg/security/random.go`