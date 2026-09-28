package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type ssmAPI interface {
	GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
	PutParameter(input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error)
}

func newSSMClient(region string) (ssmAPI, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("ssm: create session: %w", err)
	}
	return ssm.New(sess), nil
}

// GetLastRunTime reads the last successful run timestamp from SSM.
// Returns zero time when the parameter is missing or its value is not a valid RFC3339 timestamp
// (e.g. the placeholder "initial" written at first Terraform apply).
func GetLastRunTime(client ssmAPI, paramName string) (time.Time, error) {
	out, err := client.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(paramName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == ssm.ErrCodeParameterNotFound {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("ssm: get parameter %q: %w", paramName, err)
	}
	t, err := time.Parse(time.RFC3339, aws.StringValue(out.Parameter.Value))
	if err != nil {
		// Uninitialised placeholder — treat as first run.
		return time.Time{}, nil
	}
	return t, nil
}

// SetLastRunTime writes t as an RFC3339 string to SSM Parameter Store.
func SetLastRunTime(client ssmAPI, paramName string, t time.Time) error {
	_, err := client.PutParameter(&ssm.PutParameterInput{
		Name:      aws.String(paramName),
		Value:     aws.String(t.UTC().Format(time.RFC3339)),
		Type:      aws.String(ssm.ParameterTypeString),
		Overwrite: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("ssm: put parameter %q: %w", paramName, err)
	}
	return nil
}
