package main

import (
	"context"
	"fmt"
	"os"
)

// handlerFeeds prints all feeds in the database with their creator's name.
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		fmt.Printf("Failed to get feeds: %v\n", err)
		os.Exit(1)
	}

	for _, feed := range feeds {
		fmt.Printf("* %s\n", feed.Name)
		fmt.Printf("  URL:  %s\n", feed.Url)
		fmt.Printf("  User: %s\n", feed.UserName)
	}
	return nil
}
