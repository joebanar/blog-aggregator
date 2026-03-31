package main

import (
	"context"
	"fmt"
	"os"

	"gator/internal/database"
)

// middlewareLoggedIn wraps a handler that requires a logged-in user,
// looking up the current user from the DB and passing them in.
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			fmt.Printf("Could not find current user: %v\n", err)
			os.Exit(1)
		}
		return handler(s, cmd, user)
	}
}
