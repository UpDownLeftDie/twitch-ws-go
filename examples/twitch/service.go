package main

import (
	"github.com/sirupsen/logrus"
)

// @param db *sqlx.DB
func (tc Client) Start(args ...*interface{}) {
	logrus.Println("Starting Twitch Client...")

	go func() {
		for {
			msg := tc.getWSMessage()
			if msg != "" {
				// TODO handle your events here
			}
		}
	}()
}
func (tc Client) Stop() {
	tc.WebsocketClient.Stop()
}

func (tc Client) getWSMessage() string {
	msg := <-tc.receiveChan
	logrus.Debugf("Received Twitch message: %s\n", string(msg))
	return string(msg)
}
