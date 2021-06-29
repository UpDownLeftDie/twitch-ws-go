package twitch

import (
	"context"
	"fmt"
	"net/http"
	"os"

	gorillaWs "github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/oauth"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/websocket"
	"golang.org/x/oauth2/twitch"
)

type Client struct {
	TwitchWsConn  *gorillaWs.Conn
	wsReceiveChan chan []byte
}

func NewTwitchClient(db *sqlx.DB, done chan interface{}, interrupt chan os.Signal) (*Client, error) {
	wsReceiveChan := make(chan []byte)
	twitchClient, err := setup(db, wsReceiveChan, done, interrupt)
	if err != nil {
		return &Client{}, err
	}

	return &twitchClient, nil
}

func setup(db *sqlx.DB, wsReceiveChan chan []byte, done chan interface{}, interrupt chan os.Signal) (Client, error) {
	var twitchWsConn *gorillaWs.Conn

	// get oauth token
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
		return Client{}, err
	} else if twitchOauthToken.ClientID == "" {
		port := viper.GetInt("PORT")
		logrus.Errorf("No twitch oauth token. Auth first: http://localhost:%d", port)
		var responseCode int
		mux := oauth.MakeHTTPHandler(twitchOauthService, &responseCode)

		oauthServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		}
		go oauthServer.ListenAndServe()
		for responseCode != http.StatusOK {
			select {
			case <-interrupt:
				return Client{}, nil
			default:
			}
		}
		oauthServer.Shutdown(context.Background())
		twitchOauthToken, err = twitchOauthRepository.GetOauthToken()
		if err != nil {
			logrus.Error("Error getting twitch token from DB: ", err)
			return Client{}, err
		}
	}

	tokenSource := twitchOauthConfig.TokenSource(context.Background(), twitchOauthToken.Token())
	newToken, err := tokenSource.Token()
	if err != nil {
		logrus.Fatalln(err)
	}
	if newToken.AccessToken != twitchOauthToken.AccessToken {
		err = twitchOauthRepository.UpsertOauthToken(newToken, twitchOauthConfig.ClientID)
		if err != nil {
			logrus.Errorln("Error updating Twitch token: ", err)
		} else {
			logrus.Printf("Twitch token updated! (%s...)", newToken.AccessToken[0:5])
		}
	}

	// setup websocket clients
	twitchTopics := viper.GetStringSlice("TWITCH.TOPICS")
	twitchWsConn, _, err = gorillaWs.DefaultDialer.Dial("wss://pubsub-edge.twitch.tv", nil)
	if err != nil {
		logrus.Fatal("Error connecting to Websocket Server:", err)
	}
	websocket.NewWebsocketClient(twitchWsConn, twitchOauthToken.AccessToken, twitchTopics, wsReceiveChan)

	return Client{
		twitchWsConn,
		wsReceiveChan,
	}, nil
}

func (tc Client) Stop(done chan interface{}) {
	tc.Stop(done)
}
