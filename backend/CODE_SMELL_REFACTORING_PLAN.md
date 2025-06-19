# Code Smell Refactoring Plan

## Overview
Total code smells: 1,461
Cognitive complexity: 9,001
Technical debt: ~153 hours

## Progress Tracking

### ✅ Completed (Phase 1 & 2 - Constants & Database Layer)
1. **Error Message Constants** - Created `internal/constants/errors.go`
   - 110+ error message patterns including database operations
   - Eliminates "character not found", "session not found", "game session not found" duplicates
   - Added database error format strings (marshal/unmarshal, rows affected)
   - Added migration error constants
   
2. **Common String Constants** - Created `internal/constants/strings.go`
   - HTTP headers, content types, SQL fragments
   - Database column names, status values
   
3. **Test Data Constants** - Created/Updated `internal/testutil/constants.go`
   - Common test data like "user-456", "user-42", "user-123"
   - Test password hash, email, API endpoints
   - Eliminates duplication in test files
   
4. **SQL Query Helpers** - Created `internal/database/query_helpers.go`
   - Query builder pattern for common SQL operations
   - Reduces "ORDER BY", "LIMIT ? OFFSET ?" duplicates
   - Added threePartFormat constant for query construction

5. **Database Repository Fixes** - Updated multiple repository files
   - Fixed character_repository.go, game_session_repository.go
   - Fixed emergent_world_repository.go, narrative_repository.go
   - Fixed migrate.go, inventory_repository_test.go, user_repository_test.go
   
6. **Handler Layer Fixes** - Updated handler files
   - Fixed auth.go (3 occurrences of "Invalid request body")
   - Fixed character.go (6 occurrences of "Invalid request body", 3 of "Character not found")
   - Added API endpoint constants for integration tests

### 🔄 In Progress (Phase 2 - High Complexity Methods)

Based on SonarCloud analysis, the following methods need refactoring:

#### Critical Complexity Issues (Cognitive Complexity > 20)
1. **backend/internal/middleware/cache.go** - Two methods with complexity 23 and 22
2. **backend/internal/pagination/pagination.go** - Complexity 23
3. **backend/pkg/security/random_test.go** - Complexity 22

#### High Complexity Issues (Cognitive Complexity > 15)
1. **backend/internal/cache/strategies.go** - Complexity 19
2. **backend/internal/health/db_pool_checker.go** - Complexity 17

### 📋 Upcoming (Phase 3 - Model Refactoring)

Large files that need breaking down:
1. **procedural_culture.go** (1,371 lines)
2. **living_ecosystem.go** (1,104 lines)
3. **combat_analytics.go** (951 lines)
4. **settlement_generator.go** (749 lines)

### 📊 Metrics to Track

- **Before**: 1,461 code smells, 9,001 cognitive complexity
- **After Phase 1**: Expected ~100 fewer duplicate string issues
- **Target**: < 1,000 code smells, < 7,000 cognitive complexity

## Refactoring Strategies

### For High Complexity Methods:
1. **Extract Method** - Break large methods into smaller, focused ones
2. **Early Returns** - Use guard clauses to reduce nesting
3. **Strategy Pattern** - Replace large switch statements
4. **Extract Validation** - Move validation logic to separate functions

### For Large Files:
1. **Single Responsibility** - One file per major concept
2. **Interface Segregation** - Split large interfaces
3. **Composition** - Use smaller, composable structures

## Next Steps

1. Tackle the cache middleware complexity (2 methods)
2. Refactor pagination logic to reduce complexity
3. Break down the procedural culture generator
4. Continue replacing duplicate strings with constants