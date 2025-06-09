---
name: SQL Migration Task
about: Track SQL parameter placeholder migration for a repository
title: '[SQL Migration] Update {repository_name} to use database-agnostic SQL'
labels: 'enhancement, technical-debt, backend'
assignees: ''

---

## Repository to Migrate
**File**: `backend/internal/database/{repository_name}.go`
**Number of queries**: {X} queries need updating

## Description
Update all SQL queries in this repository to use database-agnostic `?` placeholders instead of PostgreSQL-specific `$1, $2` syntax.

## Tasks
- [ ] Replace all `$1, $2, etc.` with `?` placeholders
- [ ] Update parameter order if needed (? placeholders are positional)
- [ ] Add appropriate rebind calls:
  - For `*DB` wrapper: Use `QueryRowContextRebind`, `ExecContextRebind`, etc.
  - For `*sqlx.DB`: Add `query = r.db.Rebind(query)` before execution
- [ ] Update test file if it exists: `{repository_name}_test.go`
- [ ] Run repository tests to verify changes
- [ ] Test with both SQLite (local) and PostgreSQL if possible

## Example Migration

### Before:
```go
query := `SELECT * FROM table WHERE id = $1 AND status = $2`
err := r.db.QueryRowContext(ctx, query, id, status).Scan(&result)
```

### After:
```go
query := `SELECT * FROM table WHERE id = ? AND status = ?`
err := r.db.QueryRowContextRebind(ctx, query, id, status).Scan(&result)
```

## Testing
```bash
# Run specific repository tests
go test ./internal/database -run Test{RepositoryName}

# Run all database tests
go test ./internal/database -v
```

## References
- [SQL Migration Tasks Documentation](../../backend/SQL_MIGRATION_TASKS.md)
- [Helper Script](../../backend/fix_sql_placeholders.sh)
- [Original PR](#) <!-- Link to the main SQL compatibility PR -->

## Acceptance Criteria
- [ ] All queries use `?` placeholders
- [ ] Tests pass locally with SQLite
- [ ] No PostgreSQL-specific syntax remains
- [ ] Code follows the migration guide patterns