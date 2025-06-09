# Backend Test Improvements Roadmap

## Overview
This document tracks the ongoing improvements to backend testing and CI/CD compatibility following the SQL migration work completed on January 9, 2025.

## Current State
- **Test Coverage**: ~84% (services), but lower overall due to skipped tests
- **CI/CD Status**: Passing with critical path fixes
- **Database Compatibility**: Working with both SQLite (local) and PostgreSQL (CI/CD)

## Priority 1: Complete SQL Migration (1-2 weeks)

### Large Repositories (35+ queries each)
These repositories have the most queries and should be split into smaller PRs:

1. **dm_assistant_repository.go** (35 queries)
   - [ ] Split into 3-4 PRs by feature area
   - [ ] Test DM tool functionality thoroughly
   - Critical for: DM Assistant features

2. **emergent_world_repository.go** (35 queries)
   - [ ] Complex world generation queries
   - [ ] May need performance testing after migration
   - Critical for: World simulation features

3. **world_building_repository.go** (35 queries)
   - [ ] Settlement and faction queries
   - [ ] Economic simulation queries
   - Critical for: World builder tools

### Medium Repositories (15-25 queries)
4. **campaign_repository.go** (20 queries)
5. **narrative_repository.go** (23 queries)
6. **encounter_repository.go** (19 queries)
7. **npc_repository.go** (17 queries)
8. **custom_race_repository.go** (17 queries)

### Small Repositories (<15 queries)
9. **combat_analytics_repository.go** (12 queries)
10. **custom_class_repository.go** (7 queries)
11. **rule_builder_repository.go** (12 queries)

## Priority 2: Restore Skipped Tests (2-3 weeks)

### Service Tests (High Value)
These tests provide critical coverage and should be restored first:

1. **Combat-related tests**
   - [ ] `combat_analytics_test.go.skip`
   - [ ] `combat_automation_test.go.skip`
   - [ ] `combat_test.go.skip` (handler)

2. **AI-related tests**
   - [ ] `ai_character_test.go.skip`
   - [ ] `ai_class_generator_test.go.skip`
   - [ ] `ai_dm_assistant_test.go.skip`

3. **Core gameplay tests**
   - [ ] `game_session_test.go.skip`
   - [ ] `game_test.go.skip`
   - [ ] `encounter_test.go.skip`

4. **Character-related tests**
   - [ ] `character_builder_test.go.skip`
   - [ ] `character_test.go.skip` (handler)
   - [ ] `custom_race_test.go.skip`

5. **Other service tests**
   - [ ] `campaign_test.go.skip`
   - [ ] `dm_assistant_test.go.skip`
   - [ ] `npc_test.go.skip`
   - [ ] `rule_engine_test.go.skip`
   - [ ] `settlement_generator_test.go.skip`
   - [ ] `world_event_engine_test.go.skip`

### Handler Tests (Integration)
6. **Handler integration tests**
   - [ ] `auth_integration_test.go.skip`
   - [ ] `dice_test.go.skip`
   - [ ] `inventory_test.go.skip`
   - [ ] `refresh_token_test.go.skip`

## Priority 3: Test Quality Improvements (1 week)

### Test Infrastructure
1. **Test Helpers**
   - [ ] Create repository test helper for SQL migration testing
   - [ ] Add test fixtures for complex queries
   - [ ] Implement test data builders

2. **Mock Improvements**
   - [ ] Update mock generators
   - [ ] Ensure all mocks match interfaces
   - [ ] Add mock validation helpers

3. **Integration Test Suite**
   - [ ] Create end-to-end test scenarios
   - [ ] Add performance benchmarks
   - [ ] Implement load testing for repositories

### Documentation
4. **Test Documentation**
   - [ ] Document test patterns
   - [ ] Create testing best practices guide
   - [ ] Add troubleshooting guide for common test failures

## Priority 4: CI/CD Enhancements (1 week)

1. **Pipeline Optimization**
   - [ ] Add test result caching
   - [ ] Implement parallel test execution
   - [ ] Add test flakiness detection

2. **Database Testing**
   - [ ] Add PostgreSQL integration tests in CI
   - [ ] Test migration scripts in CI
   - [ ] Add database version compatibility tests

3. **Coverage Improvements**
   - [ ] Generate detailed coverage reports
   - [ ] Add coverage trends tracking
   - [ ] Set up coverage gates for PRs

## Low Priority: Nice-to-Have

1. **Logger Test Fixes**
   - [ ] Fix JSON parsing issues in logger tests
   - [ ] Update logger test utilities

2. **Cleanup Tasks**
   - [ ] Remove `.old` test files
   - [ ] Clean up duplicate test code
   - [ ] Standardize test naming conventions

## Success Metrics

### Short Term (2 weeks)
- [ ] All critical repositories migrated to database-agnostic SQL
- [ ] CI/CD pipeline passing consistently
- [ ] No test failures due to SQL syntax

### Medium Term (1 month)
- [ ] Test coverage back to 84%+ overall
- [ ] All skipped tests restored and passing
- [ ] Zero flaky tests in CI/CD

### Long Term (2 months)
- [ ] 90%+ test coverage
- [ ] Full integration test suite
- [ ] Sub-5 minute CI/CD runs

## Risk Mitigation

1. **Large Repository Migration**
   - Split into smaller PRs
   - Test incrementally
   - Have rollback plan

2. **Test Restoration**
   - Fix compilation errors first
   - Update mocks before logic
   - Run locally before CI

3. **Performance Impact**
   - Benchmark before/after migration
   - Monitor query performance
   - Optimize if needed

## Tracking Progress

Use GitHub Issues with labels:
- `sql-migration` - For SQL migration tasks
- `test-restoration` - For restoring skipped tests
- `test-improvement` - For test quality improvements
- `ci-cd` - For pipeline enhancements

Create a GitHub Project board to track:
1. To Do
2. In Progress
3. In Review
4. Done

---

**Created**: January 9, 2025
**Last Updated**: January 9, 2025
**Owner**: Backend Team
**Status**: Active Development