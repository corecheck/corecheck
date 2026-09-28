package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type cwLogsAPI interface {
	CreateLogStream(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error)
	PutLogEvents(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error)
}

// CWLogsWriter writes log events to a CloudWatch Logs log group.
// Each run creates a fresh log stream so no sequence token management is needed.
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

	streamName := fmt.Sprintf("github-events/%s",
		time.Now().UTC().Format("2006-01-02/15-04-05"),
	)

	client := cloudwatchlogs.New(sess)

	_, err = client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(streamName),
	})
	if err != nil {
		return nil, fmt.Errorf("cwlogs: create log stream %q: %w", streamName, err)
	}
	log.Printf("cwlogs: writing to log stream %s/%s", logGroupName, streamName)

	return &CWLogsWriter{
		client:       client,
		logGroupName: logGroupName,
		streamName:   streamName,
	}, nil
}

// Write sends a single event to CloudWatch Logs.
func (w *CWLogsWriter) Write(event *cloudwatchlogs.InputLogEvent) error {
	_, err := w.client.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(w.logGroupName),
		LogStreamName: aws.String(w.streamName),
		LogEvents:     []*cloudwatchlogs.InputLogEvent{event},
	})
	if err != nil {
		return fmt.Errorf("cwlogs: put log event: %w", err)
	}
	return nil
}
