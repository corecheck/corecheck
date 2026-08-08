package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadAndExtract downloads the zip archive from url, extracts it into destDir,
// and returns the path of the top-level extracted directory.
func DownloadAndExtract(url, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	log.Printf("stats: downloading zip from %s", url)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	tmpZip := filepath.Join(destDir, "archive.zip")
	f, err := os.Create(tmpZip)
	if err != nil {
		return "", fmt.Errorf("create temp zip: %w", err)
	}
	written, err := io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		os.Remove(tmpZip)
		return "", fmt.Errorf("write zip: %w", err)
	}
	log.Printf("stats: download complete (%.1f MB)", float64(written)/(1024*1024))

	log.Printf("stats: extracting zip to %s", destDir)
	extracted, err := extractZip(tmpZip, destDir)
	os.Remove(tmpZip)
	if err != nil {
		return "", err
	}
	log.Printf("stats: extraction complete, repo path: %s", extracted)
	return extracted, nil
}

func extractZip(src, dest string) (string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	log.Printf("stats: zip contains %d entries", len(r.File))

	cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
	var topDir string

	for _, f := range r.File {
		outPath := filepath.Join(dest, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(outPath, cleanDest) {
			return "", fmt.Errorf("zip path traversal: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return "", err
			}
			if topDir == "" {
				topDir = strings.SplitN(f.Name, "/", 2)[0]
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return "", err
		}
		if err := extractFile(f, outPath); err != nil {
			return "", err
		}
	}

	if topDir == "" {
		return dest, nil
	}
	return filepath.Join(dest, topDir), nil
}

func extractFile(f *zip.File, outPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

