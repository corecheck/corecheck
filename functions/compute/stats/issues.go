package main

import (
	"bitcoin-stats-datadog/types"
	"encoding/json"
	"os"
	"path/filepath"
)

//func (bc *BitcoinCoreData) GetNumberOfPulls() (total int, open int, closed int, merged int) {
//	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
//	if err != nil {
//		panic(err)
//	}
//
//	total = len(dir)
//	for _, entry := range dir {
//		pullRaw, err := os.ReadFile(filepath.Join(bc.Path, "pulls", entry.Name()))
//		if err != nil {
//			panic(err)
//		}
//		pull := types.Pull{}
//		err = json.Unmarshal(pullRaw, &pull)
//		if err != nil {
//			panic(err)
//		}
//
//		if pull.Pull.State == "open" {
//			open++
//		}
//
//		if pull.Pull.State == "closed" {
//			closed++
//		}
//
//		if !pull.Pull.MergedAt.IsZero() {
//			merged++
//		}
//	}
//
//	return total, open, closed, merged
//}
//
//func (bc *BitcoinCoreData) GetUniqueUsers() []string {
//	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
//	if err != nil {
//		panic(err)
//	}
//
//	users := make(map[string]struct{})
//
//	for _, entry := range dir {
//		pullRaw, err := os.ReadFile(filepath.Join(bc.Path, "pulls", entry.Name()))
//		if err != nil {
//			panic(err)
//		}
//		pull := types.Pull{}
//		err = json.Unmarshal(pullRaw, &pull)
//		if err != nil {
//			panic(err)
//		}
//
//		users[pull.Pull.User.Login] = struct{}{}
//	}
//
//	var uniqueUsers []string
//	for user := range users {
//		uniqueUsers = append(uniqueUsers, user)
//	}
//
//	return uniqueUsers
//}
//
//func (bc *BitcoinCoreData) GetPullsByUser() (open map[string]int, closed map[string]int, merged map[string]int) {
//	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
//	if err != nil {
//		panic(err)
//	}
//
//	open = make(map[string]int)
//	closed = make(map[string]int)
//	merged = make(map[string]int)
//
//	uniqueUsers := bc.GetUniqueUsers()
//	for _, user := range uniqueUsers {
//		open[user] = 0
//		closed[user] = 0
//		merged[user] = 0
//	}
//
//	for _, entry := range dir {
//		pullRaw, err := os.ReadFile(filepath.Join(bc.Path, "pulls", entry.Name()))
//		if err != nil {
//			panic(err)
//		}
//		pull := types.Pull{}
//		err = json.Unmarshal(pullRaw, &pull)
//		if err != nil {
//			panic(err)
//		}
//
//		if pull.Pull.State == "open" {
//			open[pull.Pull.User.Login]++
//		}
//
//		if pull.Pull.State == "closed" {
//			closed[pull.Pull.User.Login]++
//		}
//
//		if !pull.Pull.MergedAt.IsZero() {
//			merged[pull.Pull.User.Login]++
//		}
//	}
//
//	return open, closed, merged
//}
//
//func (bc *BitcoinCoreData) getUniqueLabels() []string {
//	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
//	if err != nil {
//		panic(err)
//	}
//
//	labels := make(map[string]struct{})
//
//	for _, entry := range dir {
//		pullRaw, err := os.ReadFile(filepath.Join(bc.Path, "pulls", entry.Name()))
//		if err != nil {
//			panic(err)
//		}
//		pull := types.Pull{}
//		err = json.Unmarshal(pullRaw, &pull)
//		if err != nil {
//			panic(err)
//		}
//
//		for _, label := range pull.Pull.Labels {
//			labels[label.Name] = struct{}{}
//		}
//	}
//
//	var uniqueLabels []string
//	for label := range labels {
//		uniqueLabels = append(uniqueLabels, label)
//	}
//
//	return uniqueLabels
//}
//
//func (bc *BitcoinCoreData) GetPullsByLabel() (open map[string]int, closed map[string]int, merged map[string]int) {
//	dir, err := os.ReadDir(filepath.Join(bc.Path, "pulls"))
//	if err != nil {
//		panic(err)
//	}
//
//	open = make(map[string]int)
//	closed = make(map[string]int)
//	merged = make(map[string]int)
//
//	uniqueLabels := bc.getUniqueLabels()
//	for _, label := range uniqueLabels {
//		open[label] = 0
//		closed[label] = 0
//		merged[label] = 0
//	}
//
//	for _, entry := range dir {
//		pullRaw, err := os.ReadFile(filepath.Join(bc.Path, "pulls", entry.Name()))
//		if err != nil {
//			panic(err)
//		}
//		pull := types.Pull{}
//		err = json.Unmarshal(pullRaw, &pull)
//		if err != nil {
//			panic(err)
//		}
//
//		for _, label := range pull.Pull.Labels {
//			if pull.Pull.State == "open" {
//				open[label.Name]++
//			}
//
//			if pull.Pull.State == "closed" {
//				closed[label.Name]++
//			}
//
//			if !pull.Pull.MergedAt.IsZero() {
//				merged[label.Name]++
//			}
//		}
//	}
//
//	return open, closed, merged
//}

