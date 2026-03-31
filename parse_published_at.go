package main

import "time"

// parsePublishedAt attempts to parse a published date string from an RSS feed.
func parsePublishedAt(s string) time.Time {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}
