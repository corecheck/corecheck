package main

import "github.com/corecheck/corecheck/internal/config"

type Config struct {
	config.DatabaseConfig

	BucketDataURL string `env:"BUCKET_DATA_URL" env-required:"true"`
}
