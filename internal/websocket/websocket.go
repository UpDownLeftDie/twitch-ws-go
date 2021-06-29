package websocket

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	gorillaWs "github.com/gorilla/websocket"
)

//type WebSocket interface {
//	SendMessage(message string) error
//}

type websocket struct {
	conn      *gorillaWs.Conn
	done      chan interface{} // Channel to indicate that the receiverHandler is done
	interrupt chan os.Signal   // Channel to listen for interrupt signal to terminate gracefully
}

func NewWebsocketClient(conn *gorillaWs.Conn, oauthToken string, topics []string, wsReceiveChan chan []byte) error {
	go receiveHandler(conn, wsReceiveChan)

	err := conn.WriteJSON(twitchWSOutgoingMessage{Type: "LISTEN", Nonce: "twitchPubSubNonce", Data: authMessageData{AuthToken: oauthToken, Topics: topics}})
	if err != nil {
		logrus.Println("Error during LISTEN to websocket:", err)
		return err
	}

	// Our main loop for the client
	// We send our relevant packets here
	go func() {
		for {
			select {
			case <-time.After(time.Duration(5) * time.Millisecond * 1000):
				// Send an echo packet every second
				err := conn.WriteJSON(twitchWSOutgoingMessage{Type: "PING"})
				if err != nil {
					logrus.Println("Error during writing to websocket:", err)
					return
				}
			}
		}
	}()

	return nil
}

func (ws websocket) SendMessage(message string) error {
	err := ws.conn.WriteMessage(gorillaWs.TextMessage, []byte(message))
	if err != nil {
		logrus.Println("Error during writing to websocket:", err)
		return err
	}
	return nil
}

func receiveHandler(conn *gorillaWs.Conn, wsReceiveChan chan []byte) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), strconv.Itoa(gorillaWs.CloseNormalClosure)) || strings.Contains(err.Error(), strconv.Itoa(gorillaWs.ClosePolicyViolation)) {
				return
			}
			logrus.Println("Error in receive:", err)
			return
		}
		wsReceiveChan <- msg
		//fmt.Printf("inside receiveHandler: %s", msg)
	}
}

func (ws websocket) Stop(done chan interface{}) {
	// We received a SIGINT (Ctrl + C). Terminate gracefully...
	logrus.Println("Received SIGINT interrupt signal. Closing all pending connections")

	// Close our websocket connection
	//err := conn.WriteJSON(twitchWSOutgoingMessage{Type: "DISCONNECT"})
	err := ws.conn.WriteMessage(gorillaWs.CloseMessage, gorillaWs.FormatCloseMessage(gorillaWs.CloseNormalClosure, ""))
	if err != nil {
		logrus.Println("Error during closing websocket:", err)
		return
	}

	select {
	case <-done:
		logrus.Println("Receiver Channel Closed! Exiting....")
		return
	case <-time.After(time.Duration(5) * time.Second):
		logrus.Println("Timeout in closing receiving channel. Exiting....")
		return
	}
}
