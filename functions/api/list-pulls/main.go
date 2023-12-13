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

func listPulls(c echo.Context) error {
	title := c.QueryParam("title")
	page := c.QueryParam("page")
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		log.Warn(err)
		pageNum = 1
	}
	pulls, err := db.ListPulls(db.SearchPRsOptions{
		Title: title,
		Page:  pageNum,
	})
	if err != nil {
		log.Error(err)
		return err
	}

	return c.JSON(200, pulls)
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
	e.GET("/pulls", listPulls)
	api.Start(e)
}
