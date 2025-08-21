package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/shahanmmiah/blogAggregator/internal/config"
	"github.com/shahanmmiah/blogAggregator/internal/database"

	_ "github.com/lib/pq"
)

func main() {

	c, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}

	db, err := sql.Open("postgres", c.Db_url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	state := State{Config: &c, DbQueries: dbQueries}
	cmds := Commands{Cmds: map[string]func(*State, Command) error{}, Helps: make(map[string]string)}

	cmds.Register("login", "login to an existing username", HandlerLogin)
	cmds.Register("register", "register a new username", HandlerRegister)
	cmds.Register("reset", "reset all database data", HandlerReset)
	cmds.Register("users", "view all existing users", HandlerGetUsers)
	cmds.Register("agg", "collecting post from user followed feeds", MiddlewareLoggedIn(HandlerAggregate))
	cmds.Register("addfeed", "add a neww rss feed", MiddlewareLoggedIn(HandlerAddFeed))
	cmds.Register("feeds", "view all rss feeds", HandlerFeeds)
	cmds.Register("follow", "follow a rss feed", MiddlewareLoggedIn(HandlerFollow))
	cmds.Register("unfollow", "unfollow an rss feed", MiddlewareLoggedIn(HandlerUnfollow))
	cmds.Register("following", "veiw all followed feeds", MiddlewareLoggedIn(HandlerFollowing))
	cmds.Register("browse", "browse all collected posts", MiddlewareLoggedIn(HandlerBrowse))
	cmds.Register("help", "view help message for commands", MiddleWareHelp(HandlerHelp, cmds))
	cmds.Register("removefeed", "remove a feed from saved feeds", HandlerRemoveFeed)

	cmd, err := CreateCommand()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = cmds.Run(&state, cmd)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
