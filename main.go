package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"gator/internal"
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
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerFeed)

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

func handlerUsers(s *state, cmd command) error {
	dbQuery := s.db

	users, err := dbQuery.GetUsers(context.Background())

	if err != nil {
		log.Fatalf("Failed to delete:\n - error: %v", err.Error())
	}

	currentlyLoggedInUser := s.config.CurrentUserName

	for _, user := range users {

		if user.Name == currentlyLoggedInUser {
			fmt.Printf("* %v (current) \n", user.Name)

		} else {
			fmt.Printf("* %v \n", user.Name)
		}

	}

	return nil

}

func handlerAgg(s *state, cmd command) error {

	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")

	if err != nil {
		return errors.New("could not read feed")
	}

	fmt.Println(feed)

	return nil
}

func handlerFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 2 {
		return errors.New("expecting two arguments")
	}

	dbQuery := s.db

	currentUser, err := dbQuery.GetUser(context.Background(), s.config.CurrentUserName)

	if err != nil {
		return errors.New("user does not exist")
	}

	name := cmd.arguments[0]
	url := cmd.arguments[1]

	newFeed := database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      name,
		Url:       url,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: uuid.NullUUID{
			UUID:  currentUser.ID,
			Valid: true,
		},
	}

	createdFeed, err := dbQuery.CreateFeed(context.Background(), newFeed)

	fmt.Println(createdFeed)
	return nil
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)

	request.Header.Set("User-Agent", "gator")

	if err != nil {
		log.Fatalf("Error fetching feed: %v\n", err.Error())
	}

	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		log.Fatalf("Error returning response: %v\n", err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("Error reading body: %v\n", err.Error())
	}

	if resp.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, body)
	}

	rssFeed := &RSSFeed{}

	err = xml.Unmarshal(body, rssFeed)

	if err != nil {
		log.Fatalf("Failed to unmarshal xml: %v\n", err.Error())
	}

	html.UnescapeString(rssFeed.Channel.Title)
	html.UnescapeString(rssFeed.Channel.Description)

	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)

	return rssFeed, nil

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
