package main

import (
	"github.com/sirupsen/logrus"
	"github.com/updownleftdie/twitch-ws-go/v3/plugins"
)

func (tc *Client) Start() {
	logrus.Println("Starting Twitch Client...")
	receiveChan := make(chan []byte)
	db, err := plugins.SetupDB()
	if err != nil {
		logrus.Error(err)
	}
	websocketClient, err := setup(db, receiveChan)
	if err != nil {
		logrus.Error(err)
	}
	tc.WebsocketClient = websocketClient
	tc.receiveChan = receiveChan
	tc.db = db

	go func() {
		for {
			msg := getWSMessage(receiveChan)
			if msg != "" {
				// TODO handle your events here
			}
		}
	}()
}
func (tc *Client) Stop() {
	tc.WebsocketClient.Stop()
}
