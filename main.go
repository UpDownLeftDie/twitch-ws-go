package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sacOO7/socketcluster-client-go/scclient"
	"github.com/spf13/viper"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/configs"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/logger"
	"github.com/updownleftdie/twitch-ws-go/v2/internal/services"
)

var httpClient = http.Client{}

func onConnect(client scclient.Client) {
	fmt.Println("Connected to server")
}

func onDisconnect(client scclient.Client, err error) {
	fmt.Printf("Error: %s\n", err.Error())
}

func onConnectError(client scclient.Client, err error) {
	fmt.Printf("Error: %s\n", err.Error())
}

func onSetAuthentication(client scclient.Client, token string) {
	fmt.Println("Auth token received :", token)

}

func onAuthentication(client scclient.Client, isAuthenticated bool) {
	fmt.Println("Client authenticated :", isAuthenticated)
	go startCode(client)
}

func main() {
	// Initialize Viper across the application
	configs.InitializeViper()

	// Initialize Logger across the application
	logger.InitializeZapCustomLogger()

	// Initialize Oauth2 Services
	services.InitializeOAuthTwitch()

	// Routes for the application
	http.HandleFunc("/", services.HandleMain)
	http.HandleFunc("/login-twitch", services.HandleTwitchLogin)
	http.HandleFunc("/callback", services.CallBackFromTwitch)

	// logger.Log.Info("Started running on http://localhost:" + viper.GetString("port"))
	log.Fatal(http.ListenAndServe(":"+viper.GetString("port"), nil))

	// var reader scanner.Scanner
	// client := scclient.New("wss://pubsub-edge.twitch.tv")
	// client.SetBasicListener(onConnect, onConnectError, onDisconnect)
	// client.SetAuthenticationListener(onSetAuthentication, onAuthentication)
	// go client.Connect()

	// fmt.Println("Enter any key to terminate the program")
	// reader.Init(os.Stdin)
	// reader.Next()
	// os.Exit(0)
}

func startCode(client scclient.Client) {

	// client.Subscribe("mychannel")

	// //with acknowledgement
	// client.SubscribeAck("mychannel", func(channelName string, error interface{}, data interface{}) {
	// 	if error == nil {
	// 		fmt.Println("Subscribed to channel ", channelName, "successfully")
	// 	}
	// })
}
