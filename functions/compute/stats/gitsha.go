package main

import (
	"fmt"

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

// GetGitSHA reads the stored HEAD commit SHA from SSM Parameter Store.
// Returns "" when the parameter does not exist yet (first run).
func GetGitSHA(client ssmAPI, paramName string) (string, error) {
	out, err := client.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(paramName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == ssm.ErrCodeParameterNotFound {
			return "", nil
		}
		return "", fmt.Errorf("ssm: get parameter %q: %w", paramName, err)
	}
	return aws.StringValue(out.Parameter.Value), nil
}

// SetGitSHA writes sha to SSM Parameter Store, overwriting any existing value.
func SetGitSHA(client ssmAPI, paramName, sha string) error {
	_, err := client.PutParameter(&ssm.PutParameterInput{
		Name:      aws.String(paramName),
		Value:     aws.String(sha),
		Type:      aws.String(ssm.ParameterTypeString),
		Overwrite: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("ssm: put parameter %q: %w", paramName, err)
	}
	return nil
}
