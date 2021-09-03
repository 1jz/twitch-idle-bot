package irc

// IRC Reply Code consts
const (
	RplWelcome  int = 001
	RplYourhost int = 002
	RplCreated  int = 003
	RplMyInfo   int = 004
	RplBounce   int = 005

	RplNamReply   int = 353
	RplEndOfNames int = 366

	RplMOTDStart int = 375
	RplMOTD      int = 372
	RplEndOfMOTD int = 376
)

// IRC Command consts
const (
	CmdJoin    string = "JOIN"
	CmdNick    string = "NICK"
	CmdPass    string = "PASS"
	CmdPing    string = "PING"
	CmdPong    string = "PONG"
	CmdPrivmsg string = "PRIVMSG"
	CmdUser    string = "USER"
)
