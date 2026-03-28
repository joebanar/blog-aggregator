package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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

// RSSFeed represents the top-level RSS feed structure.
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

// RSSItem represents a single item/post in an RSS feed.
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// fetchFeed fetches and parses an RSS feed from the given URL.
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var feed RSSFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("unmarshal xml: %w", err)
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, item := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(item.Title)
		feed.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}

	return &feed, nil
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

	fmt.Println("User created successfully!")
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

// handlerAgg fetches and prints the wagslane RSS feed.
func handlerAgg(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("fetch feed: %w", err)
	}
	fmt.Printf("%+v\n", feed)
	return nil
}

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

// handlerFollow creates a feed follow record for the current user by feed URL.
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("usage: follow <url>")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		fmt.Printf("Could not find feed with URL %q: %v\n", url, err)
		os.Exit(1)
	}

	now := time.Now()
	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		fmt.Printf("Failed to follow feed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User %q is now following %q\n", follow.UserName, follow.FeedName)
	return nil
}

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
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))

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
