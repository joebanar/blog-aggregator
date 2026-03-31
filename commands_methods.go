package main

import "fmt"

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
