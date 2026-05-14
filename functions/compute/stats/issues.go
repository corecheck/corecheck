package main

import (
	"github.com/corecheck/corecheck/functions/compute/stats/types"
	"github.com/corecheck/corecheck/internal/telemetry"
)

type NumberOfIssuesConsumer struct {
	Open float64
}

func (c *NumberOfIssuesConsumer) Init() {}

func (c *NumberOfIssuesConsumer) ProcessPull(pull *types.Pull) {}

func (c *NumberOfIssuesConsumer) ProcessIssue(issue *types.Issue) {
	if issue.Issue.State == "open" {
		c.Open++
	}
}

func (c *NumberOfIssuesConsumer) SendMetrics(metrics telemetry.Client) {
	metrics.Metric("bitcoin.bitcoin.issues.open", c.Open)
}
