package health

import (
	"context"
	"fmt"

	"github.com/ctclostio/DnD-Game/backend/internal/database"
)

// DBChecker verifies database connectivity.
// by pinging the database connection.
type DBChecker struct {
	DB *database.DB
}

func (d *DBChecker) Name() string { return "database" }

func (d *DBChecker) Check(ctx context.Context) error {
	if d.DB == nil {
		return fmt.Errorf("db not initialized")
	}
	return d.DB.PingContext(ctx)
}
