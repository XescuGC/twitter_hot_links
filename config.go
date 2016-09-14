package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	AccessToken    string `json:"access_token"`
	AccessSecret   string `json:"access_secret"`
}

type Opts struct {
	configFile string
}

func readConfig() Config {
	var c Config

	getopts := getOpts()

	configFile, err := os.Open(getopts.configFile)
	if err != nil {
		fmt.Println("\x1b[31m[Config file]\x1b[39m", err.Error())
		os.Exit(1)
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&c); err != nil {
		fmt.Println("\x1b[31m", "Parsing config file error: ", err.Error(), "\x1b[39m")
		os.Exit(1)
	}

	return c
}

func getOpts() Opts {
	configFile := flag.String("config-file", "config.json", "Config file to get twitter credentials")
	flag.Parse()
	return Opts{*configFile}
}
