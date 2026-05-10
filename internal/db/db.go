package db

import (
	"os"

	"github.com/corecheck/corecheck/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg config.DatabaseConfig) error {
	dbURI := "host=" + cfg.Database.Host + " user=" + cfg.Database.User + " password=" + cfg.Database.Password + " dbname=" + cfg.Database.Name + " port=" + cfg.Database.Port + " sslmode=" + cfg.Database.SSLMode + " TimeZone=Europe/Paris"
	database, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	if err != nil {
		return err
	}

	// For local dev (e.g. LocalStack), run migrations on connect so tables exist.
	if os.Getenv("AUTO_MIGRATE") == "true" {
		_ = database.AutoMigrate(
			&CoverageReport{},
			&CoverageFileHunk{},
			&CoverageFileHunkLine{},
			&BenchmarkResult{},
			&PR{},
			&MutationResult{},
		)
	}

	DB = database

	return nil
}
