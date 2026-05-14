package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/telemetry"
)

const (
	bitcoinDataZipURL = "https://github.com/bitcoin-data/github-metadata-backup-bitcoin-bitcoin/archive/refs/heads/master.zip"
	dest              = "/tmp/data"
)

func handleMetrics(ctx context.Context) (string, error) {
	extractedPath, err := DownloadAndExtract(bitcoinDataZipURL, dest)
	if err != nil {
		log.Fatalf("download: %v", err)
	}

	bc := BitcoinCoreData{Path: extractedPath}
	bc.AddConsumers(&NumberOfPullsConsumer{}, &NumberOfIssuesConsumer{})
	bc.Run()

	logGroupName := os.Getenv("GITHUB_EVENTS_LOG_GROUP")
	lastRunParam := os.Getenv("GITHUB_EVENTS_LAST_RUN_PARAM")
	awsRegion := os.Getenv("TELEMETRY_CLOUDWATCH_REGION")
	if awsRegion == "" {
		awsRegion = os.Getenv("AWS_REGION")
	}
	if awsRegion == "" {
		awsRegion = os.Getenv("AWS_DEFAULT_REGION")
	}

	if logGroupName == "" || lastRunParam == "" || awsRegion == "" {
		log.Println("stats: GITHUB_EVENTS_LOG_GROUP / GITHUB_EVENTS_LAST_RUN_PARAM / region not set; skipping event stream")
		return "OK", nil
	}

	ssmClient, err := newSSMClient(awsRegion)
	if err != nil {
		log.Printf("stats: could not create SSM client: %v; skipping event stream", err)
		return "OK", nil
	}

	lastRunTime, err := GetLastRunTime(ssmClient, lastRunParam)
	if err != nil {
		log.Printf("stats: could not read last run time from SSM: %v; skipping event stream", err)
		return "OK", nil
	}

	// Capture run time before processing so events emitted during this run are caught next time.
	runTime := time.Now().UTC()

	cwWriter, err := NewCWLogsWriter(awsRegion, logGroupName)
	if err != nil {
		log.Printf("stats: could not create CW Logs writer: %v; skipping event stream", err)
		return "OK", nil
	}

	producer := NewEventStreamProducer(extractedPath, lastRunTime, cwWriter)
	if err := producer.Run(); err != nil {
		log.Printf("stats: event stream write error: %v", err)
		return "OK", nil
	}

	if err := SetLastRunTime(ssmClient, lastRunParam, runTime); err != nil {
		log.Printf("stats: could not store last run time in SSM: %v", err)
	}

	return "OK", nil
}

func main() {
	if err := telemetry.ConfigureDefaultFromEnv(); err != nil {
		log.Fatalf("Error configuring telemetry: %s", err)
	}

	lambda.Start(handleMetrics)
}
