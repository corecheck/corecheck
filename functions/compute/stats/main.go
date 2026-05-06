package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/telemetry"
)

const (
	bitcoinDataURL = "https://github.com/bitcoin-data/github-metadata-backup-bitcoin-bitcoin/archive/refs/heads/master.zip"
	dest           = "/tmp/data"
)

func handleMetrics(ctx context.Context) (string, error) {

	// Download zip file
	err := DownloadFile(bitcoinDataURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unzip file
	err = Unzip("/tmp/bitcoin-data.zip", dest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bc := BitcoinCoreData{Path: dest + "/github-metadata-backup-bitcoin-bitcoin-master"}
	bc.AddConsumers(&NumberOfPullsConsumer{}, &UniqueAuthorsConsumer{}, &PullsByUserConsumer{}, &PullsByLabelConsumer{}, &TotalCommentsAndReviewsByPullConsumer{}, &PeriodCommentsAndReviewsByPullConsumer{})
	bc.AddConsumers(&NumberOfIssuesConsumer{}, &UniqueIssueUsersConsumer{}, &IssuesByUserConsumer{}, &IssuesByLabelConsumer{}, &TotalCommentsIssueConsumer{}, &PeriodCommentsIssueConsumer{})
	bc.Run()

	return "OK", nil
}

func main() {
	if err := telemetry.ConfigureDefaultFromEnv(); err != nil {
		log.Fatalf("Error configuring telemetry: %s", err)
	}

	lambda.Start(handleMetrics)
}
