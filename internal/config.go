package internal

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(name string) {
	dirPath, err := getConfigFilePath()

	if err != nil {
		log.Fatal(err)
	}

	c.CurrentUserName = name

	dat, err := json.Marshal(c)

	if err != nil {
		log.Fatal("could not change json data")
	}

	os.WriteFile(dirPath, dat, 0644)
}

func ReadGatorConfig() *Config {
	config := &Config{}

	dirPath, err := getConfigFilePath()

	if err != nil {
		log.Fatal(err)
	}

	dat, err := os.ReadFile(dirPath)

	if err != nil {
		log.Fatal(err)
		log.Fatal("could not read file")
	}

	json.Unmarshal(dat, config)

	return config
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {

		return "", errors.New("could get user home directory")
	}

	dirPath := strings.Join([]string{homeDir, configFileName}, "/")

	return dirPath, nil
}
