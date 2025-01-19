package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gator/internal"
	"gator/internal/database"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
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
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)

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

	err = commands.run(&newState, command)

	if err != nil {
		log.Fatal(err.Error())
	}

}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("login handler expects a single argument, the username.")
	}

	config := s.config

	dbQuery := s.db

	user, err := dbQuery.GetUser(context.Background(), cmd.arguments[0])

	if err != nil {
		log.Fatalf("User does not exist\n - error:%v\n", err.Error())
	}

	config.SetUser(user.Name)

	fmt.Printf("username: %v successfully set\n", config.CurrentUserName)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("register handler expects a single argument, the username.")
	}

	config := s.config
	dbQuery := s.db

	newUser := database.CreateUserParams{
		ID:        uuid.New(),
		Name:      cmd.arguments[0],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user, err := dbQuery.CreateUser(context.Background(), newUser)

	if err != nil {
		log.Fatal("User already exists")
	}

	config.SetUser(user.Name)
	fmt.Println("User was created")
	fmt.Printf("username: %v successfully set\n", config.CurrentUserName)

	return nil
}

func handlerReset(s *state, cmd command) error {
	dbQuery := s.db

	err := dbQuery.DeleteUsers(context.Background())

	if err != nil {
		log.Fatalf("Failed to delete:\n - error: %v", err.Error())
	}

	log.Println("User table succesfully reset !")
	return nil
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {

	_, ok := c.commands[name]

	if ok {
		log.Fatal("command already registered")
	}

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
