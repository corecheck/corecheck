package main

import "time"

// parseEventTime converts a timeline event timestamp (typed as any because some
// event kinds omit it) into a time.Time. Returns false when the value is absent
// or unparseable.
func parseEventTime(v any) (time.Time, bool) {
	if v == nil {
		return time.Time{}, false
	}
	s, ok := v.(string)
	if !ok {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	return t, err == nil
}
