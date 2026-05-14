package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/corecheck/corecheck/functions/compute/stats/types"
)

// cwLogsMaxAge is the CloudWatch Logs maximum lookback for event timestamps.
const cwLogsMaxAge = 14 * 24 * time.Hour

// eventTypesToEmit are the PR/issue event kinds we care about.
// Structural noise (committed, referenced, subscribed, mentioned, etc.) is excluded.
var eventTypesToEmit = map[string]bool{
	"commented":  true,
	"reviewed":   true,
	"labeled":    true,
	"unlabeled":  true,
	"closed":     true,
	"reopened":   true,
	"merged":     true,
	"assigned":   true,
	"unassigned": true,
	"locked":     true,
	"unlocked":   true,
}

// LogEvent is the JSON object written as a single CloudWatch Logs log line.
type LogEvent struct {
	SourceType string   `json:"source_type"`
	EventType  string   `json:"event_type"`
	Number     int      `json:"number"`
	Title      string   `json:"title"`
	User       string   `json:"user"`
	Actor      string   `json:"actor"`
	Labels     []string `json:"labels"`
	State      string   `json:"state"`
	EventTime  string   `json:"event_time"`
	IngestedAt string   `json:"ingested_at"`
}

// EventStreamProducer processes changed (or all) PR/issue files and writes
// LogEvents to CloudWatch Logs.
type EventStreamProducer struct {
	repoPath     string
	prevSHA      string   // empty = first run
	changedFiles []string // nil = first run / fallback (process all files)
	writer       *CWLogsWriter
	cutoff       time.Time
	ingestedAt   string
}

func NewEventStreamProducer(repoPath, prevSHA string, changedFiles []string, writer *CWLogsWriter) *EventStreamProducer {
	return &EventStreamProducer{
		repoPath:     repoPath,
		prevSHA:      prevSHA,
		changedFiles: changedFiles,
		writer:       writer,
		cutoff:       time.Now().UTC().Add(-cwLogsMaxAge),
		ingestedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}

// Run processes all relevant files and flushes events to CloudWatch Logs.
func (p *EventStreamProducer) Run() error {
	var pullFiles, issueFiles []string

	if p.changedFiles == nil {
		// First run / fallback: walk the entire repo.
		pullFiles = p.listFiles("pulls")
		issueFiles = p.listFiles("issues")
	} else {
		for _, f := range p.changedFiles {
			dir := filepath.Dir(f)
			switch dir {
			case "pulls":
				pullFiles = append(pullFiles, f)
			case "issues":
				issueFiles = append(issueFiles, f)
			}
		}
	}

	var events []*cloudwatchlogs.InputLogEvent

	for _, relPath := range pullFiles {
		evts, err := p.processChangedPull(relPath)
		if err != nil {
			log.Printf("eventstream: skip pull %s: %v", relPath, err)
			continue
		}
		events = append(events, evts...)
	}

	for _, relPath := range issueFiles {
		evts, err := p.processChangedIssue(relPath)
		if err != nil {
			log.Printf("eventstream: skip issue %s: %v", relPath, err)
			continue
		}
		events = append(events, evts...)
	}

	if len(events) == 0 {
		log.Println("eventstream: no new events to write")
		return nil
	}

	log.Printf("eventstream: writing %d events to CloudWatch Logs", len(events))
	return p.writer.Write(events)
}

// listFiles returns relative paths of all JSON files under the named subdirectory.
func (p *EventStreamProducer) listFiles(subdir string) []string {
	dir := filepath.Join(p.repoPath, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("eventstream: listFiles %s: %v", subdir, err)
		return nil
	}
	paths := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			paths = append(paths, filepath.Join(subdir, e.Name()))
		}
	}
	return paths
}

