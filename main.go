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
	cmds := Commands{Cmds: map[string]func(*State, Command) error{}}

	cmds.Register("login", HandlerLogin)
	cmds.Register("register", HandlerRegister)
	cmds.Register("reset", HandlerReset)
	cmds.Register("users", HandlerGetUsers)
	cmds.Register("agg", MiddlewareLoggedIn(HandlerAggregate))
	cmds.Register("addfeed", MiddlewareLoggedIn(HandlerAddFeed))
	cmds.Register("feeds", HandlerFeeds)
	cmds.Register("follow", MiddlewareLoggedIn(HandlerFollow))
	cmds.Register("unfollow", MiddlewareLoggedIn(HandlerUnfollow))
	cmds.Register("following", MiddlewareLoggedIn(HandlerFollowing))
	cmds.Register("browse", MiddlewareLoggedIn(HandlerBrowse))

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
