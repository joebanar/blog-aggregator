package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"gator/internal/database"

	"github.com/google/uuid"
)

// scrapeFeeds fetches the next feed, marks it as fetched, and saves posts to the DB.
func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Printf("Error getting next feed to fetch: %v", err)
		return
	}

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Error marking feed as fetched: %v", err)
		return
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("Error fetching feed %s: %v", feed.Url, err)
		return
	}

	for _, item := range rssFeed.Channel.Item {
		publishedAt := parsePublishedAt(item.PubDate)

		_, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: sql.NullTime{Time: publishedAt, Valid: !publishedAt.IsZero()},
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Printf("Error saving post %s: %v", item.Link, err)
		}
	}
	log.Printf("Collected %d posts from %s", len(rssFeed.Channel.Item), feed.Name)
}
