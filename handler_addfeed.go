package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gator/internal/database"

	"github.com/google/uuid"
)

// handlerAddFeed adds a new RSS feed to the database for the current user,
// and automatically creates a follow record for them.
func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}
	name := cmd.args[0]
	url := cmd.args[1]

	now := time.Now()
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		fmt.Printf("Failed to create feed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Feed created successfully!")
	log.Printf("DEBUG feed: %+v\n", feed)

	now = time.Now()
	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		fmt.Printf("Feed created but failed to follow: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Now following %q\n", follow.FeedName)
	return nil
}
