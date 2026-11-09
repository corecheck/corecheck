package main

import (
	"testing"
	"time"

	"github.com/google/go-github/v57/github"
)

func pull(number int, updated time.Time) *github.PullRequest {
	stamp := github.Timestamp{Time: updated}
	return &github.PullRequest{
		Number:    github.Int(number),
		UpdatedAt: &stamp,
	}
}

func numbers(prs []*github.PullRequest) []int {
	out := make([]int, len(prs))
	for i, pr := range prs {
		out[i] = pr.GetNumber()
	}
	return out
}

func TestPRsUpdatedSinceKeepsBacklogOlderThanTwoHours(t *testing.T) {
	cursor := time.Date(2026, 10, 19, 22, 40, 0, 0, time.UTC)
	newestFirst := []*github.PullRequest{
		pull(30, time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)),
		pull(20, time.Date(2026, 10, 21, 3, 0, 0, 0, time.UTC)),
		pull(10, cursor.Add(-time.Minute)),
	}

	got := numbers(prsUpdatedSince(newestFirst, cursor))
	want := []int{30, 20}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("updated since cursor = %v, want %v", got, want)
	}
}

func TestSelectPRSyncWalksOldestFirstAndDefersPastTheJobCap(t *testing.T) {
	cursor := time.Date(2026, 10, 19, 22, 40, 0, 0, time.UTC)
	newestFirst := []*github.PullRequest{
		pull(4, time.Date(2026, 10, 23, 4, 0, 0, 0, time.UTC)),
		pull(3, time.Date(2026, 10, 22, 3, 0, 0, 0, time.UTC)),
		pull(2, time.Date(2026, 10, 21, 2, 0, 0, 0, time.UTC)),
		pull(1, time.Date(2026, 10, 20, 1, 0, 0, 0, time.UTC)),
		pull(0, cursor.Add(-time.Hour)),
	}
	needsJob := map[int]bool{1: true, 2: true, 4: true}

	got := numbers(selectPRSync(newestFirst, cursor, 1, func(pr *github.PullRequest) bool {
		return needsJob[pr.GetNumber()]
	}))
	// #1 is the oldest job. #2 also needs a job, so the run stops before it
	// and does not record the newer metadata update on #3.
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("selected = %v, want [1]", got)
	}

	between := []*github.PullRequest{
		pull(3, time.Date(2026, 10, 22, 3, 0, 0, 0, time.UTC)),
		pull(2, time.Date(2026, 10, 21, 2, 0, 0, 0, time.UTC)),
		pull(1, time.Date(2026, 10, 20, 1, 0, 0, 0, time.UTC)),
	}
	betweenNeeds := map[int]bool{1: true, 3: true}
	got = numbers(selectPRSync(between, cursor, 1, func(pr *github.PullRequest) bool {
		return betweenNeeds[pr.GetNumber()]
	}))
	want := []int{1, 2}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("selected with metadata between jobs = %v, want %v", got, want)
	}
}
