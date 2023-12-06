package main

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

var (
	cfg = Config{}
	log = logger.New()
	svc *sqs.SQS
)

func checkMasterCoverage(c *github.Client) error {
	log.Info("Checking master coverage...")
	master, _, err := c.Repositories.GetBranch(context.Background(), "bitcoin", "bitcoin", "master", 0)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug("Latest commit: %s", master.GetCommit().GetSHA())

	hasReport, err := db.HasCoverageReportForCommit(master.GetCommit().GetSHA())
	if err != nil {
		log.Error(err)
		return err
	}

	if hasReport {
		log.Info("Master has coverage for latest commit")
	} else {
		log.Info("Master does not have coverage for latest commit, adding to queue")

		_, err := svc.SendMessage(&sqs.SendMessageInput{
			DelaySeconds: aws.Int64(10),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"commit": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(master.GetCommit().GetSHA()),
				},
				"is_master": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("true"),
				},
			},
			MessageBody: aws.String("Master coverage for latest commit (" + master.GetCommit().GetSHA() + ")"),
			QueueUrl:    aws.String(cfg.SQS.QueueURL),
		})

		if err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

func handlePullRequest(pr *github.PullRequest) error {
	log.Debugf("Processing PR %d", pr.GetNumber())
	err := db.UpdateOrCreatePR(pr)
	if err != nil {
		log.Error(err)
		return err
	}

	dbPR, err := db.GetPR(pr.GetNumber())
	if err != nil {
		log.Error(err)
		return err
	}

	if pr.GetState() == "open" {
		hasReport, err := db.HasCoverageReportForCommit(dbPR.Head)
		if err != nil {
			log.Error(err)
			return err
		}

		if hasReport {
			log.Infof("PR %d has coverage for latest commit", dbPR.Number)
			return nil
		}

		log.Info("PR does not have coverage for latest commit, triggering coverage job")

		_, err = svc.SendMessage(&sqs.SendMessageInput{
			DelaySeconds: aws.Int64(10),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"commit": {
					DataType:    aws.String("String"),
					StringValue: aws.String(dbPR.Head),
				},
				"is_master": {
					DataType:    aws.String("String"),
					StringValue: aws.String("false"),
				},
				"pr_num": {
					DataType:    aws.String("Number"),
					StringValue: aws.String(strconv.Itoa(dbPR.Number)),
				},
			},
			MessageBody: aws.String("PR #" + strconv.Itoa(dbPR.Number) + " coverage for latest commit (" + dbPR.Head + ")"),
			QueueUrl:    aws.String(cfg.SQS.QueueURL),
		})

		if err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

func checkPullsCoverage(c *github.Client) error {
	log.Info("Syncing PRs from GitHub...")
	lastDBUpdate, err := db.GetLatestPRUpdate()
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Last DB update: %s", lastDBUpdate.Format(time.RFC3339))

	now := time.Now()

	page := 0

	for lastDBUpdate.Before(now) {
		log.Debugf("Retrieving page %d", page)
		prs, _, err := c.PullRequests.List(context.Background(), "bitcoin", "bitcoin", &github.PullRequestListOptions{
			State:     "all",
			Sort:      "updated",
			Direction: "desc",
			ListOptions: github.ListOptions{
				PerPage: 100,
				Page:    page,
			},
		})

		if err != nil {
			log.Errorf("Error retrieving PRs: %s", err)
			return err
		}

		if len(prs) == 0 {
			break
		}

		for _, pr := range prs {
			if pr.GetUpdatedAt().Before(lastDBUpdate) {
				break
			}

			if err := handlePullRequest(pr); err != nil {
				log.Errorf("Error handling PR: %s", err)
				return err
			}
		}

		page++
	}

	return nil
}

func syncGitHubActivity(c *github.Client) error {
	if err := checkMasterCoverage(c); err != nil {
		log.Errorf("Error checking master coverage: %s", err)
		return err
	}

	if err := checkPullsCoverage(c); err != nil {
		log.Errorf("Error checking PRs coverage: %s", err)
		return err
	}

	return nil
}

func main() {
	if err := config.Load(&cfg); err != nil {
		log.Fatal(err)
	}
	if err := db.Connect(cfg.DatabaseConfig); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Github.AccessToken},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	sess := session.Must(session.NewSession())
	svc = sqs.New(sess)

	log.Info("GitHub Activity Sync started")
	if err := syncGitHubActivity(client); err != nil {
		log.Fatalf("Error checking GitHub activity: %s", err)
	}

	log.Info("GitHub Activity Sync finished")
}
