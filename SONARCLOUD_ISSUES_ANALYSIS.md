# SonarCloud Issues Analysis & Fix Plan

## Overview
- **Total Issues**: 1,573
- **Critical Issues**: 361 (all code smells)
- **Bugs**: 0 🎉
- **Vulnerabilities**: 0 🎉

## Critical Issues Breakdown (361 total)

### 1. Duplicate String Literals (~180 issues)
**Problem**: Hardcoded strings repeated multiple times should be constants
**Examples**:
- `"session-123"` repeated 77 times in game_session_test.go
- `"/api/v1/sessions/"` repeated 21 times in integration tests
- `"Content-Type"` and `"application/json"` repeated multiple times

**Fix Strategy**: Extract to constants at package or file level

### 2. Cognitive Complexity (~120 issues)
**Problem**: Methods exceeding complexity threshold of 15
**Worst Offenders**:
- `combat_analytics.go:165` - Complexity 50!
- `settlement_generator_test.go:292` - Complexity 37
- `combat_analytics.go:563` - Complexity 37

**Fix Strategy**: Break down into smaller, focused methods

### 3. Function Nesting Depth (~5 issues)
**Problem**: Functions nested more than 4 levels deep
**Location**: Frontend TypeScript test files
**Fix Strategy**: Extract nested functions or flatten logic

## Priority Order

### Phase 1: Quick Wins (Duplicate Literals)
1. **Test Constants** - Create test helper files with common strings
2. **API Endpoints** - Create route constants file
3. **HTTP Headers** - Create HTTP constants file

### Phase 2: Medium Effort (Function Nesting)
1. Fix CharacterBuilder test nesting
2. Fix useOptimizedState hook nesting

### Phase 3: High Effort (Cognitive Complexity)
1. Start with worst offenders (complexity > 30)
2. Focus on business-critical code (combat, analytics)
3. Ensure test coverage before refactoring

## Benefits
- **Maintainability**: Easier to update constants in one place
- **Readability**: Self-documenting constant names
- **Reliability**: Reduces typos and inconsistencies
- **Performance**: Slight memory optimization (string interning)