package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/go-plugin"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/updownleftdie/twitch-ws-go/v3/internal/oauth"
	websocketClient "github.com/updownleftdie/twitch-ws-go/v3/internal/websocket"
	"github.com/updownleftdie/twitch-ws-go/v3/shared"
	"golang.org/x/oauth2/twitch"
)

type Client struct {
	WebsocketClient *websocketClient.Websocket
	receiveChan     chan []byte
	db              *sqlx.DB
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	logger.Debug("message from plugin", "twitch")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		//Plugins: map[string]plugin.Plugin{
		//	"twitch": &shared.CustomPlugin{Impl: },
		//},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

func setup(db *sqlx.DB, receiveChan chan []byte) (*websocketClient.Websocket, error) {
	var twitchWsConn *websocket.Conn

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
		return &websocketClient.Websocket{}, err
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
			// TODO stop()
		}
		oauthServer.Shutdown(context.Background())
		twitchOauthToken, err = twitchOauthRepository.GetOauthToken()
		if err != nil {
			logrus.Error("Error getting twitch token from DB: ", err)
			return &websocketClient.Websocket{}, err
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
	twitchWsConn, _, err = websocket.DefaultDialer.Dial("wss://pubsub-edge.twitch.tv", nil)
	if err != nil {
		logrus.Fatal("Error connecting to Websocket Server:", err)
	}
	WebsocketClient, err := websocketClient.NewWebsocketClient(twitchWsConn, twitchOauthToken.AccessToken, twitchTopics, receiveChan)
	if err != nil {
		logrus.Error("Failed to setup websocket client:", err)
		return &websocketClient.Websocket{}, err
	}

	return WebsocketClient, nil
}

func getWSMessage(receiveChan <-chan []byte) string {
	msg := <-receiveChan
	logrus.Debugf("Received Twitch message: %s\n", string(msg))
	return string(msg)
}
