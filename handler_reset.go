package main

import (
	"context"
	"fmt"
	"os"
)

// handlerReset deletes all users from the database.
func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		fmt.Printf("Failed to reset database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database reset successfully.")
	return nil
}
