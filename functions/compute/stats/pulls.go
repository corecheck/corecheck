package main

import (
	"bitcoin-stats-datadog/types"
	"strconv"

	ddlambda "github.com/DataDog/datadog-lambda-go"
)

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

func (c *NumberOfPullsConsumer) SendMetrics() {
	ddlambda.Metric("bitcoin.bitcoin.pulls.open", c.Open)
	ddlambda.Metric("bitcoin.bitcoin.pulls.closed", c.Closed)
	ddlambda.Metric("bitcoin.bitcoin.pulls.merged", c.Merged)
	ddlambda.Metric("bitcoin.bitcoin.pulls.total", c.Total)
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

func (c *UniqueAuthorsConsumer) SendMetrics() {
	ddlambda.Metric("bitcoin.bitcoin.pulls.unique_authors", float64(len(c.Users)))
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

func (c *PullsByUserConsumer) SendMetrics() {
	for user, count := range c.Open {
		ddlambda.Metric("bitcoin.bitcoin.pulls.open.by_user", count, "user:"+user)
	}
	for user, count := range c.Closed {
		ddlambda.Metric("bitcoin.bitcoin.pulls.closed.by_user", count, "user:"+user)
	}
	for user, count := range c.Merged {
		ddlambda.Metric("bitcoin.bitcoin.pulls.merged.by_user", count, "user:"+user)
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

func (c *PullsByLabelConsumer) SendMetrics() {
	for label, count := range c.Open {
		ddlambda.Metric("bitcoin.bitcoin.pulls.open.by_label", count, "label:"+label)
	}
	for label, count := range c.Closed {
		ddlambda.Metric("bitcoin.bitcoin.pulls.closed.by_label", count, "label:"+label)
	}
	for label, count := range c.Merged {
		ddlambda.Metric("bitcoin.bitcoin.pulls.merged.by_label", count, "label:"+label)
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

func (c *TotalCommentsAndReviewsByPullConsumer) SendMetrics() {
	for pull, count := range c.Comments {
		ddlambda.Metric("bitcoin.bitcoin.pulls.comments", count, "pull:"+strconv.Itoa(pull))
	}
	for pull, count := range c.Reviews {
		ddlambda.Metric("bitcoin.bitcoin.pulls.reviews", count, "pull:"+strconv.Itoa(pull))
	}

	ddlambda.Metric("bitcoin.bitcoin.pulls.comments.total", float64(len(c.Comments)))
	ddlambda.Metric("bitcoin.bitcoin.pulls.reviews.total", float64(len(c.Reviews)))
}
