package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const CONSTFILENAME = ".gatorconfig.json"

type Config struct {
	Db_url            string
	Current_user_name string
}

type State struct {
	Config_ptr *Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Cmds map[string]func(*State, Command) error
}

func Read() (Config, error) {
	user, _ := os.UserHomeDir()
	fileData, err := os.ReadFile(fmt.Sprintf("%v/%v", user, CONSTFILENAME))
	config := Config{}

	if err != nil {
		return Config{}, fmt.Errorf("could not read config data")
	}

	err = json.Unmarshal(fileData, &config)
	if err != nil {
		return Config{}, fmt.Errorf("could not read config JSON data %v", err)
	}

	return config, nil

}

func (cfg *Config) SetUser(username string) error {

	cfg.Current_user_name = username
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return (err)
	}
	user, _ := os.UserHomeDir()
	filePath := fmt.Sprintf("%v/%v", user, CONSTFILENAME)
	os.WriteFile(filePath, jsonData, 0644)

	return nil
}

func HandlerLogin(s *State, cmd Command) error {

	if len(cmd.Args) == 0 {
		return fmt.Errorf("login command expects a username argument")
	}

	err := s.Config_ptr.SetUser(cmd.Args[0])

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
		return fmt.Errorf("cannot Register %v, already exists", name)
	}

	cmds.Cmds[name] = f
	return nil
}
