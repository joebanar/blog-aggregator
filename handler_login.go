package main

import (
	"context"
	"fmt"
	"os"
)

// handlerLogin sets the current user in the config file.
// Errors with exit code 1 if the user does not exist in the database.
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("usage: login <username>")
	}
	username := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		fmt.Printf("User %q does not exist\n", username)
		os.Exit(1)
	}

	if err := s.cfg.SetUser(username); err != nil {
		return fmt.Errorf("set user: %w", err)
	}
	fmt.Printf("User set to %q\n", username)
	return nil
}
