package main

import "github.com/corecheck/corecheck/internal/config"

type Config struct {
	config.DatabaseConfig
	config.AWSConfig

	StateMachineARN string `env:"STATE_MACHINE_ARN" env-required:"true"`
}
