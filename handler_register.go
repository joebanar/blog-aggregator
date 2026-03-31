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

// handlerRegister creates a new user in the database and sets them as current user.
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("usage: register <name>")
	}
	name := cmd.args[0]

	now := time.Now()
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
	})
	if err != nil {
		fmt.Printf("User %q already exists\n", name)
		os.Exit(1)
	}

	if err := s.cfg.SetUser(name); err != nil {
		return fmt.Errorf("set user: %w", err)
	}

	fmt.Println("User created successfully!")
	log.Printf("DEBUG user: %+v\n", user)
	return nil
}
