package main

import "github.com/corecheck/corecheck/internal/config"

type Config struct {
	config.DatabaseConfig

	BenchArraySize int `envconfig:"BENCH_ARRAY_SIZE" default:"5"`
}
