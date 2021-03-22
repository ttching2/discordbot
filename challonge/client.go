package challonge

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const challongeAPIURL = "https://%s:%s@api.challonge.com/v1/"

type Client struct {
	Tournament  *tournamentClient
	Match       *matchClient
	Participant *participantsClient
}

type baseClient struct {
	httpClient http.Client
	Config
}

type Config struct {
	Username string
	Apikey   string
}

func New(config Config) *Client {
	b := baseClient{
		httpClient: *http.DefaultClient,
		Config:     config}
	return &Client{
		Tournament:  &tournamentClient{baseClient: b},
		Match:       &matchClient{baseClient: b},
		Participant: &participantsClient{baseClient: b},
	}
}

func (c *baseClient) getAPIURL() string {
	return fmt.Sprintf(challongeAPIURL, c.Username, c.Apikey)
}

func (c *baseClient) getRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error status code %v", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *baseClient) putRequest(url string) error {
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return  err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("error status code %v", res.StatusCode)
	}

	return  nil
}
