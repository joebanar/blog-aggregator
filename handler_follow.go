package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"gator/internal/database"

	"github.com/google/uuid"
)

// handlerFollow creates a feed follow record for the current user by feed URL.
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("usage: follow <url>")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		fmt.Printf("Could not find feed with URL %q: %v\n", url, err)
		os.Exit(1)
	}

	now := time.Now()
	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		fmt.Printf("Failed to follow feed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User %q is now following %q\n", follow.UserName, follow.FeedName)
	return nil
}
