package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/shahanmmiah/blogAggregator/internal/config"
	"github.com/shahanmmiah/blogAggregator/internal/database"

	_ "github.com/lib/pq"
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

func HandlerReset(s *State, cmd Command) error {
	err := s.DbQueries.Reset(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Database has been reset")
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
