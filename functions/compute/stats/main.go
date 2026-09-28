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

	pulls := &NumberOfPullsConsumer{}
	issues := &NumberOfIssuesConsumer{}
	bc := BitcoinCoreData{Path: dest + "/github-metadata-backup-bitcoin-bitcoin-master"}
	bc.AddConsumers(pulls, &UniqueAuthorsConsumer{}, &PullsByUserConsumer{}, &PullsByLabelConsumer{}, &TotalCommentsAndReviewsByPullConsumer{})
	bc.AddConsumers(issues, &UniqueIssueUsersConsumer{}, &IssuesByUserConsumer{}, &IssuesByLabelConsumer{}, &TotalCommentsIssueConsumer{})
	bc.Run()

	// Datadog gauges were already sent by the consumers. CloudWatch capture must
	// not fail the lambda, or a retry would be the only extra effect and the
	// live Datadog dashboards would still have been updated.
	if err := putOpenCountMetrics(pulls.Open, issues.Open); err != nil {
		fmt.Printf("stats: cloudwatch open counts failed: %v\n", err)
	}
	if err := writeGitHubEventStream(); err != nil {
		fmt.Printf("stats: github event stream failed: %v\n", err)
	}

	return "OK", nil
}

func writeGitHubEventStream() error {
	logGroupName := os.Getenv("GITHUB_EVENTS_LOG_GROUP")
	lastRunParam := os.Getenv("GITHUB_EVENTS_LAST_RUN_PARAM")
	awsRegion := cloudWatchRegion()
	if logGroupName == "" || lastRunParam == "" || awsRegion == "" {
		fmt.Println("stats: GITHUB_EVENTS_LOG_GROUP / GITHUB_EVENTS_LAST_RUN_PARAM / region not set; skipping event stream")
		return nil
	}

	ssmClient, err := newSSMClient(awsRegion)
	if err != nil {
		return fmt.Errorf("ssm client: %w", err)
	}

	lastRunTime, err := GetLastRunTime(ssmClient, lastRunParam)
	if err != nil {
		return fmt.Errorf("read last run: %w", err)
	}
	if lastRunTime.IsZero() {
		fmt.Println("stats: first run — github events cutoff is the last 24 hours")
	} else {
		fmt.Printf("stats: last run time: %s\n", lastRunTime.Format(time.RFC3339))
	}

	runTime := time.Now().UTC()
	cwWriter, err := NewCWLogsWriter(awsRegion, logGroupName)
	if err != nil {
		return fmt.Errorf("log writer: %w", err)
	}

	producer := NewEventStreamProducer(dest+"/github-metadata-backup-bitcoin-bitcoin-master", lastRunTime, cwWriter)
	if err := producer.Run(); err != nil {
		return err
	}
	if err := SetLastRunTime(ssmClient, lastRunParam, runTime); err != nil {
		return fmt.Errorf("store last run: %w", err)
	}
	fmt.Printf("stats: stored last run time %s\n", runTime.Format(time.RFC3339))
	return nil
}

func main() {
	lambda.Start(ddlambda.WrapFunction(handleMetrics, &ddlambda.Config{
		DebugLogging:    true,
		Site:            "datadoghq.eu",
		BatchInterval:   time.Millisecond * 500,
		EnhancedMetrics: true,
	}))
}
