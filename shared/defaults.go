package shared

import (
	"fmt"
	"github.com/spf13/viper"
)

func SetupDefaults() map[string]interface{} {
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
