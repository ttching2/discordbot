package strawpoll

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	httpClient http.Client
	apiKey string
}

type StrawPollConfig struct {
	ApiKey string
}

func New(config StrawPollConfig) *Client {
	return &Client {
		httpClient: *http.DefaultClient,
		apiKey: config.ApiKey,
	}
}

func (c *Client) GetPoll(ID string) (*StrawPollResults, error) {
	
	req, err := http.NewRequest("GET", strawpollGetEndpoint + ID , nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("API-KEY", c.apiKey)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	strawPollResults := StrawPollResults{}
	err = json.Unmarshal([]byte(body), &strawPollResults)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	
	return &strawPollResults, nil
}