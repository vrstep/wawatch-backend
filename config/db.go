package config

import (
	"database/sql" // Import database/sql
	"log"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres" // Alias migrate's postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"                     // Import file source driver
	gormpostgres "gorm.io/driver/postgres"                                   // Import GORM's postgres driver
	"gorm.io/gorm"
)

var DB *gorm.DB

// func ConnectDB() {
// 	// Ensure sslmode is explicitly handled if needed by your setup
// 	dsn := "postgres://postgres:postgres@localhost:5432/wawatchdb?sslmode=disable"
// 	// Use the imported GORM postgres driver here
// 	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database with GORM: %v", err)
// 	}

// 	// Get the underlying sql.DB connection for migrate
// 	sqlDB, err := db.DB()
// 	if err != nil {
// 		log.Fatalf("Failed to get underlying sql.DB: %v", err)
// 	}

// 	// Run migrations using the existing DB connection
// 	log.Println("Running database migrations...")
// 	err = runMigrations(sqlDB) // Pass the sql.DB instance
// 	if err != nil {
// 		log.Fatalf("Failed to run database migrations: %v", err)
// 	}
// 	log.Println("Database migrations completed successfully.")

// 	// --- AutoMigrate is commented out ---
// 	// ... (keep AutoMigrate commented out)

// 	DB = db
// }

func ConnectDB() {
	// GORM connection
	dsn := "postgres://postgres:postgres@localhost:5432/wawatchdb?sslmode=disable"
	gormDB, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database with GORM: %v", err)
	}

	// Run migrations using a **separate** sql.DB
	rawDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open sql.DB for migration: %v", err)
	}
	defer rawDB.Close() // okay to close, it's not used by GORM

	log.Println("Running database migrations...")
	if err := runMigrations(rawDB); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	log.Println("Database migrations completed successfully.")

	DB = gormDB
}

// runMigrations function executes the database migrations using an existing sql.DB
func runMigrations(sqlDB *sql.DB) error {
	// Use the aliased migrate postgres driver here
	driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	if err != nil {
		return err
	}

	// Point to your existing directory containing migration files
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations", // Correct path to your migrations
		"wawatchdb",            // Database name/identifier (keep as "postgres")
		driver,                 // The database instance driver
	)
	if err != nil {
		return err
	}

	// Apply all available "up" migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	// Check for migration source/database errors after running
	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		log.Printf("Migration source error on close: %v", sourceErr)
	}
	if dbErr != nil {
		log.Printf("Migration database error on close: %v", dbErr)
	}

	if err == migrate.ErrNoChange {
		log.Println("No new migrations to apply.")
		return nil // Not an actual error
	}

	return err // Return the original error from m.Up() if it wasn't ErrNoChange
}
