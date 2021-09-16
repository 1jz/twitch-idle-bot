package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"twitch-idle/clientmanager"
	"twitch-idle/utils"

	"github.com/asaskevich/EventBus"
)

// Socket connection consts
const (
	HOST = "irc.chat.twitch.tv"
	PORT = "6667"
)

func main() {
	var wg sync.WaitGroup
	config := loadConfig()
	go captureTerm(&config)

	bus := EventBus.New()

	botManager := clientmanager.Setup(&bus, &config, HOST, PORT)
	err := bus.Subscribe("manager:joined_channel", botManager.JoinedChannel)
	if err != nil {
		fmt.Println(err)
	}
	bus.Subscribe("manager:received_data", botManager.ReceivedData)
	botManager.Start()

	wg.Add(1)
	wg.Wait()
}

// config loading/saving stuff
func captureTerm(config *utils.Config) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig

	fmt.Println()
	log.Println("Program killed!")
	saveConfig(*config)
	os.Exit(0)
}

func loadConfig() utils.Config {
	config, err := utils.GetConfig()
	if err != nil {
		log.Println("Error:", err.Error())
		os.Exit(1)
	}

	return config
}

func saveConfig(config utils.Config) {
	file, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		log.Println("Error:", err.Error())
		os.Exit(1)
	}

	err = ioutil.WriteFile("config.json", file, 0644)
	if err != nil {
		log.Println("Error:", err.Error())
		os.Exit(1)
	}
}
