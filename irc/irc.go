package irc

import (
	"bufio"
	"chat-idle/utils"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"
	"time"
)

// IRClient connection struct
type IRClient struct {
	Conn           *net.Conn
	joinedChannels map[string]bool
	config         *utils.Config
	monitoring     bool
}

// Create creates a IRClient instance
func Create(config *utils.Config) *IRClient {
	var irc *IRClient = &IRClient{
		Conn:           nil,
		joinedChannels: nil,
		config:         config,
		monitoring:     true,
	}
	(*irc).init()
	return irc
}

// Connect connects to server
func (irc *IRClient) Connect(host string, port string) {
	log.Println("Connecting to " + host + ":" + port)
	con, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		log.Println("Error:", err.Error())
		os.Exit(1)
	}

	irc.Conn = &con
	irc.login(*irc.config)

	if irc.config.ReceiveData {
		go irc.receiveData()
	}

	time.Sleep(2 * time.Second)
}

// Disconnect disconnects from server
func (irc *IRClient) Disconnect() {
	irc.monitoring = false
	time.Sleep(time.Millisecond * 200)
	if irc.Conn != nil {
		log.Println("Disconnecting...")
		(*irc.Conn).Close()
	}
}

// Login sends username and oauth2 authentication to server
func (irc *IRClient) login(config utils.Config) {
	passCmd := CmdPass + " " + config.Token + "\r\n"
	userCmd := CmdUser + " " + config.User + "\r\n"
	nickCmd := CmdNick + " " + config.User + "\r\n"
	irc.sendData(passCmd + userCmd + nickCmd)
}

// Join sends a join command to a user's channel
func (irc *IRClient) Join(user string) {
	joinCmd := CmdJoin + " #" + user + "\r\n"

	if !irc.config.ReceiveData {
		irc.joinedChannels[user] = true
	}

	if irc.config.LogLevel >= 2 {
		log.Printf("<- %s", joinCmd)
	}
	irc.sendData(joinCmd)
}

// Pong replies to pings with pong
func (irc *IRClient) Pong() {
	pongCmd := CmdPong + " :tmi.twitch.tv\r\n"
	if irc.config.LogLevel >= 1 {
		log.Printf("<- %s", pongCmd)
	}
	irc.sendData(pongCmd)
}

// send data.
func (irc *IRClient) sendData(message string) {
	fmt.Fprintf(*irc.Conn, "%s\r\n", message)
}

// print all incoming data
func (irc *IRClient) receiveData() {

	if irc.config.LogLevel >= 1 {
		log.Println("Tracking incoming data...")
	}

	tp := textproto.NewReader(bufio.NewReader(*irc.Conn))

	for {
		message, err := tp.ReadLine()
		if err != nil {
			if !irc.monitoring {
				panic(err)
			}
			return
		}
		if strings.HasPrefix(message, CmdPing) {
			if irc.config.LogLevel >= 2 {
				log.Printf("-> %s\n", message)
			}
			irc.Pong()
		}

		splitMessage := strings.Split(message, " ")

		if irc.config.LogLevel >= 3 {
			log.Println(message)
		} else if len(splitMessage) > 2 {
			switch splitMessage[1] {
			case CmdJoin:
				if irc.config.LogLevel >= 2 {
					log.Println("->", splitMessage[1], splitMessage[2])
				}
				if irc.config.ReceiveData {
					irc.joinedChannels[splitMessage[2][1:]] = true
				}
			case CmdPrivmsg:
				if irc.config.LogLevel >= 2 && strings.Contains(splitMessage[3], irc.config.User) {
					log.Println("->", splitMessage[0], splitMessage[2], splitMessage[3])
				}
			}
		}
	}
}

// InChannel returns if bot has joined specified channel
func (irc *IRClient) InChannel(channel string) bool {
	_, exists := irc.joinedChannels[channel]
	return exists
}

// init initializes stuff
func (irc *IRClient) init() {
	irc.joinedChannels = make(map[string]bool)
}
