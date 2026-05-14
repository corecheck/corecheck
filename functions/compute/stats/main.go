package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/telemetry"
)

const (
	bitcoinDataURL = "https://github.com/bitcoin-data/github-metadata-backup-bitcoin-bitcoin"
	dest           = "/tmp/data"
)

func handleMetrics(ctx context.Context) (string, error) {
	// Clone the bitcoin-data repository (shallow, depth=100).
	headSHA, err := GitClone(bitcoinDataURL, dest)
	if err != nil {
		log.Fatalf("git clone: %v", err)
	}

	// Run existing gauge-metric consumers over the full dataset (unchanged).
	bc := BitcoinCoreData{Path: dest}
	bc.AddConsumers(&NumberOfPullsConsumer{}, &NumberOfIssuesConsumer{})
	bc.Run()

	// Event stream — skip silently if env vars are not configured.
	logGroupName := os.Getenv("GITHUB_EVENTS_LOG_GROUP")
	shaSsmParam := os.Getenv("GITHUB_EVENTS_SHA_PARAM")
	awsRegion := os.Getenv("TELEMETRY_CLOUDWATCH_REGION")
	if awsRegion == "" {
		awsRegion = os.Getenv("AWS_REGION")
	}
	if awsRegion == "" {
		awsRegion = os.Getenv("AWS_DEFAULT_REGION")
	}

	if logGroupName == "" || shaSsmParam == "" || awsRegion == "" {
		log.Println("stats: GITHUB_EVENTS_LOG_GROUP / GITHUB_EVENTS_SHA_PARAM / region not set; skipping event stream")
		return "OK", nil
	}

	ssmClient, err := newSSMClient(awsRegion)
	if err != nil {
		log.Printf("stats: could not create SSM client: %v; skipping event stream", err)
		return "OK", nil
	}

	prevSHA, err := GetGitSHA(ssmClient, shaSsmParam)
	if err != nil {
		log.Printf("stats: could not read git SHA from SSM: %v; skipping event stream", err)
		return "OK", nil
	}

	// Determine which files changed since the previous run.
	var changedFiles []string
	if prevSHA != "" {
		files, ok := GitChangedFiles(dest, prevSHA)
		if ok {
			changedFiles = files
		} else {
			log.Printf("stats: prev SHA %s not in shallow history; falling back to 14-day scan", prevSHA)
			// changedFiles stays nil → EventStreamProducer processes all files
		}
	}

	cwWriter, err := NewCWLogsWriter(awsRegion, logGroupName)
	if err != nil {
		log.Printf("stats: could not create CW Logs writer: %v; skipping event stream", err)
		return "OK", nil
	}

	producer := NewEventStreamProducer(dest, prevSHA, changedFiles, cwWriter)
	if err := producer.Run(); err != nil {
		log.Printf("stats: event stream write error: %v", err)
		return "OK", nil
	}

	if err := SetGitSHA(ssmClient, shaSsmParam, headSHA); err != nil {
		log.Printf("stats: could not store git SHA in SSM: %v", err)
	}

	return "OK", nil
}

func main() {
	if err := telemetry.ConfigureDefaultFromEnv(); err != nil {
		log.Fatalf("Error configuring telemetry: %s", err)
	}

	lambda.Start(handleMetrics)
}
