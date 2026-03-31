# gator

gator is a small CLI RSS aggregator that stores feeds and posts in Postgres.

Prerequisites
- Go (1.21 or later recommended)
- Postgres (a running Postgres instance and a database you can connect to)

Install

From this repository (local install):

```bash
# build and install the `gator` binary into $GOBIN (or $GOPATH/bin)
go install .
```

If you host this repository on GitHub (or another VCS), you can also install directly using the module path:

```bash
# replace with the proper repo path, e.g. github.com/you/blog-aggregator
go install github.com/<your-username>/<repo>@latest
```

Configuration

gator reads its config from a JSON file at `~/.gatorconfig.json`.
Create that file with at least a `db_url` field. Example:

```json
{
  "db_url": "postgres://user:password@localhost:5432/gator_db?sslmode=disable",
  "current_user_name": ""
}
```

Example to create the config quickly:

```bash
cat > ~/.gatorconfig.json <<'EOF'
{
  "db_url": "postgres://user:password@localhost:5432/gator_db?sslmode=disable"
}
EOF
```

Running

Once installed, run the CLI as `gator <command> [args]`.
You can also run without installing using:

```bash
go run . -- <command> [args]
```

Common commands
- `register <name>` — create a new user and set them as the current user
- `login <username>` — set the current user (must already exist)
- `addfeed <name> <url>` — add an RSS feed (must be logged in)
- `feeds` — list all feeds
- `follow <url>` — follow a feed by URL (must be logged in)
- `following` — list feeds the current user follows
- `browse [limit]` — show the most recent posts for the current user (default limit 2)
- `agg <interval>` — run the aggregator loop (e.g., `agg 1m` to poll every minute)
- `users` — list users
- `reset` — delete all users (use with caution)

Notes
- The `db_url` should point to a reachable Postgres instance. If you need to run migrations, check the `sql/schema` and `sql/queries` directories; apply them to your database before using the app.
- The config file path and format are implemented in `internal/config/config.go`.

If you'd like, I can add a small `Makefile` for common tasks, add example migrations, or commit these changes for you.
