package testutil

import (
	"testing"

	"github.com/naiba/bonds/internal/config"
	"github.com/naiba/bonds/internal/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return setupSQLiteTestDB(t, true)
}

// SetupTestDBWithFKConstraints creates a test DB with foreign-key constraints
// actually present in the schema and enforced. Use this for tests that need to
// catch FK violations (e.g. cascade-delete regressions). The default
// SetupTestDB strips FK constraints during migration, so `PRAGMA foreign_keys
// = ON` does nothing there — only this helper provides production-like FK
// behavior under SQLite.
func SetupTestDBWithFKConstraints(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupSQLiteTestDB(t, false)
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}
	return db
}

func setupSQLiteTestDB(t *testing.T, disableFKConstraintMigration bool) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: disableFKConstraintMigration,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	// SQLite :memory: creates a separate database per connection.
	// Limit to 1 open connection so all queries hit the same in-memory DB.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	return db
}

func TestJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		Secret:     "test-secret-key",
		ExpiryHrs:  24,
		RefreshHrs: 168,
	}
}
