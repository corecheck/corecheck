package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type DatabaseConfig struct {
	Database struct {
		Host     string `env:"HOST" env-required:"true"`
		Port     string `env:"PORT" env-default:"5432"`
		User     string `env:"USER" env-required:"true"`
		Password string `env:"PASSWORD" env-required:"true"`
		Name     string `env:"NAME" env-default:"corecheck"`
	} `env-prefix:"DATABASE"`
}

type AWSConfig struct {
	AWS struct {
		AwsAccessKeyID     string `env:"AWS_ACCESS_KEY_ID" env-required:"true"`
		AwsSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY" env-required:"true"`
		AwsRegion          string `env:"AWS_REGION" env-required:"true"`
	}
}

func Load(cfg interface{}) error {
	return cleanenv.ReadEnv(&cfg)
}
