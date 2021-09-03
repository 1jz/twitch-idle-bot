package api

import (
	"chat-idle/utils"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var client = &http.Client{}

// Channel contains user and viewer count
type Channel struct {
	Username string `json:"user_login"`
	Viewers  int    `json:"viewer_count"`
}

// Cursor contains pagination cursor
type Cursor struct {
	Cursor string `json:"cursor"`
}

// Streams contains array of Channels and Pagination cursor
type Streams struct {
	Data       []Channel `json:"data"`
	Pagination Cursor    `json:"pagination"`
}

// GetChannels takes cursor (pagination) and number of streams to query for (max 100)
func GetChannels(config *utils.Config, cursor string, first string) (Streams, int, error) {
	base, err := url.Parse(TwitchAPIURLBase + "/streams")
	if err != nil {
		return Streams{}, 500, err
	}
	params := url.Values{}
	params.Add("first", first)
	params.Add("after", cursor)
	base.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", base.String(), nil)
	req.Header.Add("Client-ID", config.ClientID)
	req.Header.Add("Authorization", "Bearer "+config.AppToken.AcccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return Streams{}, 500, err
	}

	if resp.StatusCode != 200 {
		return Streams{}, resp.StatusCode, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Streams{}, resp.StatusCode, err
	}

	streams := Streams{}
	json.Unmarshal(body, &streams)

	return streams, resp.StatusCode, nil
}

// GetToken takes the client ID and secret, returns app oAuth token
func GetToken(clientID string, clientSecret string) (utils.Token, error) {
	base, err := url.Parse(TwitchIDURLBase + "/token")
	if err != nil {
		return utils.Token{}, err
	}
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("client_secret", clientSecret)
	params.Add("grant_type", "client_credentials")
	base.RawQuery = params.Encode()

	req, err := http.NewRequest("POST", base.String(), nil)

	resp, err := client.Do(req)
	if err != nil {
		return utils.Token{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return utils.Token{}, err
	}

	token := utils.Token{}
	json.Unmarshal(body, &token)
	token.Issued = time.Now().Unix()

	return token, nil
}
