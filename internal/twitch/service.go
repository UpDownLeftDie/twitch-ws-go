package twitch

import (
	"github.com/sirupsen/logrus"
)

func (tc Client) Start() {
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

func (tc Client) getWSMessage() string {
	msg := <-tc.wsReceiveChan
	logrus.Debugf("Received Twitch message: %s\n", string(msg))
	return string(msg)
}
