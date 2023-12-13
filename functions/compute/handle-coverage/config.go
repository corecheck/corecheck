package main

import "github.com/corecheck/corecheck/internal/config"

type Config struct {
	config.DatabaseConfig

	Github struct {
		AccessToken string `env:"ACCESS_TOKEN" env-required:"true"`
	} `env-prefix:"GITHUB_"`
}
