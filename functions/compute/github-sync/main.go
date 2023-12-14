package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/corecheck/corecheck/internal/types"
	"github.com/google/go-github/v57/github"

	"github.com/aws/aws-sdk-go/service/sfn"
	"golang.org/x/oauth2"
)

var (
	cfg          = Config{}
	log          = logger.New()
	stateMachine *sfn.SFN
)

type StateMachineInput struct {
	Params types.JobParams `json:"params"`
}

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

		params := StateMachineInput{
			Params: types.JobParams{
				Commit:   master.GetCommit().GetSHA(),
				IsMaster: "true",
			},
		}
		paramsJson, err := json.Marshal(params)
		if err != nil {
			log.Error(err)
			return err
		}
		_, err = stateMachine.StartExecution(&sfn.StartExecutionInput{
			StateMachineArn: aws.String(cfg.StateMachineARN),
			Input:           aws.String(string(paramsJson)),
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

		params := StateMachineInput{
			Params: types.JobParams{
				Commit:   dbPR.Head,
				IsMaster: "false",
				PRNumber: fmt.Sprint(dbPR.Number),
			},
		}
		paramsJson, err := json.Marshal(params)
		if err != nil {
			log.Error(err)
			return err
		}

		_, err = stateMachine.StartExecution(&sfn.StartExecutionInput{
			StateMachineArn: aws.String(cfg.StateMachineARN),
			Input:           aws.String(string(paramsJson)),
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

	checkpoint := time.Date(2023, 12, 14, 0, 0, 0, 0, time.UTC)
	page := 0

	for lastDBUpdate.Before(checkpoint) {
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
	stateMachine = sfn.New(sess)

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
