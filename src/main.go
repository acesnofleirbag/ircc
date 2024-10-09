package main

import (
	"encoding/json"
	"fmt"
	"io"
	"ircc/src/guard"
	"os"
)

var (
	CLIENT Client
	IFACE  UI
)

type Config struct {
	Port     int    `json:"port"`
	Address  string `json:"address"`
	Server   string `json:"server"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func main() {
	home := os.Getenv("HOME")

	configFile, err := os.Open(fmt.Sprintf("%v/.config/ircc", home))
	guard.Err(err)
	defer configFile.Close()

	configData, err := io.ReadAll(configFile)
	guard.Err(err)

	var config Config

	err = json.Unmarshal(configData, &config)
	guard.Err(err)

	CLIENT = NewClient(config.Address, config.Port, config.Nickname)
	CLIENT.Run(&config)

	IFACE = NewScreen()
	IFACE.Run()
}
