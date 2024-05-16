package main

import (
	"context"
	"fmt"
	"os"
	"time"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	"github.com/aws/aws-lambda-go/lambda"
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
	bc.AddConsumers(&NumberOfPullsConsumer{}, &UniqueAuthorsConsumer{}, &PullsByUserConsumer{}, &PullsByLabelConsumer{}, &TotalCommentsAndReviewsByPullConsumer{})
	bc.AddConsumers(&NumberOfIssuesConsumer{}, &UniqueIssueUsersConsumer{}, &IssuesByUserConsumer{}, &IssuesByLabelConsumer{}, &TotalCommentsIssueConsumer{})
	bc.Run()

	return "OK", nil
}

func main() {
	lambda.Start(ddlambda.WrapFunction(handleMetrics, &ddlambda.Config{
		DebugLogging:    true,
		Site:            "datadoghq.eu",
		BatchInterval:   time.Millisecond * 500,
		EnhancedMetrics: true,
	}))
}
