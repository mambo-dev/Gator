package main

import (
	"database/sql"
	"errors"
	"fmt"
	"gator/internal"
	"gator/internal/database"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type state struct {
	config *internal.Config
	db     *database.Queries
}

type command struct {
	name      string
	arguments []string
}

func main() {

	config := internal.ReadGatorConfig()

	dbUrl := config.DbUrl

	db, err := sql.Open("postgres", dbUrl)

	if err != nil {
		log.Fatal(err.Error())
	}

	dbQueries := database.New(db)

	newState := state{
		config: config,
		db:     dbQueries,
	}

	commands := commands{
		commands: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)

	args := os.Args

	if len(args) < 2 {
		log.Fatal("expecting more arguments")
	}

	commandName := args[1]
	commandArguments := args[2:]
	command := command{
		name:      commandName,
		arguments: commandArguments,
	}

	err := commands.run(&newState, command)

	if err != nil {
		log.Fatal(err.Error())
	}

}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("login handler expects a single argument, the username.")
	}

	config := s.config

	config.SetUser(cmd.arguments[0])

	fmt.Printf("username: %v successfully set\n", config.CurrentUserName)

	return nil
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f

}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.commands[cmd.name]

	if !ok {
		return errors.New("no such command")
	}

	err := handler(s, cmd)

	if err != nil {
		return errors.New(err.Error())
	}

	return nil
}
