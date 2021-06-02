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
	"github.com/updownleftdie/twitch-ws-go/v2/internal/service"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/service/twitch"
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

func setup(ctx context.Context) service.Service {
	configs.InitializeViper()
	setupDefaults()

	var err error
	host, err = os.Hostname()
	if err != nil {
		logrus.Panicln("unable to get Hostname", err)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.WithFields(logrus.Fields{
		"Version":   Version,
		"BuildTime": Buildstamp,
		"Githash":   Githash,
		"Host":      host,
	}).Info("Service Startup")

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

	twitchOauthConfig := twitch.NewOAuthConfig(
		viper.GetString("TWITCH.CLIENT_ID"),
		viper.GetString("TWITCH.CLIENT_SECRET"),
	)

	twitchRepository := service.NewRepository(db)
	twitchService := service.NewService(twitchOauthConfig, viper.GetString("TWITCH.BASE_API_URL"), twitchRepository)
	return twitchService
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
		"DB.SQLITE_FILE":      "db.db",

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
