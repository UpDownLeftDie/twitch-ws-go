package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/configs"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/oauth"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/ws"
	"golang.org/x/oauth2/twitch"
)

var (
	// Version is injected by go (should be a tag name)
	Version = "None"
	// Buildstamp is a timestamp (injected by go) of the build time
	Buildstamp = "None"
	// Githash is the tag for current hash the build represents
	Githash = "None"
	host    = "None"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "adapter",
	Short: "adapter",
	Long:  `An interface between APIs & Websockets`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setup(ctx context.Context, done chan interface{}, interrupt chan os.Signal) (oauth.Service, ws.Ws) {
	// setup environment variables
	configs.InitializeViper()
	setupDefaults()

	var err error
	host, err = os.Hostname()
	if err != nil {
		logrus.Panicln("unable to get Hostname", err)
	}

	// setup logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.WithFields(logrus.Fields{
		"Version":   Version,
		"BuildTime": Buildstamp,
		"Githash":   Githash,
		"Host":      host,
	}).Info("Service Startup")

	// setup db
	db, err := getDB(
		viper.GetString("DB.HOST"),
		viper.GetString("DB.PORT"),
		viper.GetString("DB.USER"),
		viper.GetString("DB.PASSWORD"),
		viper.GetString("DB.NAME"),
		viper.GetString("DB.SSLMODE"),
		viper.GetString("SERVICE_NAME"),
		viper.GetString("DB.SQLX_DRIVER_NAME"),
		viper.GetString("DB.SQLITE_FILE"),
	)
	if err != nil {
		logrus.Panicln("unable to connect to DB", err)
	}
	db.SetMaxOpenConns(viper.GetInt("DB.MAX_CONNECTIONS"))
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

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
	} else if twitchOauthToken.ClientID == "" {
		logrus.Error("No twitch oauth token. Auth first: %s", viper.GetString("CALLBACK_URL"))
	} else {
		// setup websocket clients
		twitchTopics := viper.GetStringSlice("TWITCH.TOPICS")
		twitchWebSocketClient := ws.NewWebsocketClient("wss://pubsub-edge.twitch.tv", twitchOauthToken.ClientID, twitchTopics, done, interrupt)

		// setup rest clients
		// TODO

		return twitchOauthService, twitchWebSocketClient
	}
	return twitchOauthService, nil

}

func setupDefaults() map[string]interface{} {
	viper.AutomaticEnv()
	defaults := map[string]interface{}{
		"APP_ENV":      "local",
		"PORT":         3000,
		"CALLBACK_URL": fmt.Sprintf("http://localhost:%d/callback", 8080),

		"DB.MAX_CONNECTIONS":  5,
		"DB.SSLMODE":          "disable",
		"DB.SQLX_DRIVER_NAME": "sqlite3",
		"DB.SQLITE_FILE":      "twitch-ws-go.db",

		"TWITCH.BASE_API_URL": "https://id.twitch.tv",
		"TWITCH.WS_URL":       "wss://pubsub-edge.twitch.tv",
	}
	for key, value := range defaults {
		viper.SetDefault(key, value)
	}
	return defaults
}

func getDB(host string, port string, user string, password string, dbname string, sslmode string, serviceName string, driverName string, sqliteFilePath string) (*sqlx.DB, error) {
	switch driverName {
	case "sqlite3":
		return sqlx.Connect(driverName, sqliteFilePath)

	default:
		var pairs = make([]string, 0, 7)
		if host != "" {
			pairs = append(pairs, fmt.Sprintf("host=%s", host))
		}
		if port != "" {
			pairs = append(pairs, fmt.Sprintf("port=%s", port))
		}
		if user != "" {
			pairs = append(pairs, fmt.Sprintf("user=%s", user))
		}
		if password != "" {
			pairs = append(pairs, fmt.Sprintf("password=%s", password))
		}
		if dbname != "" {
			pairs = append(pairs, fmt.Sprintf("dbname=%s", dbname))
		}
		if sslmode != "" {
			pairs = append(pairs, fmt.Sprintf("sslmode=%s", sslmode))
		}
		pairs = append(pairs, fmt.Sprintf("application_name=%s", serviceName))
		return sqlx.Connect(driverName, strings.Join(pairs, " "))
	}
}

//func convertTwitchScopesToTopics(scopes []string, twitchId string) []string {
//	var topics []string
//	var twitchTopics = map[string][]string{
//		"bits:read":                  {"channel-bits-events-v2.%s", "channel-bits-badge-unlocks.%s"},
//		"channel:read:redemptions":   {"channel-points-channel-v1.%s"},
//		"channel:read:subscriptions": {"channel-subscribe-events-v1.%s"},
//		"channel:moderate":           {"chat_moderator_actions.%s.%s", "automod-queue.%s.%s"},
//		"chat:read":                  {"user-moderation-notifications.%s.%s"},
//		"whispers:read":              {"whispers.<user ID>"},
//	}
//
//	for i := range scopes {
//		for j := range twitchTopics[scopes[i]] {
//			topic := fmt.Sprintf(twitchTopics[scopes[i]][j], twitchId)
//			topics = append(topics, topic)
//		}
//	}
//	return topics
//}
