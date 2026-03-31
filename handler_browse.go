package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"gator/internal/database"
)

// handlerBrowse prints the most recent posts for the current user.
func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) > 0 {
		var err error
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit %q: %w", cmd.args[0], err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		fmt.Printf("Failed to get posts: %v\n", err)
		os.Exit(1)
	}

	for _, post := range posts {
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("URL:      %s\n", post.Url)
		if post.Description.Valid {
			fmt.Printf("Desc:     %s\n", post.Description.String)
		}
		if post.PublishedAt.Valid {
			fmt.Printf("Published: %s\n", post.PublishedAt.Time.Format(time.RFC1123))
		}
		fmt.Println()
	}
	return nil
}
