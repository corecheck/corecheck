package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/waigani/diffparser"
)

func GetCoverageData(prNum int, commit string) (*RawCoverageData, error) {
	return getCoverageData(os.Getenv("BUCKET_DATA_URL") + "/" + strconv.Itoa(prNum) + "/" + commit + "/coverage.json")
}

func GetCoverageDataMaster(commit string) (*RawCoverageData, error) {
	return getCoverageData(os.Getenv("BUCKET_DATA_URL") + "/master/" + commit + "/coverage.json")
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

func GetPullDiff(baseCommit string, commit string) (*diffparser.Diff, error) {
	resp, err := http.Get(fmt.Sprintf("https://github.com/bitcoin/bitcoin/compare/%s..%s.diff", baseCommit, commit))
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

func getSourceFileMaster(filename string, commit string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(os.Getenv("BUCKET_DATA_URL")+"/master/%s/%s", commit, filename))
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

func getSourceFilePull(pullNumber int, filename string, commit string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(os.Getenv("BUCKET_DATA_URL")+"/%d/%s/%s", pullNumber, commit, filename))
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

func fetchAllFiles(pullNumber int, files []string, commit string) map[string][]string {
	var wg sync.WaitGroup
	var filesMap = make(map[string][]string)
	mut := sync.Mutex{}

	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			sourceCodeFile, err := getSourceFilePull(pullNumber, file, commit)
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

func fetchAllFilesMaster(files []string, commit string) map[string][]string {
	var wg sync.WaitGroup
	var filesMap = make(map[string][]string)
	mut := sync.Mutex{}

	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			sourceCodeFile, err := getSourceFileMaster(file, commit)
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
