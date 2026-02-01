package database

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrator handles database migrations
type Migrator struct {
	db     *DB
	logger *zap.Logger
}

// NewMigrator creates a new Migrator instance
func NewMigrator(db *DB, logger *zap.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// RunMigrations executes all pending migrations
func RunMigrations(ctx context.Context, db *DB) error {
	logger, _ := zap.NewProduction()
	migrator := NewMigrator(db, logger)
	return migrator.Migrate(ctx)
}

// Migrate runs all pending migrations in order
func (m *Migrator) Migrate(ctx context.Context) error {
	// Create schema_migrations table if not exists
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get all migration files
	migrations, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if applied[migration.Name] {
			m.logger.Debug("skipping already applied migration", zap.String("name", migration.Name))
			continue
		}

		m.logger.Info("applying migration", zap.String("name", migration.Name))

		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
		}

		m.logger.Info("migration applied successfully", zap.String("name", migration.Name))
	}

	return nil
}

// Migration represents a single migration file
type Migration struct {
	Name    string
	Content string
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_schema_migrations_name ON schema_migrations(name);
	`
	_, err := m.db.Pool().Exec(ctx, query)
	return err
}

// getAppliedMigrations returns a map of already applied migration names
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := m.db.Pool().Query(ctx, "SELECT name FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}

	return applied, rows.Err()
}

// getMigrationFiles reads all .sql files from the migrations directory
func (m *Migrator) getMigrationFiles() ([]Migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := migrationsFS.ReadFile(filepath.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		// Extract name without .sql extension
		name := strings.TrimSuffix(entry.Name(), ".sql")
		migrations = append(migrations, Migration{
			Name:    name,
			Content: string(content),
		})
	}

	// Sort migrations by name (which includes the numeric prefix)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}

// applyMigration applies a single migration within a transaction
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.Pool().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Execute the migration SQL
	if _, err := tx.Exec(ctx, migration.Content); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record the migration as applied
	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Rollback rolls back the last applied migration (for development use)
func (m *Migrator) Rollback(ctx context.Context) error {
	// Get the last applied migration
	var name string
	err := m.db.Pool().QueryRow(ctx,
		"SELECT name FROM schema_migrations ORDER BY applied_at DESC LIMIT 1",
	).Scan(&name)
	if err != nil {
		if err == pgx.ErrNoRows {
			m.logger.Info("no migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	m.logger.Warn("rollback is not implemented - manual intervention required",
		zap.String("last_migration", name),
	)

	return nil
}

// Status returns the current migration status
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	migrations, err := m.getMigrationFiles()
	if err != nil {
		return nil, err
	}

	var statuses []MigrationStatus
	for _, migration := range migrations {
		statuses = append(statuses, MigrationStatus{
			Name:    migration.Name,
			Applied: applied[migration.Name],
		})
	}

	return statuses, nil
}

// MigrationStatus represents the status of a single migration
type MigrationStatus struct {
	Name    string
	Applied bool
}
