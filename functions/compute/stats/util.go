package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GitClone performs a shallow clone of url into dest and returns the HEAD commit SHA.
// depth=100 is enough to span many hours of an hourly lambda, giving git diff enough
// history to compare against a previously stored SHA.
func GitClone(url, dest string) (string, error) {
	cmd := exec.Command("git", "clone", "--depth=100", "--single-branch", url, dest)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone: %w: %s", err, stderr.String())
	}

	sha, err := gitRevParse(dest, "HEAD")
	if err != nil {
		return "", err
	}
	return sha, nil
}

// GitChangedFiles returns the list of file paths that differ between prevSHA and HEAD.
// The second return value is false when prevSHA is not present in the shallow history,
// signalling that the caller should fall back to a full first-run scan.
func GitChangedFiles(repo, prevSHA string) ([]string, bool) {
	cmd := exec.Command("git", "-C", repo, "diff", "--name-only", prevSHA, "HEAD")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// prevSHA is likely not in the shallow history.
		return nil, false
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, true
}

// GitShowFile returns the raw bytes of relPath at the given commit SHA.
func GitShowFile(repo, sha, relPath string) ([]byte, error) {
	cmd := exec.Command("git", "-C", repo, "show", sha+":"+relPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git show %s:%s: %w: %s", sha, relPath, err, stderr.String())
	}
	return stdout.Bytes(), nil
}

func gitRevParse(repo, ref string) (string, error) {
	cmd := exec.Command("git", "-C", repo, "rev-parse", ref)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse %s: %w: %s", ref, err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}
