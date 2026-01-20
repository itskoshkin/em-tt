package testutils

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	TestDBName     = "testdb"
	TestDBUser     = "testuser"
	TestDBPassword = "testpass"
)

// PostgresContainer wraps testcontainers postgres container
type PostgresContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
	DB        *gorm.DB
}

// SetupPostgresContainer creates a new PostgreSQL container for testing
func SetupPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	container, err := tcpostgres.Run(ctx,
		"postgres:15-alpine",
		tcpostgres.WithDatabase(TestDBName),
		tcpostgres.WithUsername(TestDBUser),
		tcpostgres.WithPassword(TestDBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, mappedPort.Port(), TestDBUser, TestDBPassword, TestDBName)

	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Port(),
		DB:        db,
	}, nil
}

// RunMigrations executes the database migrations
func (pc *PostgresContainer) RunMigrations(ctx context.Context) error {
	// Create extension
	if err := pc.DB.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		return fmt.Errorf("failed to create pgcrypto extension: %w", err)
	}

	// Create subscriptions table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS subscriptions (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			service_name text NOT NULL,
			price integer NOT NULL CHECK (price >= 0),
			user_id uuid NOT NULL,
			start_date date NOT NULL,
			end_date date NULL,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now(),
			deleted_at timestamptz NULL
		);
		CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
		CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name);
		CREATE INDEX IF NOT EXISTS idx_subscriptions_deleted_at ON subscriptions(deleted_at);
	`
	if err := pc.DB.Exec(createTableSQL).Error; err != nil {
		return fmt.Errorf("failed to create subscriptions table: %w", err)
	}

	return nil
}

// Cleanup removes all data from tables
func (pc *PostgresContainer) Cleanup(ctx context.Context) error {
	return pc.DB.Exec("TRUNCATE TABLE subscriptions").Error
}

// Teardown stops and removes the container
func (pc *PostgresContainer) Teardown(ctx context.Context) error {
	if pc.Container != nil {
		return pc.Container.Terminate(ctx)
	}
	return nil
}

// ConnectionString returns the PostgreSQL connection string
func (pc *PostgresContainer) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pc.Host, pc.Port, TestDBUser, TestDBPassword, TestDBName)
}
