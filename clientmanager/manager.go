package clientmanager

import (
	"fmt"
	"log"
	"sync"
	"time"
	"twitch-idle/api"
	"twitch-idle/irc"
	"twitch-idle/utils"

	"github.com/asaskevich/EventBus"
	"github.com/paulbellamy/ratecounter"
)

// ClientManager ...
type ClientManager struct {
	mut           sync.Mutex
	clients       map[int]*irc.Client
	bus           *EventBus.Bus
	channelsQueue []string
	config        *utils.Config
	host          string
	port          string

	CID             int // current instance ID
	InstanceCount   int
	ActiveInstances int
	JoinedChannels  map[string]bool
	ChannelCount    int
	SecCounter      *ratecounter.RateCounter
	MinCounter      *ratecounter.RateCounter
}

// Setup ...
func Setup(bus *EventBus.Bus, conf *utils.Config, host string, port string) *ClientManager {
	var manager = &ClientManager{
		clients:         make(map[int]*irc.Client),
		bus:             bus,
		channelsQueue:   conf.Channels,
		config:          conf,
		host:            host,
		port:            port,
		CID:             0,
		InstanceCount:   0,
		ActiveInstances: 0,
		JoinedChannels:  make(map[string]bool),
		SecCounter:      ratecounter.NewRateCounter(1 * time.Second),
		MinCounter:      ratecounter.NewRateCounter(1 * time.Minute),
	}
	return manager
}

// Start ...
func (manager *ClientManager) Start() {
	// initial instance
	manager.InstanceCount++
	manager.CID = manager.InstanceCount
	manager.clients[manager.CID] = irc.Create(manager.config, manager.bus, manager.InstanceCount)
	manager.clients[manager.CID].Connect(manager.host, manager.port)

	time.Sleep(time.Duration(manager.config.StartDelay) * time.Millisecond)
	fmt.Println("Joining channels...")

	// rate printer
	go func() {
		delay := 20 * time.Millisecond
		for {
			fmt.Printf("\r%5d gm/s %5d gm/m | %5d im/s %5d im/m | CC: %d\r", manager.SecCounter.Rate(), manager.MinCounter.Rate(),
				manager.clients[manager.CID].SecCounter.Rate(), manager.clients[manager.CID].MinCounter.Rate(), manager.ChannelCount)
			time.Sleep(delay)
		}
	}()

	// join queue
	go func() {
		delay := time.Second * time.Duration(manager.config.JoinInterval/1000)
		for {
			if len(manager.channelsQueue) == 0 {
				time.Sleep(time.Second * 30)
			} else {
				if !manager.clients[manager.CID].Connected {
					manager.clients[manager.CID].Disconnect()
					manager.clients[manager.CID].Connect(manager.host, manager.port)
				} else {
					if manager.clients[manager.CID].MinCounter.Rate() >= 4500 {
						manager.InstanceCount++
						manager.CID = manager.InstanceCount
						manager.clients[manager.CID] = irc.Create(manager.config, manager.bus, manager.InstanceCount)
						manager.clients[manager.CID].Connect(manager.host, manager.port)
						fmt.Println("\rnew instance created\t\t\t\t\t\t")
					}
					channel := manager.channelsQueue[0]

					manager.mut.Lock()
					manager.channelsQueue = manager.channelsQueue[1:]
					manager.mut.Unlock()

					manager.clients[manager.CID].Join(channel)

					time.Sleep(delay)
				}
			}
		}
	}()

	// check for dead instances
	go func() {
		delay := 1 * time.Minute
		for {
			for i := 1; i <= manager.InstanceCount; i++ {
				if manager.clients[i].MinCounter.Rate() == 0 {
					manager.clients[i].Disconnect()
					for k := range manager.clients[i].JoinedChannels {
						manager.JoinedChannels[k] = false
					}
					//delete(manager.clients, manager.clients[i].InstanceID)
				}
			}
			time.Sleep(delay)
		}
	}()

	if manager.config.JoinAllLiveChannels {
		go manager.ScanChannels()
	}
}

// AppendToJoinQueue ...
func (manager *ClientManager) AppendToJoinQueue(channels []string) {
	manager.mut.Lock()
	manager.channelsQueue = append(manager.channelsQueue, channels...)
	manager.mut.Unlock()
}

// ScanChannels starts scanning for live channels
func (manager *ClientManager) ScanChannels() {
	cursor := ""
	config := manager.config
	for {
		currentTime := time.Now().Unix()
		tokenExpiry := time.Unix(config.AppToken.Issued+config.AppToken.ExpiresIn, 0).Unix()
		if currentTime > tokenExpiry {
			log.Println("token expired")
			config.AppToken, _ = api.GetToken(config.ClientID, config.ClientSecret)
		}

		streams, statusCode, _ := api.GetChannels(config, cursor, "100")
		if statusCode != 200 {
			config.AppToken = utils.Token{}
		}

		cursor = streams.Pagination.Cursor
		if len(streams.Data) != 0 && streams.Data[0].Viewers < config.ViewersMin {
			cursor = ""
		}

		var channelsToJoin []string
		for _, channel := range streams.Data {
			if channel.Viewers <= config.ViewersMax && !manager.InChannel(channel.Username) {
				channelsToJoin = append(channelsToJoin, channel.Username)
			}
		}

		manager.AppendToJoinQueue(channelsToJoin)

		time.Sleep(time.Second * 30)
	}
}

// InChannel returns if bot has joined specified channel
func (manager *ClientManager) InChannel(channel string) bool {
	_, exists := manager.JoinedChannels[channel]
	return exists
}

// JoinedChannel ...
func (manager *ClientManager) JoinedChannel(channel string) {
	fmt.Printf("\rjoined %s\t\t\t\t\t\t\n", channel)
	manager.JoinedChannels[channel] = true
	manager.ChannelCount++
}

// ReceivedData ...
func (manager *ClientManager) ReceivedData(client *irc.Client) {
	client.SecCounter.Incr(1)
	client.MinCounter.Incr(1)
	manager.SecCounter.Incr(1)
	manager.MinCounter.Incr(1)
}
