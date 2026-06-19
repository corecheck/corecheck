package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/corecheck/corecheck/internal/datadogexport"
)

func main() {
	cfg := datadogexport.Config{}
	var (
		from       string
		to         string
		chunk      string
		routes     string
		dashboards string
	)

	flag.StringVar(&cfg.EnvFile, "env-file", ".env", "Path to the Datadog .env file")
	flag.StringVar(&cfg.InventoryDir, "inventory", "", "Path to the dashboard inventory directory")
	flag.StringVar(&cfg.OutputDir, "output-dir", "", "Directory that will receive dashboards, manifest.json, and records.jsonl")
	flag.StringVar(&from, "from", "", "UTC start time for historical export (RFC3339 or YYYY-MM-DD)")
	flag.StringVar(&to, "to", "", "UTC end time for historical export (RFC3339 or YYYY-MM-DD)")
	flag.StringVar(&chunk, "chunk", "24h", "Per-request Datadog query window")
	flag.StringVar(&routes, "routes", "", "Comma-separated route filters (example: /tests,/benchmarks)")
	flag.StringVar(&dashboards, "dashboards", "", "Comma-separated dashboard ID filters")
	flag.StringVar(&cfg.DatadogSite, "site", "", "Datadog site (for example datadoghq.eu)")
	flag.IntVar(&cfg.MaxQueries, "max-queries", 0, "Limit the number of extracted metric queries to export")
	flag.Parse()

	if cfg.InventoryDir == "" || cfg.OutputDir == "" {
		flag.Usage()
		os.Exit(2)
	}

	if from != "" || to != "" {
		if from == "" || to == "" {
			log.Fatal("both -from and -to must be provided when exporting history")
		}

		parsedFrom, err := parseTime(from)
		if err != nil {
			log.Fatalf("invalid -from value: %v", err)
		}

		parsedTo, err := parseTime(to)
		if err != nil {
			log.Fatalf("invalid -to value: %v", err)
		}

		cfg.From = parsedFrom
		cfg.To = parsedTo
	}

	parsedChunk, err := time.ParseDuration(chunk)
	if err != nil {
		log.Fatalf("invalid -chunk value: %v", err)
	}
	cfg.Chunk = parsedChunk
	cfg.RouteFilter = splitCSV(routes)
	cfg.DashboardFilter = splitCSV(dashboards)

	result, err := datadogexport.Run(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}

	if result.RecordCount > 0 {
		fmt.Printf("exported %d dashboards, %d metric queries, and %d records into %s\n", result.DashboardCount, result.QueryCount, result.RecordCount, result.OutputDir)
		return
	}

	fmt.Printf("exported %d dashboards and %d metric queries into %s\n", result.DashboardCount, result.QueryCount, result.OutputDir)
}

func parseTime(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("expected RFC3339 or YYYY-MM-DD, got %q", value)
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	rawValues := strings.Split(value, ",")
	values := make([]string, 0, len(rawValues))
	for _, item := range rawValues {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		values = append(values, item)
	}

	return values
}
