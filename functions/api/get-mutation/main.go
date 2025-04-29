package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/corecheck/corecheck/internal/api"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/labstack/echo/v4"
)

var (
	cfg = Config{}
	log = logger.New()
)

func getLatestMutation(c echo.Context) error {
	result, err := db.GetLatestCompletedMutationResult()
	if err != nil {
		log.Error(err)
		return err
	}

	url := os.Getenv("BUCKET_DATA_URL") + "/master/" + result.Commit + "/mutation.json"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch JSON: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JSON, status: %d", resp.StatusCode)
	}

	return c.Stream(200, "application/json", resp.Body)
}

func getLatestMutationMeta(c echo.Context) error {
	result, err := db.GetLatestCompletedMutationResult()
	if err != nil {
		log.Error(err)
		return err
	}

	return c.JSON(200, result)
}

func main() {
	log.Debug("Loading config...")
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	log.Debug("Connecting to database...")
	if err := db.Connect(cfg.DatabaseConfig); err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	e := api.New()
	e.GET("/mutations", getLatestMutation)
	e.GET("/mutations/meta", getLatestMutationMeta)
	api.Start(e)
}
