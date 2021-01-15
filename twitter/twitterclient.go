package twitter

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type TwitterClientConfig struct {
	ConsumerKey string
	ConsumerSecret string
	AccessToken string
	AccessSecret string
}

type TwitterClient struct {
	Client *twitter.Client
	streamFilterParams twitter.StreamFilterParams
	demux twitter.SwitchDemux
	followStream *twitter.Stream
}

func NewClient(config TwitterClientConfig) *TwitterClient {
	//TODO obviously replace these
	consumerKey := config.ConsumerKey
	consumerSecret := config.ConsumerSecret
	accessToken := config.AccessToken
	accessSecret := config.AccessSecret
	oauthConfig := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := oauthConfig.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	demux := twitter.NewSwitchDemux()

	return &TwitterClient {
		Client: client,
		demux: demux,
	}

}

func (c *TwitterClient) SetTweetDemux(fnc func(tweet *twitter.Tweet)) {
	c.demux.Tweet = fnc
}

func (c *TwitterClient) AddUserToTrack(user string) {
	c.streamFilterParams.Follow = append(c.streamFilterParams.Follow, user)
	c.StartFilterStream()
}

func (c *TwitterClient) RemoveUserFromFollowList(user string) {
	for i := range c.streamFilterParams.Follow {
		if c.streamFilterParams.Follow[i] == user {
			c.streamFilterParams.Follow = append(c.streamFilterParams.Follow[:i], c.streamFilterParams.Follow[i+1:]...)
			c.StartFilterStream()
			return
		}
	}
}

func (c *TwitterClient) StartFilterStream() {
	if c.followStream != nil {
		c.followStream.Stop()
	}
	stream, err := c.Client.Streams.Filter(&c.streamFilterParams)

	if err != nil {
		return
	}

	go c.demux.HandleChan(stream.Messages)
	c.followStream = stream
}