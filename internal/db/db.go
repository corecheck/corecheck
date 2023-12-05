package db

import (
	"github.com/corecheck/corecheck/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dbURI := "host=" + config.Config.DB.Host + " user=" + config.Config.DB.User + " password=" + config.Config.DB.Password + " dbname=" + config.Config.DB.Name + " port=" + config.Config.DB.Port + " sslmode=disable TimeZone=Europe/Paris"
	database, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	// TODO: move to cmd/migration/main.go?
	// database.AutoMigrate(&PR{}, &BenchmarkResult{}, &CoverageReport{}, &Job{}, &CoverageLine{})
	DB = database
}
