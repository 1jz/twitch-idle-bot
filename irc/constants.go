package irc

// IRC Reply Code consts
const (
	RplWelcome  string = "001"
	RplYourhost string = "002"
	RplCreated  string = "003"
	RplMyInfo   string = "004"
	RplBounce   string = "005"

	RplNamReply   string = "353"
	RplEndOfNames string = "366"

	RplMOTDStart string = "375"
	RplMOTD      string = "372"
	RplEndOfMOTD string = "376"
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
