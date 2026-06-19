package main

import (
	"encoding/json"
	"fmt"
	stdlog "log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/corecheck/corecheck/internal/db"
)

type benchmarkLogRecord struct {
	BenchmarkName    string  `json:"benchmark_name"`
	Commit           string  `json:"commit"`
	TotalTime        float64 `json:"total_time"`
	MedianElapsed    float64 `json:"median_elapsed"`
	MinEpochTime     float64 `json:"min_epoch_time"`
	MaxEpochTime     float64 `json:"max_epoch_time"`
	MedianInstructions float64 `json:"median_instructions"`
	MedianCpucycles  float64 `json:"median_cpucycles"`
	Epochs           float64 `json:"epochs"`
	EpochIterations  float64 `json:"epoch_iterations"`
}

type benchLogsClient struct {
	svc      *cloudwatchlogs.CloudWatchLogs
	logGroup string
	pending  []*cloudwatchlogs.InputLogEvent
}

func newBenchLogsClientFromEnv() (*benchLogsClient, error) {
	logGroup := strings.TrimSpace(os.Getenv("BENCHMARK_RESULTS_LOG_GROUP"))
	if logGroup == "" {
		return nil, fmt.Errorf("BENCHMARK_RESULTS_LOG_GROUP is not set")
	}

	region := strings.TrimSpace(os.Getenv("TELEMETRY_CLOUDWATCH_REGION"))
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_REGION"))
	}
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_DEFAULT_REGION"))
	}

	awsCfg := aws.Config{}
	if region != "" {
		awsCfg.Region = aws.String(region)
	}

	sess, err := session.NewSession(&awsCfg)
	if err != nil {
		return nil, fmt.Errorf("creating AWS session: %w", err)
	}

	return &benchLogsClient{
		svc:      cloudwatchlogs.New(sess),
		logGroup: logGroup,
	}, nil
}

func (c *benchLogsClient) createLogStream(streamName string) error {
	_, err := c.svc.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(c.logGroup),
		LogStreamName: aws.String(streamName),
	})
	if err != nil {
		// ResourceAlreadyExistsException is fine (job retry / idempotency)
		if strings.Contains(err.Error(), "ResourceAlreadyExistsException") {
			return nil
		}
		return fmt.Errorf("creating log stream %q: %w", streamName, err)
	}
	return nil
}

func (c *benchLogsClient) queueResult(result *db.BenchmarkResult, commit string) error {
	rec := benchmarkLogRecord{
		BenchmarkName:      result.Name,
		Commit:             commit,
		TotalTime:          result.TotalTime,
		MedianElapsed:      result.MedianElapsed,
		MinEpochTime:       result.MinEpochTime,
		MaxEpochTime:       result.MaxEpochTime,
		MedianInstructions: result.MedianInstructions,
		MedianCpucycles:    result.MedianCpucycles,
		Epochs:             result.Epochs,
		EpochIterations:    result.EpochIterations,
	}

	msg, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshaling benchmark log record: %w", err)
	}

	c.pending = append(c.pending, &cloudwatchlogs.InputLogEvent{
		Timestamp: aws.Int64(time.Now().UnixMilli()),
		Message:   aws.String(string(msg)),
	})

	if len(c.pending) >= 100 {
		return c.flush(commit)
	}
	return nil
}

func (c *benchLogsClient) flush(streamName string) error {
	if len(c.pending) == 0 {
		return nil
	}

	_, err := c.svc.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(c.logGroup),
		LogStreamName: aws.String(streamName),
		LogEvents:     c.pending,
	})
	if err != nil {
		stdlog.Printf("benchlogs: failed to put log events: %v", err)
		return err
	}

	c.pending = nil
	return nil
}
