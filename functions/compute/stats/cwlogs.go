package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

const cwLogsBatchSize = 500

type cwLogsAPI interface {
	CreateLogStream(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error)
	PutLogEvents(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error)
}

// CWLogsWriter writes log events to a CloudWatch Logs log group.
// Each Write call creates a fresh log stream (named github-events/<date>/<seq>)
// so no sequence token management is needed.
type CWLogsWriter struct {
	client       cwLogsAPI
	logGroupName string
	streamName   string
}

func NewCWLogsWriter(region, logGroupName string) (*CWLogsWriter, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("cwlogs: create session: %w", err)
	}

	streamName := fmt.Sprintf("github-events/%s/%d",
		time.Now().UTC().Format("2006-01-02"),
		time.Now().UnixNano(),
	)

	client := cloudwatchlogs.New(sess)

	_, err = client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(streamName),
	})
	if err != nil {
		return nil, fmt.Errorf("cwlogs: create log stream %q: %w", streamName, err)
	}

	return &CWLogsWriter{
		client:       client,
		logGroupName: logGroupName,
		streamName:   streamName,
	}, nil
}

// Write sorts events chronologically (required by CloudWatch Logs) then sends
// them in batches that respect both the 500-event limit and the 24-hour span limit.
func (w *CWLogsWriter) Write(events []*cloudwatchlogs.InputLogEvent) error {
	sort.Slice(events, func(i, j int) bool {
		return aws.Int64Value(events[i].Timestamp) < aws.Int64Value(events[j].Timestamp)
	})

	const maxSpanMs = 24 * 60 * 60 * 1000 // 24 hours in milliseconds

	start := 0
	for start < len(events) {
		batchStart := aws.Int64Value(events[start].Timestamp)
		end := start + 1
		for end < len(events) &&
			end-start < cwLogsBatchSize &&
			aws.Int64Value(events[end].Timestamp)-batchStart < maxSpanMs {
			end++
		}
		batch := events[start:end]

		_, err := w.client.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
			LogGroupName:  aws.String(w.logGroupName),
			LogStreamName: aws.String(w.streamName),
			LogEvents:     batch,
		})
		if err != nil {
			return fmt.Errorf("cwlogs: put log events (batch starting %d): %w", start, err)
		}
		start = end
	}
	return nil
}
