package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/mambo-dev/Gator/internal"
	"github.com/mambo-dev/Gator/internal/database"
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
	commands.register("agg", middlewareLoggedIn(handlerAgg))
	commands.register("addfeed", middlewareLoggedIn(handlerFeed))
	commands.register("feeds", handlerFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commands.register("browse", handlerBrowse)

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
		log.Fatalf("Failed to get users:\n - error: %v", err.Error())
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

func handlerAgg(s *state, cmd command, user database.User) error {

	if len(cmd.arguments) < 1 {
		return errors.New("expecting one arguments")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.arguments[0])

	if err != nil {
		log.Fatal("could not parse command")
	}

	ticker := time.NewTicker(timeBetweenRequests)
	fmt.Printf("Collecting feeds every: %v \n", timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
		fmt.Println("Collecting feeds...")
	}

}

func handlerFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return errors.New("expecting two arguments")
	}

	dbQuery := s.db

	name := cmd.arguments[0]
	url := cmd.arguments[1]

	newFeed := database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      name,
		Url:       url,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	}

	createdFeed, err := dbQuery.CreateFeed(context.Background(), newFeed)

	if err != nil {
		return errors.New("could not create feed")
	}

	_, err = dbQuery.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		FeedID: uuid.NullUUID{
			UUID:  createdFeed.ID,
			Valid: true,
		},
		UserID:    createdFeed.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		return errors.New("failed to create user's feed follow")
	}

	fmt.Println(createdFeed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	dbQuery := s.db

	feeds, err := dbQuery.GetFeeds(context.Background())

	if err != nil {
		return errors.New("could not get feeds from db")
	}

	fmt.Println("Your feeds are:")

	for _, feed := range feeds {
		fmt.Printf("- name: %v\n - url: %v\n  - created by: %v\n", feed.FeedName, feed.Url, feed.UserName)
	}

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	dbQuery := s.db

	feedFollows, err := dbQuery.GetFeedFollowsForUser(context.Background(), uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	})

	if err != nil {
		return errors.New("could not get feed follows")
	}

	fmt.Printf("%v's feeds: \n", user.Name)

	for _, feed := range feedFollows {
		fmt.Printf("Name: %v \n", feed.FeedName)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {

	if len(cmd.arguments) < 1 {
		return errors.New("expecting an argument.")
	}

	dbQuery := s.db

	feed, err := dbQuery.GetFeed(context.Background(), cmd.arguments[0])

	if err != nil {
		return errors.New("could not get specified feed.")
	}

	feeds, err := dbQuery.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		FeedID: uuid.NullUUID{
			UUID:  feed.ID,
			Valid: true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		return errors.New("could not get feeds from db")
	}

	fmt.Printf("%v feed follows has feed %v\n", feeds.UserName, feeds.FeedName)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {

	if len(cmd.arguments) < 1 {
		return errors.New("expecting an argument.")
	}

	dbQuery := s.db

	err := dbQuery.DeleteFeedFollowForUser(context.Background(), database.DeleteFeedFollowForUserParams{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Url: cmd.arguments[0],
	})

	if err != nil {
		return fmt.Errorf("could not delete feed %v\n", err)
	}

	return nil
}

func scrapeFeeds(s *state) bool {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())

	if err != nil {
		log.Fatalf("could not get next feed to fetch %v\n", err.Error())
	}

	err = s.db.MarkFeedFetched(context.Background(), nextFeed.ID)

	if err != nil {
		log.Fatalf("Failed to mark feed as fetched %v\n", err.Error())
	}

	feeds, err := fetchFeed(context.Background(), nextFeed.Url)

	if err != nil {
		log.Fatalf("could not fetch feed from url %v\n", err.Error())
	}

	for _, feed := range feeds.Channel.Item {

		fmt.Println(feed.Title)

		pubDate, err := time.Parse("2006-01-02", feed.PubDate)

		if err != nil {
			pubDate = time.Now()
		}

		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title:     feed.Title,
			Url:       feed.Link,
			Description: sql.NullString{
				String: feed.Description,
			},
			PublishedAt: sql.NullTime{
				Time:  pubDate,
				Valid: true,
			},
			FeedID: nextFeed.ID,
		})

		if err != nil {
			fmt.Printf("Failed to create post %v\n", err.Error())
			continue
		}

	}

	return true
}

func handlerBrowse(s *state, cmd command) error {
	limit := 2
	if len(cmd.arguments) >= 1 {
		arg, err := strconv.Atoi(cmd.arguments[0])

		if err != nil {
			log.Fatalf("Invalid limit argument passed %v\n", err.Error())
		}

		limit = arg
	}

	fmt.Println(limit)
	posts, err := s.db.GetPostsForUser(context.Background(), int32(limit))

	if err != nil {

		return errors.New(fmt.Sprintf("could not get posts %v\n", err.Error()))
	}

	for _, post := range posts {
		fmt.Printf("Title: %v\nDescription: %v\nPublished at:%v\n", post.Title, post.Description.String, post.PublishedAt.Time.UTC().Format("2006-01-02"))
	}

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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	return func(s *state, c command) error {

		dbQuery := s.db

		user, err := dbQuery.GetUser(context.Background(), s.config.CurrentUserName)

		if err != nil {
			return errors.New("could not get logged in user.")
		}

		return handler(s, c, user)

	}
}
