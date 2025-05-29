package config

import (
	"encoding/json"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Db_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func getConfigPath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir = dir + "/" + configFileName
	return dir, nil
}

func Read() (Config, error) {
	dir, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}
	dat, err := os.ReadFile(dir)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(dat, &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) SetUser(username string) error {
	c.Current_user_name = username
	jsonData, err := json.Marshal(*c)
	if err != nil {
		return err
	}
	dir, err := getConfigPath()
	if err != nil {
		return err
	}
	os.WriteFile(dir, jsonData, 0666)
	return nil
}
