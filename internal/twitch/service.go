package twitch

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/oauth"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/ws"
	"golang.org/x/oauth2/twitch"
)

type Client interface{}
type client struct {
	twitchOauthService oauth.Service
	wsReceiveChan      chan []byte
}

func NewTwitchClient(db *sqlx.DB, done chan interface{}, interrupt chan os.Signal) (oauth.Service, error) {
	wsReceiveChan := make(chan []byte)
	twitchClient, err := setup(db, wsReceiveChan, done, interrupt)
	if err != nil {
		return nil, err
	}
	go twitchClient.handleWSMessages()
	return twitchClient.twitchOauthService, nil
}

func setup(db *sqlx.DB, wsReceiveChan chan []byte, done chan interface{}, interrupt chan os.Signal) (client, error) {
	// get oauth tokens
	twitchOauthConfig := oauth.NewOAuthConfig(
		viper.GetString("TWITCH.CLIENT_ID"),
		viper.GetString("TWITCH.CLIENT_SECRET"),
		viper.GetStringSlice("TWITCH.SCOPES"),
		twitch.Endpoint,
	)

	twitchOauthRepository := oauth.NewRepository(db, twitchOauthConfig)
	twitchOauthService := oauth.NewService(twitchOauthConfig, viper.GetString("TWITCH.BASE_API_URL"), twitchOauthRepository)

	twitchOauthToken, err := twitchOauthRepository.GetOauthToken()
	if err != nil {
		logrus.Error("Error getting twitch token from DB: ", err)
		return client{}, err
	} else if twitchOauthToken.ClientID == "" {
		logrus.Errorf("No twitch oauth token. Auth first: %s", viper.GetString("CALLBACK_URL"))
	} else {
		// setup websocket clients
		twitchTopics := viper.GetStringSlice("TWITCH.TOPICS")
		ws.NewWebsocketClient("wss://pubsub-edge.twitch.tv", twitchOauthToken.AccessToken, twitchTopics, wsReceiveChan, done, interrupt)
		// setup rest clients
		// TODO
	}
	return client{
		twitchOauthService,
		wsReceiveChan,
	}, nil
}

func (tw client) handleWSMessages() {
	msg := <-tw.wsReceiveChan
	fmt.Printf("Received Twitch message: %s\n", string(msg))
	//logrus.Printf("Received Twitch message: %s\n", string(msg))
}
