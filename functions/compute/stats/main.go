package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	"github.com/artdarek/go-unzip"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	bitcoinDataURL = "https://github.com/bitcoin-data/github-metadata-backup-bitcoin-bitcoin/archive/refs/heads/master.zip"
	dest           = "/tmp/data"
)

type BitcoinCoreData struct {
	Path string
}

func DownloadFile(url string) error {
	// Get the data
	out, err := os.Create("/tmp/bitcoin-data.zip")
	if err != nil {
		return err
	}

	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func Unzip(src string, dest string) error {
	uz := unzip.New(src, dest)
	return uz.Extract()
}

func handleMetrics(ctx context.Context) (string, error) {

	// Download zip file
	err := DownloadFile(bitcoinDataURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unzip file
	err = Unzip("/tmp/bitcoin-data.zip", dest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bc := BitcoinCoreData{Path: dest + "/github-metadata-backup-bitcoin-bitcoin-master"}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		total, open, closed, merged := bc.GetNumberOfPulls()
		ddlambda.Metric("bitcoin.bitcoin.pulls.open", open)
		ddlambda.Metric("bitcoin.bitcoin.pulls.closed", closed)
		ddlambda.Metric("bitcoin.bitcoin.pulls.merged", merged)
		ddlambda.Metric("bitcoin.bitcoin.pulls.total", total)

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		total, open, closed := bc.GetNumberOfIssues()
		ddlambda.Metric("bitcoin.bitcoin.issues.open", open)
		ddlambda.Metric("bitcoin.bitcoin.issues.closed", closed)
		ddlambda.Metric("bitcoin.bitcoin.issues.total", total)

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		openByLabels, closedByLabels, mergedByLabels := bc.GetPullsByLabel()
		for label, count := range openByLabels {
			ddlambda.Metric("bitcoin.bitcoin.pulls.open.by_label", count, "label:"+label)
		}
		for label, count := range closedByLabels {
			ddlambda.Metric("bitcoin.bitcoin.pulls.closed.by_label", count, "label:"+label)
		}
		for label, count := range mergedByLabels {
			ddlambda.Metric("bitcoin.bitcoin.pulls.merged.by_label", count, "label:"+label)
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		openByLabels, closedByLabels := bc.GetIssuesByLabel()
		for label, count := range openByLabels {
			ddlambda.Metric("bitcoin.bitcoin.issues.open.by_label", count, "label:"+label)
		}
		for label, count := range closedByLabels {
			ddlambda.Metric("bitcoin.bitcoin.issues.closed.by_label", count, "label:"+label)
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		uniqueAuthors := bc.GetUniqueAuthors(true)
		ddlambda.Metric("bitcoin.bitcoin.pulls.unique_authors", float64(len(uniqueAuthors)))

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		comments, reviews := bc.GetTotalCommentsAndReviewsByPull()
		for pull, count := range comments {
			ddlambda.Metric("bitcoin.bitcoin.pulls.comments", count, "pull:"+strconv.Itoa(pull))
		}
		for pull, count := range reviews {
			ddlambda.Metric("bitcoin.bitcoin.pulls.reviews", count, "pull:"+strconv.Itoa(pull))
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		openByUser, closedByUser, mergedByUser := bc.GetPullsByUser()
		for user, count := range openByUser {
			ddlambda.Metric("bitcoin.bitcoin.pulls.open.by_user", count, "user:"+user)
		}
		for user, count := range closedByUser {
			ddlambda.Metric("bitcoin.bitcoin.pulls.closed.by_user", count, "user:"+user)
		}
		for user, count := range mergedByUser {
			ddlambda.Metric("bitcoin.bitcoin.pulls.merged.by_user", count, "user:"+user)
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		openByUser, closedByUser := bc.GetIssuesByUser()
		for user, count := range openByUser {
			ddlambda.Metric("bitcoin.bitcoin.issues.open.by_user", count, "user:"+user)
		}
		for user, count := range closedByUser {
			ddlambda.Metric("bitcoin.bitcoin.issues.closed.by_user", count, "user:"+user)
		}

		wg.Done()
	}()

	wg.Wait()

	return "OK", nil
}

func main() {
	lambda.Start(ddlambda.WrapFunction(handleMetrics, &ddlambda.Config{
		DebugLogging:  true,
		Site:          "datadoghq.eu",
		BatchInterval: time.Millisecond * 500,
	}))
}
