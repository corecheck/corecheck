package main

import (
	"io"
	"net/http"
	"os"

	"github.com/artdarek/go-unzip"
)

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
