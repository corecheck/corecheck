package main

import (
	"strconv"
	"time"

	"github.com/corecheck/corecheck/functions/compute/stats/types"
	"github.com/corecheck/corecheck/internal/telemetry"
)

// commentsPeriod is the rolling window used for period-scoped comment/review metrics.
const commentsPeriod = 30 * 24 * time.Hour

type NumberOfPullsConsumer struct {
	Total  float64
	Open   float64
	Closed float64
	Merged float64
}

func (c *NumberOfPullsConsumer) Init()                           {}
func (c *NumberOfPullsConsumer) ProcessIssue(issue *types.Issue) {}

func (c *NumberOfPullsConsumer) ProcessPull(pull *types.Pull) {
	if pull.Pull.State == "open" {
		c.Open++
	}

	if pull.Pull.State == "closed" {
		c.Closed++
	}

	if !pull.Pull.MergedAt.IsZero() {
		c.Merged++
	}

	c.Total++
}

func (c *NumberOfPullsConsumer) SendMetrics(metrics telemetry.Client) {
	metrics.Metric("bitcoin.bitcoin.pulls.open", c.Open)
	metrics.Metric("bitcoin.bitcoin.pulls.closed", c.Closed)
	metrics.Metric("bitcoin.bitcoin.pulls.merged", c.Merged)
	metrics.Metric("bitcoin.bitcoin.pulls.total", c.Total)
}

type UniqueAuthorsConsumer struct {
	Users map[string]struct{}
}

func (c *UniqueAuthorsConsumer) Init() {
	c.Users = make(map[string]struct{})
}

func (c *UniqueAuthorsConsumer) ProcessIssue(issue *types.Issue) {}

func (c *UniqueAuthorsConsumer) ProcessPull(pull *types.Pull) {
	if pull.Pull.MergedAt.IsZero() {
		return
	}

	c.Users[pull.Pull.User.Login] = struct{}{}
}

func (c *UniqueAuthorsConsumer) SendMetrics(metrics telemetry.Client) {
	metrics.Metric("bitcoin.bitcoin.pulls.unique_authors", float64(len(c.Users)))
}

type PullsByUserConsumer struct {
	Open   map[string]float64
	Closed map[string]float64
	Merged map[string]float64
}

func (c *PullsByUserConsumer) Init() {
	c.Open = make(map[string]float64)
	c.Closed = make(map[string]float64)
	c.Merged = make(map[string]float64)
}

func (c *PullsByUserConsumer) ProcessIssue(issue *types.Issue) {}

func (c *PullsByUserConsumer) ProcessPull(pull *types.Pull) {
	if _, ok := c.Open[pull.Pull.User.Login]; !ok {
		c.Open[pull.Pull.User.Login] = 0
	}
	if _, ok := c.Closed[pull.Pull.User.Login]; !ok {
		c.Closed[pull.Pull.User.Login] = 0
	}
	if _, ok := c.Merged[pull.Pull.User.Login]; !ok {
		c.Merged[pull.Pull.User.Login] = 0
	}

	if pull.Pull.State == "open" {
		c.Open[pull.Pull.User.Login]++
	}

	if pull.Pull.State == "closed" {
		c.Closed[pull.Pull.User.Login]++
	}

	if !pull.Pull.MergedAt.IsZero() {
		c.Merged[pull.Pull.User.Login]++
	}
}

func (c *PullsByUserConsumer) SendMetrics(metrics telemetry.Client) {
	for user, count := range c.Open {
		metrics.Metric("bitcoin.bitcoin.pulls.open.by_user", count, telemetry.NewTag("user", user))
	}
	for user, count := range c.Closed {
		metrics.Metric("bitcoin.bitcoin.pulls.closed.by_user", count, telemetry.NewTag("user", user))
	}
	for user, count := range c.Merged {
		metrics.Metric("bitcoin.bitcoin.pulls.merged.by_user", count, telemetry.NewTag("user", user))
	}
}

type PullsByLabelConsumer struct {
	Open   map[string]float64
	Closed map[string]float64
	Merged map[string]float64
}

func (c *PullsByLabelConsumer) Init() {
	c.Open = make(map[string]float64)
	c.Closed = make(map[string]float64)
	c.Merged = make(map[string]float64)
}

func (c *PullsByLabelConsumer) ProcessIssue(issue *types.Issue) {}

