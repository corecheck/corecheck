package main

import (
	"time"

	"github.com/google/go-github/v57/github"
)

// Cap new coverage jobs so catching up after an outage cannot fill the
// batch queue in one invocation. The scheduler runs every 10 minutes.
const maxNewCoverageJobsPerSync = 8

// prsUpdatedSince returns the prefix of a newest-first page that is still
// newer than the database cursor. The two-hour window used to live here and
// dropped every pull request updated during a longer outage.
func prsUpdatedSince(newestFirst []*github.PullRequest, lastDBUpdate time.Time) []*github.PullRequest {
	updated := make([]*github.PullRequest, 0, len(newestFirst))
	for _, pr := range newestFirst {
		if pr.GetUpdatedAt().Before(lastDBUpdate) {
			break
		}
		updated = append(updated, pr)
	}
	return updated
}

func oldestFirst(prs []*github.PullRequest) []*github.PullRequest {
	ordered := make([]*github.PullRequest, len(prs))
	for i, pr := range prs {
		ordered[len(prs)-1-i] = pr
	}
	return ordered
}

// selectPRSync chooses which pull requests to write on this run.
//
// GitHub returns updates newest first, but the database cursor is the newest
// updated_at already stored. Writing the newest rows first would move that
// cursor past anything left queued. Walk oldest first instead, and stop
// before a pull request that needs a coverage job once the cap is reached.
func selectPRSync(newestFirst []*github.PullRequest, lastDBUpdate time.Time, jobLimit int, needsJob func(*github.PullRequest) bool) []*github.PullRequest {
	ordered := oldestFirst(prsUpdatedSince(newestFirst, lastDBUpdate))
	selected := make([]*github.PullRequest, 0, len(ordered))
	started := 0
	for _, pr := range ordered {
		job := needsJob != nil && needsJob(pr)
		if job && started >= jobLimit {
			break
		}
		selected = append(selected, pr)
		if job {
			started++
		}
	}
	return selected
}
