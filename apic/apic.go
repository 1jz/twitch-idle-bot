// Package apic == API Caller
package apic

import (
	"chat-idle/api"
	"chat-idle/irc"
	"chat-idle/utils"
	"sync"
	"time"
)

var channelsQueue []string
var mut sync.Mutex

// ScanChannels starts scanning for live channels
func ScanChannels(config *utils.Config, bot *irc.IRClient) {
	var cursor string = ""
	for {
		streams, statusCode, _ := api.GetChannels(config, cursor, "100")
		if statusCode != 200 {
			config.AppToken = utils.Token{}
		}

		cursor = streams.Pagination.Cursor
		if len(streams.Data) != 0 && streams.Data[0].Viewers < config.ViewersMin {
			cursor = ""
		}

		for _, channel := range streams.Data {
			if channel.Viewers <= config.ViewersMax && !bot.InChannel(channel.Username) {
				mut.Lock()
				channelsQueue = append(channelsQueue, channel.Username)
				mut.Unlock()
			}
		}
		time.Sleep(time.Second * 30)
	}
}

// QueueChannels queues channels from config
func QueueChannels(config *utils.Config) {
	channelsQueue = config.Channels[:]
}

// JoinChannels joins channels from queue continuously
func JoinChannels(config *utils.Config, bot *irc.IRClient) {
	for {
		if len(channelsQueue) == 0 {
			time.Sleep(time.Second * 30)
		} else {
			channel := channelsQueue[0]
			mut.Lock()
			channelsQueue = channelsQueue[1:]
			mut.Unlock()

			bot.Join(channel)
			time.Sleep(time.Second * time.Duration(config.JoinInterval/1000))
		}
	}
}

// JoinQueuedChannels joins channels from queue
func JoinQueuedChannels(config *utils.Config, bot *irc.IRClient) {
	for _, channel := range channelsQueue {
		bot.Join(channel)
		time.Sleep(time.Second * time.Duration(config.JoinInterval/1000))
	}
}
