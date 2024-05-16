package main

import (
	"bitcoin-stats-datadog/types"
	"encoding/json"
	"os"
	"path/filepath"
)

func (bc *BitcoinCoreData) GetNumberOfPulls() (total float64, open float64, closed float64, merged float64) {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
	if err != nil {
		panic(err)
	}

	total = float64(len(dir))
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

		if pull.Pull.State == "open" {
			open++
		}

		if pull.Pull.State == "closed" {
			closed++
		}

		if !pull.Pull.MergedAt.IsZero() {
			merged++
		}
	}

	return total, open, closed, merged
}

func (bc *BitcoinCoreData) GetUniqueAuthors(hasAtLeastOneMerge bool) []string {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
	if err != nil {
		panic(err)
	}

	users := make(map[string]struct{})

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

		if hasAtLeastOneMerge && pull.Pull.MergedAt.IsZero() {
			continue
		}

		users[pull.Pull.User.Login] = struct{}{}
	}

	var uniqueUsers []string
	for user := range users {
		uniqueUsers = append(uniqueUsers, user)
	}

	return uniqueUsers
}

func (bc *BitcoinCoreData) GetPullsByUser() (open map[string]float64, closed map[string]float64, merged map[string]float64) {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
	if err != nil {
		panic(err)
	}

	open = make(map[string]float64)
	closed = make(map[string]float64)
	merged = make(map[string]float64)

	uniqueUsers := bc.GetUniqueAuthors(false)
	for _, user := range uniqueUsers {
		open[user] = 0
		closed[user] = 0
		merged[user] = 0
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

		if pull.Pull.State == "open" {
			open[pull.Pull.User.Login]++
		}

		if pull.Pull.State == "closed" {
			closed[pull.Pull.User.Login]++
		}

		if !pull.Pull.MergedAt.IsZero() {
			merged[pull.Pull.User.Login]++
		}
	}

	return open, closed, merged
}

func (bc *BitcoinCoreData) getUniqueLabels() []string {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
	if err != nil {
		panic(err)
	}

	labels := make(map[string]struct{})

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

		for _, label := range pull.Pull.Labels {
			labels[label.Name] = struct{}{}
		}
	}

	var uniqueLabels []string
	for label := range labels {
		uniqueLabels = append(uniqueLabels, label)
	}

	return uniqueLabels
}

func (bc *BitcoinCoreData) GetPullsByLabel() (open map[string]float64, closed map[string]float64, merged map[string]float64) {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
	if err != nil {
		panic(err)
	}

	open = make(map[string]float64)
	closed = make(map[string]float64)
	merged = make(map[string]float64)

	uniqueLabels := bc.getUniqueLabels()
	for _, label := range uniqueLabels {
		open[label] = 0
		closed[label] = 0
		merged[label] = 0
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

		for _, label := range pull.Pull.Labels {
			if pull.Pull.State == "open" {
				open[label.Name]++
			}

			if pull.Pull.State == "closed" {
				closed[label.Name]++
			}

			if !pull.Pull.MergedAt.IsZero() {
				merged[label.Name]++
			}
		}
	}

	return open, closed, merged
}

func (bc *BitcoinCoreData) GetTotalCommentsAndReviewsByPull() (comments map[int]float64, reviews map[int]float64) {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
	if err != nil {
		panic(err)
	}

	comments = make(map[int]float64)
	reviews = make(map[int]float64)

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

		for _, event := range pull.Events {
			if event.Event == "commented" {
				comments[pull.Pull.Number]++
			} else if event.Event == "reviewed" {
				reviews[pull.Pull.Number]++
			}
		}
	}

	return comments, reviews
}
