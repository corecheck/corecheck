package main

import (
	"bitcoin-stats-datadog/types"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type StatsConsumer interface {
	Init()
	ProcessPull(pull *types.Pull)
	ProcessIssue(issue *types.Issue)
	SendMetrics()
}

type BitcoinCoreData struct {
	Path      string
	Consumers []StatsConsumer
}

func (bc *BitcoinCoreData) AddConsumers(consumer ...StatsConsumer) {
	bc.Consumers = append(bc.Consumers, consumer...)
}

func (bc *BitcoinCoreData) init() {
	for _, consumer := range bc.Consumers {
		consumer.Init()
	}
}

func (bc *BitcoinCoreData) processPull(pull *types.Pull) {
	for _, consumer := range bc.Consumers {
		consumer.ProcessPull(pull)
	}
}

func (bc *BitcoinCoreData) processPulls() {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
	if err != nil {
		panic(err)
	}

	for _, entry := range dir {
		pullRaw, err := os.ReadFile(filepath.Join(bc.Path, "pulls", entry.Name()))
		if err != nil {
			panic(err)
		}
		pull := types.Pull{}
		err = json.Unmarshal(pullRaw, &pull)
		if err != nil {
			panic(err)
		}

		bc.processPull(&pull)
	}
}

func (bc *BitcoinCoreData) processIssue(issue *types.Issue) {
	for _, consumer := range bc.Consumers {
		consumer.ProcessIssue(issue)
	}
}

func (bc *BitcoinCoreData) processIssues() {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "issues"))
	if err != nil {
		panic(err)
	}

	for _, entry := range dir {
		issueRaw, err := os.ReadFile(filepath.Join(bc.Path, "issues", entry.Name()))
		if err != nil {
			panic(err)
		}
		issue := types.Issue{}
		err = json.Unmarshal(issueRaw, &issue)
		if err != nil {
			panic(err)
		}

		bc.processIssue(&issue)
	}
}

func (bc *BitcoinCoreData) sendMetrics() {
	for _, consumer := range bc.Consumers {
		consumer.SendMetrics()
	}
}

func (bc *BitcoinCoreData) Run() {
	bc.init()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		bc.processPulls()
	}()
	go func() {
		defer wg.Done()
		bc.processIssues()
	}()

	wg.Wait()
	bc.sendMetrics()
}
