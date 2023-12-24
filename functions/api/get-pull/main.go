package main

import (
	"strconv"

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

func getPull(c echo.Context) error {
	pullNumber := c.Param("number")
	pullNumberInt, err := strconv.Atoi(pullNumber)
	if err != nil {
		log.Errorf("Error converting pull number to int: %s", err)
		return err
	}
	pull, err := db.GetPR(pullNumberInt)
	if err != nil {
		log.Error(err)
		return err
	}

	return c.JSON(200, pull)
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
	e.GET("/pulls/:number", getPull)
	api.Start(e)
}