// processChangedPull returns LogEvents for a pull request file.
func (p *EventStreamProducer) processChangedPull(relPath string) ([]*cloudwatchlogs.InputLogEvent, error) {
	newBytes, err := os.ReadFile(filepath.Join(p.repoPath, relPath))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}

	var newPull types.Pull
	if err := json.Unmarshal(newBytes, &newPull); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", relPath, err)
	}

	labels := pullLabels(&newPull)
	user := newPull.Pull.User.Login

	// Build set of old event fingerprints if this is a modified (not added) file.
	oldFPs := map[string]bool{}
	isNewFile := p.prevSHA == "" || p.changedFiles == nil
	if !isNewFile {
		oldBytes, gitErr := GitShowFile(p.repoPath, p.prevSHA, relPath)
		if gitErr == nil {
			var oldPull types.Pull
			if json.Unmarshal(oldBytes, &oldPull) == nil {
				for _, e := range oldPull.Events {
					oldFPs[pullEventFP(e.ID, e.CreatedAt)] = true
				}
			}
		} else {
			// File did not exist at prevSHA → treat as new.
			isNewFile = true
		}
	}

	var events []*cloudwatchlogs.InputLogEvent

	// Lifecycle events — only emit for new files (first time we see this PR).
	if isNewFile {
		if !newPull.Pull.CreatedAt.IsZero() {
			events = append(events, p.makeEvent(LogEvent{
				SourceType: "pull", EventType: "opened",
				Number: newPull.Pull.Number, Title: newPull.Pull.Title,
				User: user, Actor: user, Labels: labels, State: newPull.Pull.State,
				EventTime: newPull.Pull.CreatedAt.UTC().Format(time.RFC3339),
			}))
		}
		if !newPull.Pull.MergedAt.IsZero() {
			events = append(events, p.makeEvent(LogEvent{
				SourceType: "pull", EventType: "merged",
				Number: newPull.Pull.Number, Title: newPull.Pull.Title,
				User: user, Actor: user, Labels: labels, State: newPull.Pull.State,
				EventTime: newPull.Pull.MergedAt.UTC().Format(time.RFC3339),
			}))
		} else if newPull.Pull.State == "closed" && !newPull.Pull.ClosedAt.IsZero() {
			events = append(events, p.makeEvent(LogEvent{
				SourceType: "pull", EventType: "closed",
				Number: newPull.Pull.Number, Title: newPull.Pull.Title,
				User: user, Actor: user, Labels: labels, State: newPull.Pull.State,
				EventTime: newPull.Pull.ClosedAt.UTC().Format(time.RFC3339),
			}))
		}
	}

	// Timeline events from the events array.
	for _, e := range newPull.Events {
		if !isNewFile && oldFPs[pullEventFP(e.ID, e.CreatedAt)] {
			continue
		}
		if !eventTypesToEmit[e.Event] {
			continue
		}
		t, ok := parseEventTime(e.CreatedAt)
		if !ok {
			continue
		}
		actor := actorLogin(e.Actor)
		events = append(events, p.makeEvent(LogEvent{
			SourceType: "pull", EventType: e.Event,
			Number: newPull.Pull.Number, Title: newPull.Pull.Title,
			User: user, Actor: actor, Labels: labels, State: newPull.Pull.State,
			EventTime: t.UTC().Format(time.RFC3339),
		}))
	}

	return p.filterByAge(events), nil
}

