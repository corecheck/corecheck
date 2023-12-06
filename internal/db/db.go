package db

import (
	"github.com/corecheck/corecheck/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg config.DatabaseConfig) error {
	dbURI := "host=" + cfg.Database.Host + " user=" + cfg.Database.User + " password=" + cfg.Database.Password + " dbname=" + cfg.Database.Name + " port=" + cfg.Database.Port + " sslmode=disable TimeZone=Europe/Paris"
	database, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	if err != nil {
		return err
	}

	// TODO: move to cmd/migration/main.go?
	// database.AutoMigrate(&PR{}, &BenchmarkResult{}, &CoverageReport{}, &Job{}, &CoverageLine{})
	DB = database

	return nil
}