func (bc *BitcoinCoreData) GetNumberOfIssues() (total float64, open float64, closed float64) {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "issues"))
	if err != nil {
		panic(err)
	}

	total = float64(len(dir))
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

		if issue.Issue.State == "open" {
			open++
		}

		if issue.Issue.State == "closed" {
			closed++
		}
	}

	return total, open, closed
}

func (bc *BitcoinCoreData) GetUniqueIssueUsers() []string {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "issues"))
	if err != nil {
		panic(err)
	}

	users := make(map[string]struct{})

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

		users[issue.Issue.User.Login] = struct{}{}
	}

	var uniqueUsers []string
	for user := range users {
		uniqueUsers = append(uniqueUsers, user)
	}

	return uniqueUsers
}

func (bc *BitcoinCoreData) GetIssuesByUser() (open map[string]float64, closed map[string]float64) {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "issues"))
	if err != nil {
		panic(err)
	}

	open = make(map[string]float64)
	closed = make(map[string]float64)

	uniqueUsers := bc.GetUniqueIssueUsers()
	for _, user := range uniqueUsers {
		open[user] = 0
		closed[user] = 0
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

		if issue.Issue.State == "open" {
			open[issue.Issue.User.Login]++
		}

		if issue.Issue.State == "closed" {
			closed[issue.Issue.User.Login]++
		}
	}

	return open, closed
}

func (bc *BitcoinCoreData) GetUniqueIssueLabels() []string {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "issues"))
	if err != nil {
		panic(err)
	}

	labels := make(map[string]struct{})

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

		for _, label := range issue.Issue.Labels {
			labels[label.Name] = struct{}{}
		}
	}

	var uniqueLabels []string
	for label := range labels {
		uniqueLabels = append(uniqueLabels, label)
	}

	return uniqueLabels
}

func (bc *BitcoinCoreData) GetIssuesByLabel() (open map[string]float64, closed map[string]float64) {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "issues"))
	if err != nil {
		panic(err)
	}

	open = make(map[string]float64)
	closed = make(map[string]float64)

	uniqueLabels := bc.GetUniqueIssueLabels()
	for _, label := range uniqueLabels {
		open[label] = 0
		closed[label] = 0
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

		for _, label := range issue.Issue.Labels {
			if issue.Issue.State == "open" {
				open[label.Name]++
			}

			if issue.Issue.State == "closed" {
				closed[label.Name]++
			}
		}
	}

	return open, closed
}

func (bc *BitcoinCoreData) GetTotalCommentsByIssue() (comments map[int]int) {
	dir, err := os.ReadDir(filepath.Join(bc.Path, "issues"))
	if err != nil {
		panic(err)
	}

	comments = make(map[int]int)

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

		for _, event := range issue.Events {
			if event.Event == "commented" {
				comments[issue.Issue.Number]++
			}
		}
	}

	return comments
}
