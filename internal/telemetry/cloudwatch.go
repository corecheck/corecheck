package telemetry

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type CloudWatchConfig struct {
	Namespace string
	Region    string
}

type cloudWatchAPI interface {
	PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error)
}

type cloudWatchClient struct {
	client    cloudWatchAPI
	namespace string
	now       func() time.Time
}

var newCloudWatchAPI = func(cfg CloudWatchConfig) (cloudWatchAPI, error) {
	awsConfig := aws.Config{}
	if cfg.Region != "" {
		awsConfig.Region = aws.String(cfg.Region)
	}

	sess, err := session.NewSession(&awsConfig)
	if err != nil {
		return nil, err
	}

	return cloudwatch.New(sess), nil
}

func NewCloudWatchClientFromEnv() (Client, error) {
	return NewCloudWatchClient(cloudWatchConfigFromEnv())
}

func NewCloudWatchClient(cfg CloudWatchConfig) (Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	client, err := newCloudWatchAPI(cfg)
	if err != nil {
		return nil, err
	}

	return cloudWatchClient{
		client:    client,
		namespace: cfg.Namespace,
		now:       time.Now,
	}, nil
}

func (c cloudWatchClient) Metric(name string, value float64, tags ...Tag) {
	_, err := c.client.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String(c.namespace),
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: aws.String(name),
				Value:      aws.Float64(value),
				Timestamp:  aws.Time(c.now()),
				Dimensions: buildCloudWatchDimensions(tags),
			},
		},
	})
	if err != nil {
		log.Printf("telemetry: failed to write metric %q to cloudwatch: %v", name, err)
	}
}

func cloudWatchConfigFromEnv() CloudWatchConfig {
	namespace := strings.TrimSpace(os.Getenv(EnvCloudWatchNamespace))
	if namespace == "" {
		namespace = DefaultCloudWatchNamespace
	}

	region := strings.TrimSpace(os.Getenv(EnvCloudWatchRegion))
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_REGION"))
	}
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_DEFAULT_REGION"))
	}

	return CloudWatchConfig{
		Namespace: namespace,
		Region:    region,
	}
}

func (cfg CloudWatchConfig) validate() error {
	if cfg.Namespace == "" {
		return fmt.Errorf("%s must be set when using the cloudwatch telemetry backend", EnvCloudWatchNamespace)
	}
	if cfg.Region == "" {
		return fmt.Errorf("%s or AWS_REGION must be set when using the cloudwatch telemetry backend", EnvCloudWatchRegion)
	}

	return nil
}

func buildCloudWatchDimensions(tags []Tag) []*cloudwatch.Dimension {
	dimensions := make([]*cloudwatch.Dimension, 0, len(tags))
	indexByName := map[string]int{}

	appendDimension := func(name, value string) {
		sanitizedName := sanitizeDimensionName(name)
		if index, ok := indexByName[sanitizedName]; ok {
			dimensions[index] = &cloudwatch.Dimension{
				Name:  aws.String(sanitizedName),
				Value: aws.String(value),
			}
			return
		}

		indexByName[sanitizedName] = len(dimensions)
		dimensions = append(dimensions, &cloudwatch.Dimension{
			Name:  aws.String(sanitizedName),
			Value: aws.String(value),
		})
	}

	for _, tag := range tags {
		appendDimension(tag.Key, tag.Value)
	}

	return dimensions
}

func sanitizeDimensionName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "tag"
	}

	var builder strings.Builder
	lastUnderscore := false

	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}

		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}

	sanitized := strings.Trim(builder.String(), "_")
	if sanitized == "" {
		return "tag"
	}

	return sanitized
}
