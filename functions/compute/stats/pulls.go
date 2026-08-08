package main

import (
	"time"

	"github.com/corecheck/corecheck/functions/compute/stats/types"
	"github.com/corecheck/corecheck/internal/telemetry"
)

type NumberOfPullsConsumer struct {
	Open float64
}

func (c *NumberOfPullsConsumer) Init()                            {}
func (c *NumberOfPullsConsumer) ProcessIssue(issue *types.Issue) {}

func (c *NumberOfPullsConsumer) ProcessPull(pull *types.Pull) {
	if pull.Pull.State == "open" {
		c.Open++
	}
}

func (c *NumberOfPullsConsumer) SendMetrics(metrics telemetry.Client) {
	metrics.Metric("bitcoin.bitcoin.pulls.open", c.Open)
}

// parseEventTime converts the pull timeline event created_at field (typed as any because
// some event kinds omit it) into a time.Time. Returns false when the value is absent or
// unparseable.
func parseEventTime(v any) (time.Time, bool) {
	if v == nil {
		return time.Time{}, false
	}
	s, ok := v.(string)
	if !ok {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	return t, err == nil
}

// PeriodCommentsAndReviewsByPullConsumer removed — superseded by the CloudWatch Logs event stream.