func (c *PullsByLabelConsumer) ProcessPull(pull *types.Pull) {
	for _, label := range pull.Pull.Labels {
		if _, ok := c.Open[label.Name]; !ok {
			c.Open[label.Name] = 0
		}
		if _, ok := c.Closed[label.Name]; !ok {
			c.Closed[label.Name] = 0
		}
		if _, ok := c.Merged[label.Name]; !ok {
			c.Merged[label.Name] = 0
		}

		if pull.Pull.State == "open" {
			c.Open[label.Name]++
		}

		if pull.Pull.State == "closed" {
			c.Closed[label.Name]++
		}

		if !pull.Pull.MergedAt.IsZero() {
			c.Merged[label.Name]++
		}
	}
}

func (c *PullsByLabelConsumer) SendMetrics(metrics telemetry.Client) {
	for label, count := range c.Open {
		metrics.Metric("bitcoin.bitcoin.pulls.open.by_label", count, telemetry.NewTag("label", label))
	}
	for label, count := range c.Closed {
		metrics.Metric("bitcoin.bitcoin.pulls.closed.by_label", count, telemetry.NewTag("label", label))
	}
	for label, count := range c.Merged {
		metrics.Metric("bitcoin.bitcoin.pulls.merged.by_label", count, telemetry.NewTag("label", label))
	}
}

type TotalCommentsAndReviewsByPullConsumer struct {
	Comments map[int]float64
	Reviews  map[int]float64
}

func (c *TotalCommentsAndReviewsByPullConsumer) Init() {
	c.Comments = make(map[int]float64)
	c.Reviews = make(map[int]float64)
}

func (c *TotalCommentsAndReviewsByPullConsumer) ProcessIssue(issue *types.Issue) {}

func (c *TotalCommentsAndReviewsByPullConsumer) ProcessPull(pull *types.Pull) {
	for _, event := range pull.Events {
		if event.Event == "commented" {
			c.Comments[pull.Pull.Number]++
		} else if event.Event == "reviewed" {
			c.Reviews[pull.Pull.Number]++
		}
	}
}

func (c *TotalCommentsAndReviewsByPullConsumer) SendMetrics(metrics telemetry.Client) {
	for pull, count := range c.Comments {
		metrics.Metric("bitcoin.bitcoin.pulls.comments", count, telemetry.NewTag("pull", strconv.Itoa(pull)))
	}
	for pull, count := range c.Reviews {
		metrics.Metric("bitcoin.bitcoin.pulls.reviews", count, telemetry.NewTag("pull", strconv.Itoa(pull)))
	}

	metrics.Metric("bitcoin.bitcoin.pulls.comments.total", float64(len(c.Comments)))
	metrics.Metric("bitcoin.bitcoin.pulls.reviews.total", float64(len(c.Reviews)))
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

// PeriodCommentsAndReviewsByPullConsumer emits comment and review counts for the rolling
// 30-day window so dashboards can show activity for the selected time period rather than
// all-time totals.
type PeriodCommentsAndReviewsByPullConsumer struct {
	Comments map[int]float64
	Reviews  map[int]float64
	since    time.Time
}

func (c *PeriodCommentsAndReviewsByPullConsumer) Init() {
	c.Comments = make(map[int]float64)
	c.Reviews = make(map[int]float64)
	c.since = time.Now().UTC().Add(-commentsPeriod)
}

func (c *PeriodCommentsAndReviewsByPullConsumer) ProcessIssue(issue *types.Issue) {}

func (c *PeriodCommentsAndReviewsByPullConsumer) ProcessPull(pull *types.Pull) {
	for _, event := range pull.Events {
		t, ok := parseEventTime(event.CreatedAt)
		if !ok || t.Before(c.since) {
			continue
		}
		if event.Event == "commented" {
			c.Comments[pull.Pull.Number]++
		} else if event.Event == "reviewed" {
			c.Reviews[pull.Pull.Number]++
		}
	}
}

func (c *PeriodCommentsAndReviewsByPullConsumer) SendMetrics(metrics telemetry.Client) {
	for pull, count := range c.Comments {
		metrics.Metric("bitcoin.bitcoin.pulls.comments_period", count, telemetry.NewTag("pull", strconv.Itoa(pull)))
	}
	for pull, count := range c.Reviews {
		metrics.Metric("bitcoin.bitcoin.pulls.reviews_period", count, telemetry.NewTag("pull", strconv.Itoa(pull)))
	}
}
