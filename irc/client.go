package irc

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"
	"time"
	"twitch-idle/utils"

	"github.com/asaskevich/EventBus"
	"github.com/paulbellamy/ratecounter"
)

// Client connection struct
type Client struct {
	Conn           *net.Conn
	config         *utils.Config
	bus            *EventBus.Bus
	JoinedChannels map[string]bool
	Connected      bool
	SecCounter     *ratecounter.RateCounter
	MinCounter     *ratecounter.RateCounter
	InstanceID     int
}

// Create creates a Client instance
func Create(conf *utils.Config, bus *EventBus.Bus, ID int) *Client {
	var irc *Client = &Client{
		Conn:           nil,
		config:         conf,
		bus:            bus,
		JoinedChannels: make(map[string]bool),
		Connected:      false,
		SecCounter:     ratecounter.NewRateCounter(1 * time.Second),
		MinCounter:     ratecounter.NewRateCounter(1 * time.Minute),
		InstanceID:     ID,
	}
	return irc
}

// Connect connects to server
func (irc *Client) Connect(host string, port string) {
	log.Println("Connecting to " + host + ":" + port)
	con, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		log.Println("Error:", err.Error())
		os.Exit(1)
	}
	irc.Conn = &con
	go irc.receiveData()
	irc.login(*irc.config)

	time.Sleep(2 * time.Second)
}

// Disconnect disconnects from server
func (irc *Client) Disconnect() {
	time.Sleep(time.Millisecond * 200)
	if irc.Conn != nil {
		log.Println("Disconnecting...")
		(*irc.Conn).Close()
	}
}

// Login sends username and oauth2 authentication to server
func (irc *Client) login(config utils.Config) {
	passCmd := CmdPass + " " + config.Token + "\r\n"
	userCmd := CmdUser + " " + config.User + "\r\n"
	nickCmd := CmdNick + " " + config.User + "\r\n"
	irc.sendData(passCmd + userCmd + nickCmd)
}

// Join sends a join command to a user's channel
func (irc *Client) Join(user string) {
	joinCmd := CmdJoin + " #" + user + "\r\n"

	if irc.config.LogLevel >= 2 {
		log.Printf("<- %s", joinCmd)
	}
	irc.sendData(joinCmd)
}

// Pong replies to pings with pong
func (irc *Client) Pong() {
	pongCmd := CmdPong + " :tmi.twitch.tv\r\n"
	if irc.config.LogLevel >= 1 {
		log.Printf("<- %s", pongCmd)
	}
	irc.sendData(pongCmd)
}

// send data.
func (irc *Client) sendData(message string) {
	fmt.Fprintf(*irc.Conn, "%s\r\n", message)
}

// print all incoming data
func (irc *Client) receiveData() {
	tp := textproto.NewReader(bufio.NewReader(*irc.Conn))
	for {
		message, err := tp.ReadLine()
		if err != nil {
			continue
		}

		(*irc.bus).Publish("manager:received_data", irc)

		if strings.HasPrefix(message, CmdPing) {
			if irc.config.LogLevel >= 1 {
				log.Printf("-> %s\n", message)
			}
			irc.Pong()
		}

		splitMessage := strings.Split(message, " ")

		if irc.config.LogLevel >= 3 {
			log.Println(message)
		}

		if len(splitMessage) > 2 {
			switch splitMessage[1] {
			case RplWelcome:
				log.Println("Connected")
				irc.Connected = true
			case CmdJoin:
				if irc.config.LogLevel >= 2 {
					log.Println("->", splitMessage[1], splitMessage[2])
				}
				(*irc.bus).Publish("manager:joined_channel", splitMessage[2][1:])
				irc.JoinedChannels[splitMessage[2][1:]] = true
			case CmdPrivmsg:
				if irc.config.LogLevel >= 2 && strings.Contains(splitMessage[3], irc.config.User) {
					log.Println("->", splitMessage[0], splitMessage[2], splitMessage[3])
				}
			}
		}
	}
}
