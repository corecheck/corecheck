package main

import "github.com/corecheck/corecheck/internal/config"

type Config struct {
	config.DatabaseConfig
	config.AWSConfig

	Github struct {
		AccessToken string `env:"ACCESS_TOKEN" env-required:"true"`
	} `env-prefix:"GITHUB"`

	SQS struct {
		QueueURL string `env:"QUEUE_URL" env-required:"true"`
	} `env-prefix:"SQS"`
}
