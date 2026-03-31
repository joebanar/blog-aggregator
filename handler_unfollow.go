package main

import (
	"context"
	"fmt"
	"os"

	"gator/internal/database"
)

// handlerUnfollow removes a feed follow record for the current user by feed URL.
func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("usage: unfollow <url>")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		fmt.Printf("Could not find feed with URL %q: %v\n", url, err)
		os.Exit(1)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		fmt.Printf("Failed to unfollow feed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User %q unfollowed %q\n", user.Name, feed.Name)
	return nil
}
