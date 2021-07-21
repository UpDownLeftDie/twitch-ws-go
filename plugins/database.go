package plugins

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/configs"
	"github.com/updownleftdie/twitch-ws-go/v2/shared"
)

func SetupDB() (*sqlx.DB, error) {
	// setup environment variables
	configs.InitializeViper()
	shared.SetupDefaults()

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
		return nil, err
	}
	db.SetMaxOpenConns(viper.GetInt("DB.MAX_CONNECTIONS"))
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
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
