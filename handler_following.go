package main

import (
	"context"
	"fmt"
	"os"

	"gator/internal/database"
)

// handlerFollowing prints all feeds the current user is following.
func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Printf("Failed to get follows: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Feeds followed by %q:\n", user.Name)
	for _, follow := range follows {
		fmt.Printf("* %s\n", follow.FeedName)
	}
	return nil
}
