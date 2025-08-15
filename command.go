package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/shahanmmiah/blogAggregator/internal/config"
	"github.com/shahanmmiah/blogAggregator/internal/database"
	"github.com/shahanmmiah/blogAggregator/rss"
)

type State struct {
	Config    *config.Config
	DbQueries *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Cmds map[string]func(*State, Command) error
}

func (cmds *Commands) Register(name string, f func(*State, Command) error) error {

	_, exists := cmds.Cmds[name]
	if exists {
		return fmt.Errorf("cannot Register %s, already exists", name)
	}

	cmds.Cmds[name] = f
	return nil
}

func CreateCommand() (Command, error) {

	inputArgs := os.Args
	if len(inputArgs) < 2 {
		return Command{}, fmt.Errorf("error: No command argument specified")

	}

	cmd := Command{Name: inputArgs[1], Args: inputArgs[2:]}
	return cmd, nil
}

func (cmds *Commands) Run(s *State, cmd Command) error {

	funcName, exists := cmds.Cmds[cmd.Name]
	if !exists {
		return fmt.Errorf("command does not exists")
	}
	err := funcName(s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {

	return func(s *State, cmd Command) error {

		currentUser, err := s.DbQueries.GetUser(context.Background(), s.Config.Current_user_name)

		if err != nil {
			return err
		}

		return handler(s, cmd, currentUser)

	}

}

// cli handler functions
func HandlerReset(s *State, cmd Command) error {
	ctx := context.Background()
	err := s.DbQueries.ResetUsers(ctx)
	if err != nil {
		return err
	}

	err = s.DbQueries.ResetFeeds(ctx)
	if err != nil {
		return err
	}

	err = s.DbQueries.ResetFeedFollow(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Databases has been reset")
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("Register command expects a user argument")
	}

	_, err := s.DbQueries.GetUser(
		context.Background(),
		cmd.Args[0])

	if err == nil {
		return fmt.Errorf("error: user %s already exists", cmd.Args[0])
	}

	newUser, err := s.DbQueries.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.Args[0]},
	)
	if err != nil {
		return err
	}

	err = s.Config.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("User %v has been registered: %v\n", newUser.Name, newUser)
	return nil

}

func HandlerLogin(s *State, cmd Command) error {

	if len(cmd.Args) == 0 {
		return fmt.Errorf("login command expects a username argument")
	}

	_, err := s.DbQueries.GetUser(
		context.Background(),
		cmd.Args[0])

	if err != nil {
		return fmt.Errorf("error: user %v doesnt exists", cmd.Args[0])
	}

	err = s.Config.SetUser(cmd.Args[0])

	if err != nil {
		return err
	}

	fmt.Printf("User has been set to: %v\n", cmd.Args[0])
	return nil
}

func HandlerGetUsers(s *State, cmd Command) error {

	names, err := s.DbQueries.GetUsers(context.Background())

	if err != nil {
		return err
	}

	for _, name := range names {

		if name == s.Config.Current_user_name {
			fmt.Printf("* %s (current)\n", name)
		} else {
			fmt.Printf("* %s\n", name)
		}

	}
	return nil

}

func HandlerAggegate(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("agg command expects a rss URL")

	}

	feed, err := rss.FetchFeed(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Println(feed)

	return nil
}

func HandlerAddFeed(s *State, cmd Command, currentUser database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("addfeed command expects a name and URL")

	}
	ctx := context.Background()

	feedEntry, err := s.DbQueries.CreateFeed(ctx,
		database.CreateFeedParams{
			Name:   cmd.Args[0],
			Url:    cmd.Args[1],
			UserID: currentUser.ID,
		})

	if err != nil {
		return err
	}

	fmt.Printf("feed entry added: %v\n", feedEntry)

	followArgs := Command{Name: "follow", Args: cmd.Args[1:]}
	err = HandlerFollow(s, followArgs, currentUser)

	if err != nil {
		return err
	}

	return nil

}

func HandlerFeeds(s *State, cmd Command) error {
	ctx := context.Background()
	feeds, err := s.DbQueries.GetFeeds(ctx)

	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("feed name: %v\n", feed.Name)
		fmt.Printf("feed url: %v\n", feed.Url)
		user, err := s.DbQueries.GetUserFromId(ctx, feed.UserID)

		if err != nil {
			return nil
		}
		fmt.Printf("user: %v\n", user.Name)
		fmt.Println("***********************")

	}

	return nil
}

func HandlerFollow(s *State, cmd Command, currentUser database.User) error {
	ctx := context.Background()
	if len(cmd.Args) < 1 {
		return fmt.Errorf("follow command expects a url arg")
	}

	followRes, err := s.DbQueries.CreateFeedFollow(
		ctx,
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    currentUser.ID,
			FeedID:    cmd.Args[0]})

	if err != nil {
		return err
	}

	fmt.Printf("user %v has followed feed: %v\n", followRes.UserName, followRes.FeedName)
	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {

	ctx := context.Background()
	if len(cmd.Args) < 1 {
		return fmt.Errorf("unfollow command expects a url arg")
	}

	err := s.DbQueries.RemoveFeedFollow(
		ctx,
		database.RemoveFeedFollowParams{
			UserID: user.ID,
			FeedID: cmd.Args[0],
		})
	if err != nil {
		return err
	}

	feed, err := s.DbQueries.GetFeed(ctx, cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("%v has unfollowed %v", user.Name, feed.Name)
	return nil

}
func HandlerFollowing(s *State, cmd Command, user database.User) error {

	ctx := context.Background()

	res, err := s.DbQueries.GetFeedsForUser(ctx, user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("%v is following:\n", user.Name)
	for _, feed := range res {
		fmt.Printf("	%v\n", feed.FeedName)
	}

	return nil

}
