package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

var (
	cfg         = Config{}
	log         = logger.New()
	eventBridge *eventbridge.EventBridge
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

		_, err := eventBridge.PutEvents(&eventbridge.PutEventsInput{
			Entries: []*eventbridge.PutEventsRequestEntry{
				{
					Detail:     aws.String(`{"commit":"` + master.GetCommit().GetSHA() + `","is_master":"true"}`),
					DetailType: aws.String("start-jobs"),
					Source:     aws.String("github-sync"),
				},
			},
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

		_, err = eventBridge.PutEvents(&eventbridge.PutEventsInput{
			Entries: []*eventbridge.PutEventsRequestEntry{
				{
					Detail:     aws.String(`{"commit":"` + dbPR.Head + `","is_master":"false","pr_num":"` + fmt.Sprint(dbPR.Number) + `"}`),
					DetailType: aws.String("start-jobs"),
					Source:     aws.String("github-sync"),
				},
			},
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
			State:     "open",
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

func HandleRequest(ctx context.Context, event interface{}) (string, error) {
	log.Info("GitHub Activity Sync starting")
	log.Debug("Loading config...")
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	log.Debug("Connecting to database...")
	if err := db.Connect(cfg.DatabaseConfig); err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Github.AccessToken},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	sess := session.Must(session.NewSession())
	eventBridge = eventbridge.New(sess)

	log.Info("Now syncing GitHub activity...")
	if err := syncGitHubActivity(client); err != nil {
		log.Fatalf("Error checking GitHub activity: %s", err)
	}

	log.Info("GitHub Activity Sync finished")

	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}
