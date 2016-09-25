package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type fileConfig struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	AccessToken    string `json:"access_token"`
	AccessSecret   string `json:"access_secret"`
}

var Twitter *fileConfig

func init() {

	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("\x1b[31m[Config file]\x1b[39m", err.Error())
		os.Exit(1)
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&Twitter); err != nil {
		fmt.Println("\x1b[31m", "Parsing config file error: ", err.Error(), "\x1b[39m")
		os.Exit(1)
	}
}
