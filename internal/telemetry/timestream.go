package telemetry

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
)

type TimestreamConfig struct {
	Database string
	Table    string
	Region   string
}

type timestreamWriteAPI interface {
	WriteRecords(input *timestreamwrite.WriteRecordsInput) (*timestreamwrite.WriteRecordsOutput, error)
}

type timestreamClient struct {
	writer   timestreamWriteAPI
	database string
	table    string
	now      func() time.Time
}

var newTimestreamWriteAPI = func(cfg TimestreamConfig) (timestreamWriteAPI, error) {
	awsConfig := aws.Config{}
	if cfg.Region != "" {
		awsConfig.Region = aws.String(cfg.Region)
	}

	sess, err := session.NewSession(&awsConfig)
	if err != nil {
		return nil, err
	}

	return timestreamwrite.New(sess), nil
}

func NewTimestreamClientFromEnv() (Client, error) {
	return NewTimestreamClient(timestreamConfigFromEnv())
}

func NewTimestreamClient(cfg TimestreamConfig) (Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	writer, err := newTimestreamWriteAPI(cfg)
	if err != nil {
		return nil, err
	}

	return timestreamClient{
		writer:   writer,
		database: cfg.Database,
		table:    cfg.Table,
		now:      time.Now,
	}, nil
}

func (c timestreamClient) Metric(name string, value float64, tags ...Tag) {
	_, err := c.writer.WriteRecords(&timestreamwrite.WriteRecordsInput{
		DatabaseName: aws.String(c.database),
		TableName:    aws.String(c.table),
		Records: []*timestreamwrite.Record{
			{
				Dimensions:       buildTimestreamDimensions(name, tags),
				MeasureName:      aws.String("value"),
				MeasureValue:     aws.String(strconv.FormatFloat(value, 'f', -1, 64)),
				MeasureValueType: aws.String(timestreamwrite.MeasureValueTypeDouble),
				Time:             aws.String(strconv.FormatInt(c.now().UnixMilli(), 10)),
				TimeUnit:         aws.String(timestreamwrite.TimeUnitMilliseconds),
			},
		},
	})
	if err != nil {
		log.Printf("telemetry: failed to write metric %q to timestream: %v", name, err)
	}
}

func timestreamConfigFromEnv() TimestreamConfig {
	table := strings.TrimSpace(os.Getenv(EnvTimestreamTable))
	if table == "" {
		table = DefaultTimestreamTable
	}

	region := strings.TrimSpace(os.Getenv(EnvTimestreamRegion))
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_REGION"))
	}
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_DEFAULT_REGION"))
	}

	return TimestreamConfig{
		Database: strings.TrimSpace(os.Getenv(EnvTimestreamDatabase)),
		Table:    table,
		Region:   region,
	}
}

func (cfg TimestreamConfig) validate() error {
	if cfg.Database == "" {
		return fmt.Errorf("%s must be set when using the timestream telemetry backend", EnvTimestreamDatabase)
	}
	if cfg.Table == "" {
		return fmt.Errorf("%s must be set when using the timestream telemetry backend", EnvTimestreamTable)
	}
	if cfg.Region == "" {
		return fmt.Errorf("%s or AWS_REGION must be set when using the timestream telemetry backend", EnvTimestreamRegion)
	}

	return nil
}

func buildTimestreamDimensions(metricName string, tags []Tag) []*timestreamwrite.Dimension {
	dimensions := make([]*timestreamwrite.Dimension, 0, len(tags)+1)
	indexByName := map[string]int{}

	appendDimension := func(name, value string) {
		sanitizedName := sanitizeDimensionName(name)
		if index, ok := indexByName[sanitizedName]; ok {
			dimensions[index] = &timestreamwrite.Dimension{
				Name:  aws.String(sanitizedName),
				Value: aws.String(value),
			}
			return
		}

		indexByName[sanitizedName] = len(dimensions)
		dimensions = append(dimensions, &timestreamwrite.Dimension{
			Name:  aws.String(sanitizedName),
			Value: aws.String(value),
		})
	}

	appendDimension("metric_name", metricName)
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
