package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func cloudWatchRegion() string {
	for _, key := range []string{"TELEMETRY_CLOUDWATCH_REGION", "AWS_REGION", "AWS_DEFAULT_REGION"} {
		if region := strings.TrimSpace(os.Getenv(key)); region != "" {
			return region
		}
	}
	return ""
}

// putOpenCountMetrics records the same open-count gauges the GitHub dashboard
// will chart. Datadog still receives these values through the existing consumers.
func putOpenCountMetrics(openPulls, openIssues float64) error {
	namespace := strings.TrimSpace(os.Getenv("TELEMETRY_CLOUDWATCH_NAMESPACE"))
	region := cloudWatchRegion()
	if namespace == "" || region == "" {
		fmt.Println("stats: TELEMETRY_CLOUDWATCH_NAMESPACE or region unset; skipping open-count metrics")
		return nil
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return fmt.Errorf("cloudwatch: create session: %w", err)
	}

	_, err = cloudwatch.New(sess).PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: aws.String("bitcoin.bitcoin.pulls.open"),
				Value:      aws.Float64(openPulls),
				Unit:       aws.String(cloudwatch.StandardUnitCount),
			},
			{
				MetricName: aws.String("bitcoin.bitcoin.issues.open"),
				Value:      aws.Float64(openIssues),
				Unit:       aws.String(cloudwatch.StandardUnitCount),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("cloudwatch: put open counts: %w", err)
	}
	return nil
}
