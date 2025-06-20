# CLAUDE.md - Project Context for D&D Game

## SQL Migration & Test Restoration (January 2025)

### Quick Reference
When working on backend tests or SQL-related issues, check these resources:

#### Key Documents
- **SQL Migration Guide**: `backend/SQL_MIGRATION_TASKS.md` - Detailed migration instructions
- **Test Improvements Roadmap**: `backend/BACKEND_TEST_IMPROVEMENTS.md` - Comprehensive test strategy
- **Migration Summary**: `backend/SQL_MIGRATION_SUMMARY.md` - Current status and metrics
- **CI/CD Progress**: `CI_CD_PROGRESS.md` - Pipeline status and achievements

#### GitHub Issues (Created Jan 9, 2025)
- **Meta Tracking**: Issue #32 - Overall progress tracking
- **SQL Migration**: Issues #20-#30 (11 repositories, ~290 queries)
- **Test Restoration**: Issue #31 (20 skipped test files)

#### Helper Scripts
- **Find SQL placeholders**: `backend/fix_sql_placeholders.sh`
- **Run backend tests**: `cd backend && go test ./...`
- **Check coverage**: `cd backend && go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out`

### Key Patterns

#### Database-Agnostic SQL
```go
// Use ? placeholders instead of $1, $2
query := `SELECT * FROM table WHERE id = ?`

// For *DB wrapper
err := r.db.QueryRowContextRebind(ctx, query, id).Scan(&result)

// For *sqlx.DB
query = r.db.Rebind(query)
err := r.db.QueryRowContext(ctx, query, id).Scan(&result)
```

#### Test Expectations
```go
// SQL expectations use \? instead of \$1
mock.ExpectQuery(`SELECT .* WHERE id = \?`).WithArgs(id)
```

### Common Issues & Fixes
1. **SQL Compatibility**: Tests fail with PostgreSQL syntax in CI/CD but work locally with SQLite
   - Solution: Use ? placeholders and Rebind methods
   
2. **Skipped Tests**: Files renamed to .skip to avoid compilation errors
   - Solution: Fix model fields (Type→ActionType, Turn→CurrentTurn) and restore

3. **Mock Interfaces**: Mocks don't match updated interfaces
   - Solution: Regenerate mocks or update manually

### Test Coverage Status
- **Current**: ~84% for services (with some tests skipped)
- **Target**: 84%+ overall after restoration
- **Critical Path**: All handler integration tests passing

### Important Context
- CI/CD uses PostgreSQL, local development uses SQLite
- Database compatibility is critical for stable pipeline
- Test restoration blocked on SQL migration completion
- 11 repositories need SQL updates (~290 queries total)

### Sensitive Tokens & Credentials
- Sonar Token: `cb8b0a7e4c6522e4aa1d016dd5785c7352b4e727`

## MCP (Model Context Protocol) Configuration

### GitHub MCP Server
The project has GitHub MCP server configured for enhanced GitHub operations.

#### Setup (already configured)
```bash
# Add GitHub MCP server with authentication
claude mcp add-json github '{"command": "npx", "args": ["-y", "@modelcontextprotocol/server-github"], "env": {"GITHUB_PERSONAL_ACCESS_TOKEN": "YOUR_GITHUB_TOKEN_HERE"}}'

# List configured servers
claude mcp list

# Check server details
claude mcp get github
```

#### Usage Patterns
When GitHub MCP tools are available (after Claude Code restart), prefer them over `gh` CLI for:

1. **Complex API Operations**
   - Instead of: `gh api repos/ctclostio/DnD-Game/pulls`
   - Use: `mcp_github__` prefixed tools for better performance

2. **Batch Operations**
   - Creating multiple issues/PRs
   - Managing labels across repositories
   - Bulk repository operations

3. **Advanced Queries**
   - Repository insights and analytics
   - Complex search operations
   - Workflow management

#### Current MCP Servers
- **github**: GitHub operations with personal access token

#### Future MCP Considerations
- **mcp_postgres**: Direct database access for complex queries
- **mcp_sonarcloud**: Direct SonarCloud integration
- **mcp_docker**: Container management for development

Note: MCP tools require Claude Code restart to become available after configuration.