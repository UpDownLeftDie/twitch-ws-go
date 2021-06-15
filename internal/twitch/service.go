package twitch

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/oauth"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/ws"
	"golang.org/x/oauth2/twitch"
)

type Client struct {
	TwitchOauthService oauth.Service
	TwitchWsConn       *websocket.Conn
	WsReceiveChan      chan []byte
}

func NewTwitchClient(db *sqlx.DB, done chan interface{}, interrupt chan os.Signal) (*Client, error) {
	wsReceiveChan := make(chan []byte)
	twitchClient, err := setup(db, wsReceiveChan, done, interrupt)
	if err != nil {
		return &Client{}, err
	}
	go twitchClient.handleWSMessages()
	return &twitchClient, nil
}

func setup(db *sqlx.DB, wsReceiveChan chan []byte, done chan interface{}, interrupt chan os.Signal) (Client, error) {
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
	var twitchWsConn *websocket.Conn
	if err != nil {
		logrus.Error("Error getting twitch token from DB: ", err)
		return Client{}, err
	} else if twitchOauthToken.ClientID == "" {
		logrus.Errorf("No twitch oauth token. Auth first: %s", viper.GetString("CALLBACK_URL"))
	} else {
		tokenSource := twitchOauthConfig.TokenSource(context.Background(), twitchOauthToken.Token())
		newToken, err := tokenSource.Token()
		if err != nil {
			logrus.Fatalln(err)
		}
		if newToken.AccessToken != twitchOauthToken.AccessToken {
			twitchOauthRepository.UpsertOauthToken(oauth.Token{
				ClientID:     twitchOauthConfig.ClientID,
				AccessToken:  newToken.AccessToken,
				RefreshToken: newToken.RefreshToken,
				TokenType:    newToken.TokenType,
			})
			log.Println("Saved new token:", newToken.AccessToken)
		}

		// setup websocket clients
		twitchTopics := viper.GetStringSlice("TWITCH.TOPICS")
		twitchWsConn, _, err := websocket.DefaultDialer.Dial("wss://pubsub-edge.twitch.tv", nil)
		if err != nil {
			logrus.Fatal("Error connecting to Websocket Server:", err)
		}
		ws.NewWebsocketClient(twitchWsConn, twitchOauthToken.AccessToken, twitchTopics, wsReceiveChan, done, interrupt)

		// setup rest clients
		// TODO
	}
	return Client{
		twitchOauthService,
		twitchWsConn,
		wsReceiveChan,
	}, nil
}

func (tw Client) handleWSMessages() {
	msg := <-tw.WsReceiveChan
	fmt.Printf("Received Twitch message: %s\n", string(msg))
	//logrus.Printf("Received Twitch message: %s\n", string(msg))
}
