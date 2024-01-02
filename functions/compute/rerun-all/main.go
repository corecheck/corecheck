package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/corecheck/corecheck/internal/types"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sfn"
)

var (
	cfg          = Config{}
	log          = logger.New()
	stateMachine *sfn.SFN
)

type StateMachineInput struct {
	Params types.JobParams `json:"params"`
}

func HandleRequest(ctx context.Context, event interface{}) (string, error) {
	log.Debug("Loading config...")
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	log.Debug("Connecting to database...")
	if err := db.Connect(cfg.DatabaseConfig); err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	sess := session.Must(session.NewSession())
	stateMachine = sfn.New(sess)
	s3sess := s3.New(sess)

	pulls, err := db.ListAllPulls()
	if err != nil {
		log.Error(err)
		return "", err
	}

	for _, pull := range pulls {
		report, err := db.GetLatestPullCoverageReport(pull.Number)
		if err != nil {
			log.Error(err)
			return "", err
		}

		res, err := s3sess.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket: aws.String(os.Getenv("BUCKET_DATA_URL")),
			Prefix: aws.String(fmt.Sprintf("%d/%s/coverage.json", pull.Number, report.Commit)),
		})

		if err != nil {
			log.Error(err)
			return "", err
		}

		if len(res.Contents) == 0 {
			continue
		}

		log.Infof("Rerunning coverage job for PR %d", pull.Number)
		params := StateMachineInput{
			Params: types.JobParams{
				Commit:     report.Commit,
				IsMaster:   "false",
				PRNumber:   fmt.Sprint(pull.Number),
				BaseCommit: report.BaseCommit,
			},
		}
		paramsJson, err := json.Marshal(params)
		if err != nil {
			log.Error(err)
			return "", err
		}

		_, err = stateMachine.StartExecution(&sfn.StartExecutionInput{
			StateMachineArn: aws.String(cfg.StateMachineARN),
			Input:           aws.String(string(paramsJson)),
		})

		if err != nil {
			log.Error(err)
			return "", err
		}
	}

	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}
