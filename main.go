package main

import (
	"chat-idle/apic"
	"chat-idle/irc"
	"chat-idle/utils"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Socket connection consts
const (
	HOST = "irc.chat.twitch.tv"
	PORT = "6667"
)

func main() {
	var wg sync.WaitGroup
	config := loadConfig()

	var bot *irc.IRClient = irc.Create(&config)
	go captureTerm(bot, &config)

	if config.LogLevel >= 1 {
		log.Printf("Joining %d channels as %s...\n", len(config.Channels), config.User)
		log.Printf("Starting in %d Milliseconds...\n", config.StartDelay)
		log.Printf("App Token expires: %s", time.Unix(config.AppToken.Issued+config.AppToken.ExpiresIn, 0))
	}

	time.Sleep(time.Second * time.Duration(config.StartDelay/1000))

	bot.Connect(HOST, PORT)

	apic.QueueChannels(&config)
	if config.JoinAllLiveChannels {
		go apic.ScanChannels(&config, bot)
		go apic.JoinChannels(&config, bot)
	} else {
		go apic.JoinQueuedChannels(&config, bot)
	}

	if !config.ReceiveData {
		go func() {
			for {
				time.Sleep(time.Minute * 5)
				bot.Pong()
			}
		}()
	}

	wg.Add(1)
	wg.Wait()
}

func captureTerm(ircClient *irc.IRClient, config *utils.Config) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Println("Program killed!")
	saveConfig(*config)
	ircClient.Disconnect()
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
