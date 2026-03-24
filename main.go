package main

import (
	"context"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"log"
	"os"
	"time"

	"database/sql"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// state holds shared application state passed to command handlers.
type state struct {
	db  *database.Queries
	cfg *config.Config
}

// command represents a CLI command with its name and arguments.
type command struct {
	name string
	args []string
}

// commands holds a registry of named command handlers.
type commands struct {
	handlers map[string]func(*state, command) error
}

// register adds a new handler function for the given command name.
func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

// run looks up and executes the handler for the given command.
// Returns an error if the command is not registered.
func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

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

	fmt.Printf("User created successfully!\n")
	log.Printf("DEBUG user: %+v\n", user)
	return nil
}

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

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("read config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	dbQueries := database.New(db)

	s := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cmds := &commands{
		handlers: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)

	if len(os.Args) < 2 {
		log.Fatal("usage: gator <command> [args...]")
	}

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	if err := cmds.run(s, cmd); err != nil {
		log.Fatalf("error: %v", err)
	}
}

// handlerUsers prints all users, marking the currently logged-in user.
func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Failed to get users: %v\n", err)
		os.Exit(1)
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}
