package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/corecheck/corecheck/functions/compute/stats/types"
)

// cwLogsMaxAge is the default lookback window used on first run.
const cwLogsMaxAge = 24 * time.Hour

// eventTypesToEmit are the PR/issue event kinds we care about.
// Structural noise (committed, referenced, subscribed, mentioned, etc.) is excluded.
// "merged" is intentionally absent: merged PRs are captured via the pull.MergedAt lifecycle
// field, so including it here would produce a duplicate log entry.
var eventTypesToEmit = map[string]bool{
	"commented":  true,
	"reviewed":   true,
	"labeled":    true,
	"unlabeled":  true,
	"closed":     true,
	"reopened":   true,
	"assigned":   true,
	"unassigned": true,
	"locked":     true,
	"unlocked":   true,
}

// botActorsToExclude are automated accounts whose events should not be written to the log.
// Keys must be lowercase; lookups use strings.ToLower.
var botActorsToExclude = map[string]bool{
	"drahtbot": true,
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

// EventStreamProducer walks all PR/issue files and emits LogEvents for any
// event newer than the cutoff time.
type EventStreamProducer struct {
	repoPath   string
	cutoff     time.Time // only emit events strictly after this time
	writer     *CWLogsWriter
	ingestedAt string
}

// NewEventStreamProducer creates a producer. lastRunTime is the timestamp stored from the
// previous successful run; zero means first run. On first run the cutoff defaults to
// now-cwLogsMaxAge. On subsequent runs the cutoff is exactly lastRunTime so that rolling
// back the SSM parameter re-processes the same events.
func NewEventStreamProducer(repoPath string, lastRunTime time.Time, writer *CWLogsWriter) *EventStreamProducer {
	var cutoff time.Time
	if lastRunTime.IsZero() {
		cutoff = time.Now().UTC().Add(-cwLogsMaxAge)
	} else {
		cutoff = lastRunTime.UTC()
	}
	return &EventStreamProducer{
		repoPath:   repoPath,
		cutoff:     cutoff,
		writer:     writer,
		ingestedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// Run processes all PR and issue files and writes new events to CloudWatch Logs one at a time.
func (p *EventStreamProducer) Run() error {
	var written, skipped int

	writeEvent := func(le LogEvent) {
		event, err := p.makeEvent(le)
		if err != nil {
			log.Printf("eventstream: skip event (source=%s, number=%d, type=%s): %v", le.SourceType, le.Number, le.EventType, err)
			skipped++
			return
		}
		if err := p.writer.Write(event); err != nil {
			log.Printf("eventstream: failed to write event (source=%s, number=%d, type=%s): %v", le.SourceType, le.Number, le.EventType, err)
			skipped++
			return
		}
		written++
	}

	for _, relPath := range p.listFiles("pulls") {
		events, err := p.processPull(relPath)
		if err != nil {
			log.Printf("eventstream: skip pull %s: %v", relPath, err)
			continue
		}
		for _, le := range events {
			writeEvent(le)
		}
	}
	log.Printf("eventstream: processed pulls (%d written, %d skipped so far)", written, skipped)

	for _, relPath := range p.listFiles("issues") {
		events, err := p.processIssue(relPath)
		if err != nil {
			log.Printf("eventstream: skip issue %s: %v", relPath, err)
			continue
		}
		for _, le := range events {
			writeEvent(le)
		}
	}
	log.Printf("eventstream: done — %d events written, %d skipped (cutoff: %s)", written, skipped, p.cutoff.Format(time.RFC3339))

	if skipped > 0 {
		return fmt.Errorf("eventstream: %d event(s) could not be written", skipped)
	}
	return nil
}

func (p *EventStreamProducer) listFiles(subdir string) []string {
	entries, err := os.ReadDir(filepath.Join(p.repoPath, subdir))
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

func (p *EventStreamProducer) processPull(relPath string) ([]LogEvent, error) {
	data, err := os.ReadFile(filepath.Join(p.repoPath, relPath))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}
	var pull types.Pull
	if err := json.Unmarshal(data, &pull); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", relPath, err)
	}

	labels := pullLabels(&pull)
	user := pull.Pull.User.Login
	var events []LogEvent

	// Lifecycle events — emit only if their timestamp is after the cutoff.
	if !pull.Pull.CreatedAt.IsZero() && pull.Pull.CreatedAt.UTC().After(p.cutoff) {
		events = append(events, LogEvent{
			SourceType: "pull", EventType: "opened",
			Number: pull.Pull.Number, Title: pull.Pull.Title,
			User: user, Actor: user, Labels: labels, State: pull.Pull.State,
			EventTime: pull.Pull.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	if !pull.Pull.MergedAt.IsZero() && pull.Pull.MergedAt.UTC().After(p.cutoff) {
		events = append(events, LogEvent{
			SourceType: "pull", EventType: "merged",
			Number: pull.Pull.Number, Title: pull.Pull.Title,
			User: user, Actor: pull.Pull.MergedBy.Login, Labels: labels, State: pull.Pull.State,
			EventTime: pull.Pull.MergedAt.UTC().Format(time.RFC3339),
		})
	} else if pull.Pull.State == "closed" && !pull.Pull.ClosedAt.IsZero() && pull.Pull.ClosedAt.UTC().After(p.cutoff) {
		events = append(events, LogEvent{
			SourceType: "pull", EventType: "closed",
			Number: pull.Pull.Number, Title: pull.Pull.Title,
			User: user, Actor: user, Labels: labels, State: pull.Pull.State,
			EventTime: pull.Pull.ClosedAt.UTC().Format(time.RFC3339),
		})
	}

	// Timeline events.
	for _, e := range pull.Events {
		if !eventTypesToEmit[e.Event] {
			continue
		}
		// GitHub emits a spurious "closed" timeline event alongside every merge.
		// Suppress it here — the lifecycle MergedAt check above already emits "merged".
		if e.Event == "closed" && !pull.Pull.MergedAt.IsZero() {
			continue
		}
		t, ok := parseEventTime(e.CreatedAt)
		if !ok || !t.UTC().After(p.cutoff) {
			continue
		}
		actor := actorLogin(e.Actor)
		if botActorsToExclude[strings.ToLower(actor)] {
			continue
		}
		events = append(events, LogEvent{
			SourceType: "pull", EventType: e.Event,
			Number: pull.Pull.Number, Title: pull.Pull.Title,
			User: user, Actor: actorLogin(e.Actor), Labels: labels, State: pull.Pull.State,
			EventTime: t.UTC().Format(time.RFC3339),
		})
	}

	return events, nil
}

func (p *EventStreamProducer) processIssue(relPath string) ([]LogEvent, error) {
	data, err := os.ReadFile(filepath.Join(p.repoPath, relPath))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}
	var issue types.Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", relPath, err)
	}

	labels := issueLabels(&issue)
	user := issue.Issue.User.Login
	var events []LogEvent

	// Lifecycle events.
	if !issue.Issue.CreatedAt.IsZero() && issue.Issue.CreatedAt.UTC().After(p.cutoff) {
		events = append(events, LogEvent{
			SourceType: "issue", EventType: "opened",
			Number: issue.Issue.Number, Title: issue.Issue.Title,
			User: user, Actor: user, Labels: labels, State: issue.Issue.State,
			EventTime: issue.Issue.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	if issue.Issue.State == "closed" && !issue.Issue.ClosedAt.IsZero() && issue.Issue.ClosedAt.UTC().After(p.cutoff) {
		events = append(events, LogEvent{
			SourceType: "issue", EventType: "closed",
			Number: issue.Issue.Number, Title: issue.Issue.Title,
			User: user, Actor: user, Labels: labels, State: issue.Issue.State,
			EventTime: issue.Issue.ClosedAt.UTC().Format(time.RFC3339),
		})
	}

	// Timeline events.
	for _, e := range issue.Events {
		if !eventTypesToEmit[e.Event] {
			continue
		}
		if e.CreatedAt.IsZero() || !e.CreatedAt.UTC().After(p.cutoff) {
			continue
		}
		if botActorsToExclude[strings.ToLower(e.Actor.Login)] {
			continue
		}
		events = append(events, LogEvent{
			SourceType: "issue", EventType: e.Event,
			Number: issue.Issue.Number, Title: issue.Issue.Title,
			User: user, Actor: e.Actor.Login, Labels: labels, State: issue.Issue.State,
			EventTime: e.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return events, nil
}

func (p *EventStreamProducer) makeEvent(le LogEvent) (*cloudwatchlogs.InputLogEvent, error) {
	t, err := time.Parse(time.RFC3339, le.EventTime)
	if err != nil {
		return nil, fmt.Errorf("parse event time %q: %w", le.EventTime, err)
	}
	le.IngestedAt = p.ingestedAt
	body, _ := json.Marshal(le)
	log.Printf("eventstream: [%s] #%d %q — %s (actor: %s, time: %s)",
		le.SourceType, le.Number, le.Title, le.EventType, le.Actor, le.EventTime)
	return &cloudwatchlogs.InputLogEvent{
		Message:   aws.String(string(body)),
		Timestamp: aws.Int64(t.UnixMilli()),
	}, nil
}

func pullLabels(p *types.Pull) []string {
	labels := make([]string, 0, len(p.Pull.Labels))
	for _, l := range p.Pull.Labels {
		labels = append(labels, l.Name)
	}
	return labels
}

func issueLabels(i *types.Issue) []string {
	labels := make([]string, 0, len(i.Issue.Labels))
	for _, l := range i.Issue.Labels {
		labels = append(labels, l.Name)
	}
	return labels
}

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
