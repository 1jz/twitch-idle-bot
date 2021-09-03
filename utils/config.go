package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config contains User and OAuth strings and array of channels
type Config struct {
	User                string   `json:"user"`
	Token               string   `json:"token"`
	Channels            []string `json:"channels"`
	StartDelay          int      `json:"start_delay"`
	ClientID            string   `json:"client_id"`
	ClientSecret        string   `json:"client_secret"`
	AppToken            Token    `json:"app_token"`
	LogLevel            int      `json:"log_level"`
	JoinInterval        int      `json:"join_interval"`
	ViewersMin          int      `json:"viewers_min"`
	ViewersMax          int      `json:"viewers_max"`
	ReceiveData         bool     `json:"receive_data"`
	JoinAllLiveChannels bool     `json:"join_all_live_channels"`
}

// Token holds access token info.
type Token struct {
	AcccessToken string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Issued       int64  `json:"issued"`
}

// GetConfig extracts user info from auth.json
func GetConfig() (Config, error) {
	var config Config

	file, err := os.Open("config.json")
	if err != nil {
		return config, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return config, err
	}

	json.Unmarshal(data, &config)

	return config, nil
}
