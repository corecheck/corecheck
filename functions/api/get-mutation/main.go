package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	commit, err := db.GetLatestMutationResultCommit()
	if err != nil {
		log.Error(err)
		return err
	}

	data, err := getMutationJSONFromS3(commit)
	if err != nil {
		log.Error(err)
		return err
	}

	return c.JSON(200, data)
}

func getMutationJSONFromS3(commit string) ([]byte, error) {
	url := os.Getenv("BUCKET_DATA_URL") + "/master/" + commit + "/mutation.json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JSON: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JSON, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return json.Marshal(data)
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
	api.Start(e)
}
