package main

import "github.com/corecheck/corecheck/internal/config"

type Config struct {
	config.DatabaseConfig
	config.AWSConfig

	Github struct {
		AccessToken string `env:"ACCESS_TOKEN" env-required:"true"`
	} `env-prefix:"GITHUB_"`

	StateMachineARN    string `env:"STATE_MACHINE_ARN" env-required:"true"`
	MutationMachineARN string `env:"MUTATION_STATE_MACHINE_ARN" env-required:"true"`
}
