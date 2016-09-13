package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	AccessToken    string `json:"access_token"`
	AccessSecret   string `json:"access_secret"`
}

func readConfig() Config {
	var c Config

	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Opening config file", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&c); err != nil {
		fmt.Println("Parsing config file", err.Error())
	}

	return c
}
