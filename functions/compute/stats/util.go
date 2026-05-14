package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitClone performs a shallow clone of url into dest and returns the HEAD commit SHA.
// depth=100 is enough to span many hours of an hourly lambda, giving enough history
// to compare against a previously stored SHA.
func GitClone(url, dest string) (string, error) {
	r, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:          url,
		Depth:        100,
		SingleBranch: true,
	})
	if err != nil {
		return "", fmt.Errorf("git clone: %w", err)
	}

	head, err := r.Head()
	if err != nil {
		return "", fmt.Errorf("git head: %w", err)
	}
	return head.Hash().String(), nil
}

// GitChangedFiles returns the list of file paths that differ between prevSHA and HEAD.
// The second return value is false when prevSHA is not present in the shallow history,
// signalling that the caller should fall back to a full first-run scan.
func GitChangedFiles(repo, prevSHA string) ([]string, bool) {
	r, err := git.PlainOpen(repo)
	if err != nil {
		return nil, false
	}

	head, err := r.Head()
	if err != nil {
		return nil, false
	}
	headCommit, err := r.CommitObject(head.Hash())
	if err != nil {
		return nil, false
	}
	prevCommit, err := r.CommitObject(plumbing.NewHash(prevSHA))
	if err != nil {
		// prevSHA not in shallow history.
		return nil, false
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, false
	}
	prevTree, err := prevCommit.Tree()
	if err != nil {
		return nil, false
	}

	changes, err := object.DiffTree(prevTree, headTree)
	if err != nil {
		return nil, false
	}

	seen := make(map[string]struct{})
	for _, ch := range changes {
		if ch.To.Name != "" {
			seen[ch.To.Name] = struct{}{}
		}
		if ch.From.Name != "" {
			seen[ch.From.Name] = struct{}{}
		}
	}

	files := make([]string, 0, len(seen))
	for f := range seen {
		files = append(files, f)
	}
	return files, true
}

// GitShowFile returns the raw bytes of relPath at the given commit SHA.
func GitShowFile(repo, sha, relPath string) ([]byte, error) {
	r, err := git.PlainOpen(repo)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	commit, err := r.CommitObject(plumbing.NewHash(sha))
	if err != nil {
		return nil, fmt.Errorf("commit %s: %w", sha, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("tree at %s: %w", sha, err)
	}

	// relPath may use OS separators; normalise to forward slashes for git tree lookup.
	relPath = filepath.ToSlash(relPath)
	file, err := tree.File(relPath)
	if err != nil {
		return nil, fmt.Errorf("file %s at %s: %w", relPath, sha, err)
	}

	rc, err := file.Reader()
	if err != nil {
		return nil, fmt.Errorf("reader %s: %w", relPath, err)
	}
	defer rc.Close()

	return io.ReadAll(rc)
}
