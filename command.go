package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/savioxavier/termlink"
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
	Cmds  map[string]func(*State, Command) error
	Helps map[string]string
}

func (cmds *Commands) Register(name, help string, f func(*State, Command) error) error {

	_, exists := cmds.Cmds[name]
	if exists {
		return fmt.Errorf("cannot Register %s, already exists", name)
	}

	cmds.Cmds[name] = f
	cmds.Helps[name] = help

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

func ScrapeFeed(s *State, user database.User) error {
	ctx := context.Background()
	feed, err := s.DbQueries.GetNextFetchedFeed(ctx, user.ID)

	if err != nil {
		return err
	}
	err = s.DbQueries.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true}, Url: feed.Url})
	if err != nil {
		return err
	}

	fmt.Printf("Checking %v feed for new posts...\n", feed.Name)
	feedResp, err := rss.FetchFeed(ctx, feed.Url)
	if err != nil {
		return err
	}

	for _, item := range feedResp.Channel.Item {

		pbDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return err
		}

		postResp, err := s.DbQueries.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   feed.CreatedAt,
			UpdatedAt:   feed.UpdatedAt,
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: pbDate,
			FeedID:      feed.FeedID})

		if err == nil {
			fmt.Printf("collecting post %v from %v\n", postResp.Title, feedResp.Channel.Title)
		} else if !strings.Contains(fmt.Sprintf("%s", err), "duplicate key value violates unique constraint") {

			return err
		}

	}

	return nil

}

//middleware funcs

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {

	return func(s *State, cmd Command) error {

		currentUser, err := s.DbQueries.GetUser(context.Background(), s.Config.Current_user_name)

		if err != nil {
			return err
		}

		return handler(s, cmd, currentUser)

	}

}

func MiddleWareHelp(handler func(s *State, c Command, cmd Commands) error, cmds Commands) func(*State, Command) error {

	return func(s *State, cmd Command) error {
		return HandlerHelp(s, cmd, cmds)
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

	err = s.DbQueries.ResetPosts(ctx)
	if err != nil {
		return nil
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

func HandlerBrowse(s *State, cmd Command, user database.User) error {
	limit := 2

	if len(cmd.Args) > 0 {
		limit, _ = strconv.Atoi(cmd.Args[0])

	}
	userPostResp, err := s.DbQueries.GetUserPosts(context.Background(), database.GetUserPostsParams{UserID: user.ID, Limit: int32(limit)})
	if err != nil {
		return err
	}

	for num, post := range userPostResp {
		rss, err := s.DbQueries.GetFeed(context.Background(), post.FeedID)

		if err != nil {
			return err
		}

		fmt.Printf("post #%v: from %s: %v\n", num, rss.Name, post.CreatedAt)
		fmt.Printf("\tTitle : %v\n", termlink.Link(post.Title, post.Url))
		fmt.Printf("\t\t%v\n\n", post.Description)
	}

	return nil
}

func HandlerAggregate(s *State, cmd Command, user database.User) error {

	if len(cmd.Args) < 1 {
		return fmt.Errorf("aggregate commend expect time args eg: <1m>")
	}

	t := cmd.Args[0]
	reqDuration, err := time.ParseDuration(t)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(reqDuration)

	fmt.Printf("Collecting feeds every %v\n", t)

	for ; ; <-ticker.C {

		err = ScrapeFeed(s, user)

		if err != nil {
			return err
		}

	}

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

func HandlerHelp(s *State, cmd Command, cmds Commands) error {
	fmt.Println("commands for tool:")
	for coms, help := range cmds.Helps {
		fmt.Printf("\t %v: %v\n", coms, help)
	}
	return nil

}

func HandlerRemoveFeed(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("remove feed expect a feed name arg")
	}

	err := s.DbQueries.RemoveFeeds(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Removed %v from rss feeds\n", cmd.Args[0])

	return nil
}
