package main

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// validateDeployment is the stub for future license enforcement.
// Wire all startup checks through here; today it only ensures the
// deployment record exists.
func validateDeployment(ctx context.Context, db *pgxpool.Pool) error {
	var exists bool
	if err := db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM deployment)`,
	).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		deployID := uuid.New()
		if _, err := db.Exec(ctx,
			`INSERT INTO deployment (deployment_id) VALUES ($1)`, deployID,
		); err != nil {
			return err
		}
		slog.Info("new deployment registered", "deployment_id", deployID)
	}
	return nil
}
