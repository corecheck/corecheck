package main

import (
	"strconv"
	"time"

	"github.com/corecheck/corecheck/functions/compute/stats/types"
	"github.com/corecheck/corecheck/internal/telemetry"
)

type NumberOfIssuesConsumer struct {
	Total  float64
	Open   float64
	Closed float64
}

func (c *NumberOfIssuesConsumer) Init() {}

func (c *NumberOfIssuesConsumer) ProcessPull(pull *types.Pull) {}

func (c *NumberOfIssuesConsumer) ProcessIssue(issue *types.Issue) {
	if issue.Issue.State == "open" {
		c.Open++
	}

	if issue.Issue.State == "closed" {
		c.Closed++
	}

	c.Total++
}

func (c *NumberOfIssuesConsumer) SendMetrics(metrics telemetry.Client) {
	metrics.Metric("bitcoin.bitcoin.issues.open", c.Open)
	metrics.Metric("bitcoin.bitcoin.issues.closed", c.Closed)
	metrics.Metric("bitcoin.bitcoin.issues.total", c.Total)
}

type UniqueIssueUsersConsumer struct {
	Users map[string]struct{}
}

func (c *UniqueIssueUsersConsumer) Init() {
	c.Users = make(map[string]struct{})
}

func (c *UniqueIssueUsersConsumer) ProcessPull(pull *types.Pull) {}

func (c *UniqueIssueUsersConsumer) ProcessIssue(issue *types.Issue) {
	c.Users[issue.Issue.User.Login] = struct{}{}
}

func (c *UniqueIssueUsersConsumer) SendMetrics(metrics telemetry.Client) {
	metrics.Metric("bitcoin.bitcoin.issues.unique_users", float64(len(c.Users)))
}

type IssuesByUserConsumer struct {
	Open   map[string]float64
	Closed map[string]float64
}

func (c *IssuesByUserConsumer) Init() {
	c.Open = make(map[string]float64)
	c.Closed = make(map[string]float64)
}

func (c *IssuesByUserConsumer) ProcessPull(pull *types.Pull) {}

func (c *IssuesByUserConsumer) ProcessIssue(issue *types.Issue) {
	if _, ok := c.Open[issue.Issue.User.Login]; !ok {
		c.Open[issue.Issue.User.Login] = 0
	}
	if _, ok := c.Closed[issue.Issue.User.Login]; !ok {
		c.Closed[issue.Issue.User.Login] = 0
	}

	if issue.Issue.State == "open" {
		c.Open[issue.Issue.User.Login]++
	}

	if issue.Issue.State == "closed" {
		c.Closed[issue.Issue.User.Login]++
	}
}

func (c *IssuesByUserConsumer) SendMetrics(metrics telemetry.Client) {
	for user, count := range c.Open {
		metrics.Metric("bitcoin.bitcoin.issues.open.by_user", count, telemetry.NewTag("user", user))
	}

	for user, count := range c.Closed {
		metrics.Metric("bitcoin.bitcoin.issues.closed.by_user", count, telemetry.NewTag("user", user))
	}
}

type IssuesByLabelConsumer struct {
	Open   map[string]float64
	Closed map[string]float64
}

func (c *IssuesByLabelConsumer) Init() {
	c.Open = make(map[string]float64)
	c.Closed = make(map[string]float64)
}

func (c *IssuesByLabelConsumer) ProcessPull(pull *types.Pull) {}

func (c *IssuesByLabelConsumer) ProcessIssue(issue *types.Issue) {
	for _, label := range issue.Issue.Labels {
		if _, ok := c.Open[label.Name]; !ok {
			c.Open[label.Name] = 0
		}
		if _, ok := c.Closed[label.Name]; !ok {
			c.Closed[label.Name] = 0
		}

		if issue.Issue.State == "open" {
			c.Open[label.Name]++
		}

		if issue.Issue.State == "closed" {
			c.Closed[label.Name]++
		}
	}
}

func (c *IssuesByLabelConsumer) SendMetrics(metrics telemetry.Client) {
	for label, count := range c.Open {
		metrics.Metric("bitcoin.bitcoin.issues.open.by_label", count, telemetry.NewTag("label", label))
	}

	for label, count := range c.Closed {
		metrics.Metric("bitcoin.bitcoin.issues.closed.by_label", count, telemetry.NewTag("label", label))
	}
}

type TotalCommentsIssueConsumer struct {
	Comments map[int]int
}

func (c *TotalCommentsIssueConsumer) Init() {
	c.Comments = make(map[int]int)
}

func (c *TotalCommentsIssueConsumer) ProcessPull(pull *types.Pull) {}

func (c *TotalCommentsIssueConsumer) ProcessIssue(issue *types.Issue) {
	for _, event := range issue.Events {
		if event.Event == "commented" {
			c.Comments[issue.Issue.Number]++
		}
	}
}

func (c *TotalCommentsIssueConsumer) SendMetrics(metrics telemetry.Client) {
	for issue, count := range c.Comments {
		metrics.Metric("bitcoin.bitcoin.issues.comments", float64(count), telemetry.NewTag("issue", strconv.Itoa(issue)))
	}

	metrics.Metric("bitcoin.bitcoin.issues.comments.total", float64(len(c.Comments)))
}

// PeriodCommentsIssueConsumer emits comment counts for the rolling 30-day window so
// dashboards can show activity for the selected time period rather than all-time totals.
type PeriodCommentsIssueConsumer struct {
	Comments map[int]float64
	since    time.Time
}

func (c *PeriodCommentsIssueConsumer) Init() {
	c.Comments = make(map[int]float64)
	c.since = time.Now().UTC().Add(-commentsPeriod)
}

func (c *PeriodCommentsIssueConsumer) ProcessPull(pull *types.Pull) {}

func (c *PeriodCommentsIssueConsumer) ProcessIssue(issue *types.Issue) {
	for _, event := range issue.Events {
		if event.CreatedAt.Before(c.since) {
			continue
		}
		if event.Event == "commented" {
			c.Comments[issue.Issue.Number]++
		}
	}
}

func (c *PeriodCommentsIssueConsumer) SendMetrics(metrics telemetry.Client) {
	for issue, count := range c.Comments {
		metrics.Metric("bitcoin.bitcoin.issues.comments_period", count, telemetry.NewTag("issue", strconv.Itoa(issue)))
	}
}
