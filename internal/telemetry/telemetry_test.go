package telemetry

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/timestreamwrite"
)

type stubTimestreamWriter struct {
	input *timestreamwrite.WriteRecordsInput
	err   error
}

func (s *stubTimestreamWriter) WriteRecords(input *timestreamwrite.WriteRecordsInput) (*timestreamwrite.WriteRecordsOutput, error) {
	s.input = input
	return &timestreamwrite.WriteRecordsOutput{}, s.err
}

func TestNewClientFromEnvDefaultsToTimestream(t *testing.T) {
	t.Setenv(EnvBackend, "")
	t.Setenv(EnvTimestreamDatabase, "corecheck")
	t.Setenv(EnvTimestreamRegion, "eu-west-3")

	originalFactory := newTimestreamWriteAPI
	t.Cleanup(func() {
		newTimestreamWriteAPI = originalFactory
	})

	newTimestreamWriteAPI = func(cfg TimestreamConfig) (timestreamWriteAPI, error) {
		if cfg.Database != "corecheck" {
			t.Fatalf("unexpected database %q", cfg.Database)
		}
		if cfg.Table != DefaultTimestreamTable {
			t.Fatalf("unexpected table %q", cfg.Table)
		}
		if cfg.Region != "eu-west-3" {
			t.Fatalf("unexpected region %q", cfg.Region)
		}
		return &stubTimestreamWriter{}, nil
	}

	client, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("NewClientFromEnv() error = %v", err)
	}

	if _, ok := client.(timestreamClient); !ok {
		t.Fatalf("expected timestreamClient, got %T", client)
	}
}

func TestNewClientFromEnvBuildsTimestreamClient(t *testing.T) {
	t.Setenv(EnvBackend, BackendTimestream)
	t.Setenv(EnvTimestreamDatabase, "corecheck")
	t.Setenv(EnvTimestreamRegion, "eu-west-3")
	t.Setenv(EnvTimestreamTable, "")

	originalFactory := newTimestreamWriteAPI
	t.Cleanup(func() {
		newTimestreamWriteAPI = originalFactory
	})

	newTimestreamWriteAPI = func(cfg TimestreamConfig) (timestreamWriteAPI, error) {
		if cfg.Database != "corecheck" {
			t.Fatalf("unexpected database %q", cfg.Database)
		}
		if cfg.Table != DefaultTimestreamTable {
			t.Fatalf("unexpected table %q", cfg.Table)
		}
		if cfg.Region != "eu-west-3" {
			t.Fatalf("unexpected region %q", cfg.Region)
		}
		return &stubTimestreamWriter{}, nil
	}

	client, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("NewClientFromEnv() error = %v", err)
	}

	if _, ok := client.(timestreamClient); !ok {
		t.Fatalf("expected timestreamClient, got %T", client)
	}
}

func TestNewClientFromEnvRejectsInvalidBackend(t *testing.T) {
	t.Setenv(EnvBackend, "cloudwatch")

	if _, err := NewClientFromEnv(); err == nil {
		t.Fatal("expected invalid backend error")
	}
}

func TestTimestreamClientMetricWritesExpectedRecord(t *testing.T) {
	writer := &stubTimestreamWriter{}
	client := timestreamClient{
		writer:   writer,
		database: "corecheck",
		table:    "dashboard_metrics",
		now: func() time.Time {
			return time.UnixMilli(1700000000000)
		},
	}

	client.Metric("bitcoin.bitcoin.issues.open.by_label", 42, NewTag("Pull Number", "7"), NewTag("pull-number", "9"))

	if writer.input == nil {
		t.Fatal("expected WriteRecords to be called")
	}
	if got := *writer.input.DatabaseName; got != "corecheck" {
		t.Fatalf("unexpected database %q", got)
	}
	if got := *writer.input.TableName; got != "dashboard_metrics" {
		t.Fatalf("unexpected table %q", got)
	}
	if len(writer.input.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(writer.input.Records))
	}

	record := writer.input.Records[0]
	if got := *record.MeasureName; got != "value" {
		t.Fatalf("unexpected measure name %q", got)
	}
	if got := *record.MeasureValue; got != "42" {
		t.Fatalf("unexpected measure value %q", got)
	}
	if got := *record.Time; got != "1700000000000" {
		t.Fatalf("unexpected time %q", got)
	}
	if len(record.Dimensions) != 2 {
		t.Fatalf("expected 2 dimensions, got %d", len(record.Dimensions))
	}
	if got := *record.Dimensions[0].Name; got != "metric_name" {
		t.Fatalf("unexpected first dimension name %q", got)
	}
	if got := *record.Dimensions[0].Value; got != "bitcoin.bitcoin.issues.open.by_label" {
		t.Fatalf("unexpected first dimension value %q", got)
	}
	if got := *record.Dimensions[1].Name; got != "pull_number" {
		t.Fatalf("unexpected tag dimension name %q", got)
	}
	if got := *record.Dimensions[1].Value; got != "9" {
		t.Fatalf("unexpected tag dimension value %q", got)
	}
}