// processChangedIssue returns LogEvents for an issue file.
func (p *EventStreamProducer) processChangedIssue(relPath string) ([]*cloudwatchlogs.InputLogEvent, error) {
	newBytes, err := os.ReadFile(filepath.Join(p.repoPath, relPath))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}

	var newIssue types.Issue
	if err := json.Unmarshal(newBytes, &newIssue); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", relPath, err)
	}

	labels := issueLabels(&newIssue)
	user := newIssue.Issue.User.Login

	oldFPs := map[string]bool{}
	isNewFile := p.prevSHA == "" || p.changedFiles == nil
	if !isNewFile {
		oldBytes, gitErr := GitShowFile(p.repoPath, p.prevSHA, relPath)
		if gitErr == nil {
			var oldIssue types.Issue
			if json.Unmarshal(oldBytes, &oldIssue) == nil {
				for _, e := range oldIssue.Events {
					oldFPs[issueEventFP(e.ID, e.CreatedAt)] = true
				}
			}
		} else {
			isNewFile = true
		}
	}

	var events []*cloudwatchlogs.InputLogEvent

	if isNewFile {
		if !newIssue.Issue.CreatedAt.IsZero() {
			events = append(events, p.makeEvent(LogEvent{
				SourceType: "issue", EventType: "opened",
				Number: newIssue.Issue.Number, Title: newIssue.Issue.Title,
				User: user, Actor: user, Labels: labels, State: newIssue.Issue.State,
				EventTime: newIssue.Issue.CreatedAt.UTC().Format(time.RFC3339),
			}))
		}
		if newIssue.Issue.State == "closed" && !newIssue.Issue.ClosedAt.IsZero() {
			events = append(events, p.makeEvent(LogEvent{
				SourceType: "issue", EventType: "closed",
				Number: newIssue.Issue.Number, Title: newIssue.Issue.Title,
				User: user, Actor: user, Labels: labels, State: newIssue.Issue.State,
				EventTime: newIssue.Issue.ClosedAt.UTC().Format(time.RFC3339),
			}))
		}
	}

	for _, e := range newIssue.Events {
		if !isNewFile && oldFPs[issueEventFP(e.ID, e.CreatedAt)] {
			continue
		}
		if !eventTypesToEmit[e.Event] {
			continue
		}
		if e.CreatedAt.IsZero() {
			continue
		}
		events = append(events, p.makeEvent(LogEvent{
			SourceType: "issue", EventType: e.Event,
			Number: newIssue.Issue.Number, Title: newIssue.Issue.Title,
			User: user, Actor: e.Actor.Login, Labels: labels, State: newIssue.Issue.State,
			EventTime: e.CreatedAt.UTC().Format(time.RFC3339),
		}))
	}

	return p.filterByAge(events), nil
}

// makeEvent converts a LogEvent to a cloudwatchlogs.InputLogEvent.
// The CW timestamp is the event's real time (must be within 14 days — enforced by filterByAge).
func (p *EventStreamProducer) makeEvent(le LogEvent) *cloudwatchlogs.InputLogEvent {
	le.IngestedAt = p.ingestedAt
	body, _ := json.Marshal(le)
	t, _ := time.Parse(time.RFC3339, le.EventTime)
	return &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(string(body)),
		Timestamp: aws.Int64(t.UnixMilli()),
	}
}

// filterByAge drops events whose timestamp is older than the 14-day CW Logs limit.
func (p *EventStreamProducer) filterByAge(events []*cloudwatchlogs.InputLogEvent) []*cloudwatchlogs.InputLogEvent {
	out := events[:0]
	for _, e := range events {
		if !time.UnixMilli(aws.Int64Value(e.Timestamp)).Before(p.cutoff) {
			out = append(out, e)
		}
	}
	return out
}

// pullLabels returns label names for a pull request.
func pullLabels(p *types.Pull) []string {
	labels := make([]string, 0, len(p.Pull.Labels))
	for _, l := range p.Pull.Labels {
		labels = append(labels, l.Name)
	}
	return labels
}

// issueLabels returns label names for an issue.
func issueLabels(i *types.Issue) []string {
	labels := make([]string, 0, len(i.Issue.Labels))
	for _, l := range i.Issue.Labels {
		labels = append(labels, l.Name)
	}
	return labels
}

// pullEventFP returns a stable fingerprint for a PR timeline event.
// id is the events[].id field (can be int, string, or nil); createdAt is events[].created_at (any).
func pullEventFP(id any, createdAt any) string {
	if id != nil {
		return fmt.Sprintf("%v", id)
	}
	t, ok := parseEventTime(createdAt)
	if ok {
		return t.UTC().Format(time.RFC3339)
	}
	return ""
}

// issueEventFP returns a stable fingerprint for an issue timeline event.
func issueEventFP(id int, createdAt time.Time) string {
	if id != 0 {
		return fmt.Sprintf("%d", id)
	}
	return createdAt.UTC().Format(time.RFC3339)
}

// actorLogin extracts the login string from an actor field (which may be a
// map[string]any or nil in the PR timeline events).
func actorLogin(actor any) string {
	if actor == nil {
		return ""
	}
	m, ok := actor.(map[string]any)
	if !ok {
		return ""
	}
	login, _ := m["login"].(string)
	return login
}
