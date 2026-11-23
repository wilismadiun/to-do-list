package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"

	"user-service/src/Domains/entities"
	domainerr "user-service/src/Domains/errors"
	"user-service/src/Infrastructures/database"
	infrarepo "user-service/src/Infrastructures/repository"
)

func TestUserRepositoryPostgres(t *testing.T) {
	// Load only the test env file. Existing CI environment variables take precedence.
	_ = godotenv.Load("../../.env.test")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		t.Skip("DB_NAME is not set; configure .env.test to run PostgreSQL integration tests")
	}
	if !strings.Contains(strings.ToLower(dbName), "test") {
		t.Fatalf("refusing to run integration tests: DB_NAME %q must contain 'test'", dbName)
	}

	cfg := database.Config{Host: os.Getenv("DB_HOST"), Port: envOr("DB_PORT", "5432"), User: os.Getenv("DB_USER"), Password: os.Getenv("DB_PASSWORD"), Name: dbName, SSLMode: envOr("DB_SSLMODE", "disable")}
	if cfg.Host == "" || cfg.User == "" {
		t.Fatal("DB_HOST and DB_USER must be configured for integration tests")
	}
	db, err := database.NewPostgres(cfg)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, goose.SetDialect("postgres"))
	migrationDir, err := filepath.Abs("../../migrations")
	require.NoError(t, err)
	require.NoError(t, goose.Up(sqlDB, migrationDir))
	require.NoError(t, db.Exec("TRUNCATE TABLE users").Error)

	repo := infrarepo.NewGormUserRepository(db)
	ctx := context.Background()
	id := uuid.New()
	username := fmt.Sprintf("integration-%s", id.String()[:8])
	user := &entities.User{ID: id, Username: username, PasswordHash: "bcrypt-hash-for-repository-test"}
	require.NoError(t, repo.Create(ctx, user))

	byUsername, err := repo.FindByUsername(ctx, username)
	require.NoError(t, err)
	require.Equal(t, id, byUsername.ID)
	byID, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, username, byID.Username)
	require.Equal(t, id, byID.ID) // UUID persisted and retrieved without alteration.

	duplicate := &entities.User{ID: uuid.New(), Username: username, PasswordHash: "another-hash"}
	require.ErrorIs(t, repo.Create(ctx, duplicate), domainerr.ErrUserAlreadyExists)
	_, err = repo.FindByUsername(ctx, "does-not-exist-"+id.String())
	require.ErrorIs(t, err, domainerr.ErrUserNotFound)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
