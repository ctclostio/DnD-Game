#!/bin/bash

# Script to help identify and fix PostgreSQL parameter placeholders ($1, $2, etc.) 
# in repository files to use ? placeholders with rebind for database compatibility

echo "Finding repository files with PostgreSQL parameter syntax..."
echo "================================================="

# Find all repository files using $1, $2, etc. syntax
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel 2>/dev/null || echo "$SCRIPT_DIR")"

for file in "$REPO_ROOT"/backend/internal/database/*_repository.go; do
    if [ -f "$file" ] && grep -q '\$[0-9]' "$file"; then
        basename=$(basename "$file")
        count=$(grep -c '\$[0-9]' "$file")
        echo "$basename: $count queries with PostgreSQL syntax"
    fi
done

echo ""
echo "Files that need updating:"
echo "========================"
grep -l '\$[0-9]' "$REPO_ROOT"/backend/internal/database/*_repository.go 2>/dev/null | xargs -I {} basename {}

echo ""
echo "To fix a file:"
echo "1. Replace $1, $2, etc. with ?"
echo "2. For sqlx.DB repos: Add 'query = r.db.Rebind(query)' before execution"
echo "3. For DB wrapper repos: Use QueryRowContextRebind, ExecContextRebind methods"
echo ""
echo "Example transformation:"
echo "  FROM: WHERE id = \$1 AND user_id = \$2"
echo "  TO:   WHERE id = ? AND user_id = ?"