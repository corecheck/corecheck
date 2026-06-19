package telemetry

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type stubCloudWatchClient struct {
	input *cloudwatch.PutMetricDataInput
	err   error
}

func (s *stubCloudWatchClient) PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	s.input = input
	return &cloudwatch.PutMetricDataOutput{}, s.err
}

func TestNewClientFromEnvDefaultsToCloudWatch(t *testing.T) {
	t.Setenv(EnvBackend, "")
	t.Setenv(EnvCloudWatchNamespace, "Corecheck/dev")
	t.Setenv(EnvCloudWatchRegion, "eu-west-3")

	originalFactory := newCloudWatchAPI
	t.Cleanup(func() {
		newCloudWatchAPI = originalFactory
	})

	newCloudWatchAPI = func(cfg CloudWatchConfig) (cloudWatchAPI, error) {
		if cfg.Namespace != "Corecheck/dev" {
			t.Fatalf("unexpected namespace %q", cfg.Namespace)
		}
		if cfg.Region != "eu-west-3" {
			t.Fatalf("unexpected region %q", cfg.Region)
		}
		return &stubCloudWatchClient{}, nil
	}

	client, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("NewClientFromEnv() error = %v", err)
	}

	if _, ok := client.(cloudWatchClient); !ok {
		t.Fatalf("expected cloudWatchClient, got %T", client)
	}
}

func TestNewClientFromEnvBuildsCloudWatchClient(t *testing.T) {
	t.Setenv(EnvBackend, BackendCloudWatch)
	t.Setenv(EnvCloudWatchNamespace, "Corecheck/prod")
	t.Setenv(EnvCloudWatchRegion, "eu-west-3")

	originalFactory := newCloudWatchAPI
	t.Cleanup(func() {
		newCloudWatchAPI = originalFactory
	})

	newCloudWatchAPI = func(cfg CloudWatchConfig) (cloudWatchAPI, error) {
		if cfg.Namespace != "Corecheck/prod" {
			t.Fatalf("unexpected namespace %q", cfg.Namespace)
		}
		if cfg.Region != "eu-west-3" {
			t.Fatalf("unexpected region %q", cfg.Region)
		}
		return &stubCloudWatchClient{}, nil
	}

	client, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("NewClientFromEnv() error = %v", err)
	}

	if _, ok := client.(cloudWatchClient); !ok {
		t.Fatalf("expected cloudWatchClient, got %T", client)
	}
}

func TestNewClientFromEnvRejectsInvalidBackend(t *testing.T) {
	t.Setenv(EnvBackend, "invalid")

	if _, err := NewClientFromEnv(); err == nil {
		t.Fatal("expected invalid backend error")
	}
}

func TestCloudWatchClientMetricWritesExpectedRecord(t *testing.T) {
	writer := &stubCloudWatchClient{}
	client := cloudWatchClient{
		client:    writer,
		namespace: "Corecheck/dev",
		now: func() time.Time {
			return time.UnixMilli(1700000000000)
		},
	}

	client.Metric("bitcoin.bitcoin.issues.open.by_label", 42, NewTag("Pull Number", "7"), NewTag("pull-number", "9"))

	if writer.input == nil {
		t.Fatal("expected PutMetricData to be called")
	}
	if got := *writer.input.Namespace; got != "Corecheck/dev" {
		t.Fatalf("unexpected namespace %q", got)
	}
	if len(writer.input.MetricData) != 1 {
		t.Fatalf("expected 1 metric datum, got %d", len(writer.input.MetricData))
	}

	datum := writer.input.MetricData[0]
	if got := *datum.MetricName; got != "bitcoin.bitcoin.issues.open.by_label" {
		t.Fatalf("unexpected metric name %q", got)
	}
	if got := *datum.Value; got != 42 {
		t.Fatalf("unexpected value %v", got)
	}
	if got := datum.Timestamp.UnixMilli(); got != 1700000000000 {
		t.Fatalf("unexpected timestamp %d", got)
	}
	if len(datum.Dimensions) != 1 {
		t.Fatalf("expected 1 dimension, got %d", len(datum.Dimensions))
	}
	if got := *datum.Dimensions[0].Name; got != "pull_number" {
		t.Fatalf("unexpected dimension name %q", got)
	}
	if got := *datum.Dimensions[0].Value; got != "9" {
		t.Fatalf("unexpected dimension value %q", got)
	}
}
