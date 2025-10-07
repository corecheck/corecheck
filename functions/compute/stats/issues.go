package main

import (
	"bitcoin-stats-datadog/types"
	"strconv"

	ddlambda "github.com/DataDog/datadog-lambda-go"
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

func (c *NumberOfIssuesConsumer) SendMetrics() {
	ddlambda.Metric("bitcoin.bitcoin.issues.open", c.Open)
	ddlambda.Metric("bitcoin.bitcoin.issues.closed", c.Closed)
	ddlambda.Metric("bitcoin.bitcoin.issues.total", c.Total)
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

func (c *UniqueIssueUsersConsumer) SendMetrics() {
	ddlambda.Metric("bitcoin.bitcoin.issues.unique_users", float64(len(c.Users)))
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

func (c *IssuesByUserConsumer) SendMetrics() {
	for user, count := range c.Open {
		ddlambda.Metric("bitcoin.bitcoin.issues.open.by_user", count, "user:"+user)
	}

	for user, count := range c.Closed {
		ddlambda.Metric("bitcoin.bitcoin.issues.closed.by_user", count, "user:"+user)
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

func (c *IssuesByLabelConsumer) SendMetrics() {
	for label, count := range c.Open {
		ddlambda.Metric("bitcoin.bitcoin.issues.open.by_label", count, "label:"+label)
	}

	for label, count := range c.Closed {
		ddlambda.Metric("bitcoin.bitcoin.issues.closed.by_label", count, "label:"+label)
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

func (c *TotalCommentsIssueConsumer) SendMetrics() {
	for issue, count := range c.Comments {
		ddlambda.Metric("bitcoin.bitcoin.issues.comments", float64(count), "issue:"+strconv.Itoa(issue))
	}

	ddlambda.Metric("bitcoin.bitcoin.issues.comments.total", float64(len(c.Comments)))
}
