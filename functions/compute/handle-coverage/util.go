package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/waigani/diffparser"
)

func GetCoverageData(prNum int, commit string) (*RawCoverageData, error) {
	return getCoverageData("https://bitcoin-coverage-data-default.s3.eu-west-3.amazonaws.com/" + strconv.Itoa(prNum) + "/" + commit + "/coverage.json")
}

func GetCoverageDataMaster(commit string) (*RawCoverageData, error) {
	return getCoverageData("https://bitcoin-coverage-data-default.s3.eu-west-3.amazonaws.com/master/" + commit + "/coverage.json")
}

func getCoverageData(url string) (*RawCoverageData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	coverageData := RawCoverageData{}
	err = json.NewDecoder(resp.Body).Decode(&coverageData)
	if err != nil {
		return nil, err
	}

	return &coverageData, nil
}

func GetPullDiff(prNum int) (*diffparser.Diff, error) {
	resp, err := http.Get("https://patch-diff.githubusercontent.com/raw/bitcoin/bitcoin/pull/" + strconv.Itoa(prNum) + ".diff")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return diffparser.Parse(string(data))
}
func getSourceFile(filename string, commit string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/bitcoin/bitcoin/%s/%s", commit, filename))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
func fetchAllFiles(files []string, commit string) map[string][]string {
	var wg sync.WaitGroup
	var filesMap = make(map[string][]string)
	mut := sync.Mutex{}

	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			sourceCodeFile, err := getSourceFile(file, commit)
			if err != nil {
				log.Error(err)
				return
			}

			mut.Lock()
			filesMap[file] = strings.Split(sourceCodeFile, "\n")
			mut.Unlock()
		}(file)
	}
	wg.Wait()

	return filesMap
}

func fetchPullFiles(files []string, commit string, baseCommit string) (map[string][]string, error) {
	r, err := git.PlainClone("/tmp/bitcoin", false, &git.CloneOptions{
		URL: "https://github.com/bitcoin/bitcoin.git",
	})
	if err != nil {
		return nil, err
	}

	// git fetch origin {commit}
	err = r.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(commit),
		},
	})
	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commit),
	})
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("git", "rebase", baseCommit)
	cmd.Dir = "/tmp/bitcoin"
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	var filesMap = make(map[string][]string)
	for _, file := range files {
		fileContent, err := os.ReadFile("/tmp/bitcoin/" + file)
		if err != nil {
			return nil, err
		}

		filesMap[file] = strings.Split(string(fileContent), "\n")
	}

	return filesMap, nil
}
