package controller

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/vrstep/wawatch-backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MockAnyTime is used to match any time value in sqlmock
type MockAnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a MockAnyTime) Match(v interface{}) bool {
	_, ok := v.(time.Time)
	return ok
}

// SetupTestDB creates a mock DB connection for testing
func SetupTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	// sqlmock.New creates a new mock database connection and a mock object
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// Use GORM's postgres driver with the mock DB
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	}), &gorm.Config{
		// Disable logging or set to silent for cleaner test output
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	// Replace the global DB instance for tests
	originalDB := config.DB
	config.DB = gormDB

	// Return the mock object and a cleanup function
	cleanup := func() {
		config.DB = originalDB // Restore original DB
		mockDB.Close()
	}

	return mock, cleanup
}

// SetupGin sets up a test Gin engine
func SetupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

// Helper to escape SQL query for sqlmock
func EscapeQuery(query string) string {
	return regexp.QuoteMeta(query)
}
