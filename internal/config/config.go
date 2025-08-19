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
	Posts_limit       int32
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

func (cfg *Config) SaveConfig() error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return (err)
	}
	user, _ := os.UserHomeDir()
	filePath := fmt.Sprintf("%v/%v", user, CONSTFILENAME)
	os.WriteFile(filePath, jsonData, 0644)

	return nil
}

func (cfg *Config) SetUser(username string) error {

	cfg.Current_user_name = username
	err := cfg.SaveConfig()
	if err != nil {
		return err
	}

	return nil
}

func (cfg *Config) SetPostLimit(amount int32) error {
	cfg.Posts_limit = amount
	err := cfg.SaveConfig()
	if err != nil {
		return err
	}

	return nil

}
