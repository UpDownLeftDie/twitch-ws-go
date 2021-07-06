package websocket

import (
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	gorillaWs "github.com/gorilla/websocket"
)

//type WebSocket interface {
//	SendMessage(message string) error
//}

type Websocket struct {
	conn *gorillaWs.Conn
}

func NewWebsocketClient(conn *gorillaWs.Conn, oauthToken string, topics []string, receiveChan chan []byte) (*Websocket, error) {
	go receiveHandler(conn, receiveChan)

	err := conn.WriteJSON(twitchWSOutgoingMessage{Type: "LISTEN", Nonce: "twitchPubSubNonce", Data: authMessageData{AuthToken: oauthToken, Topics: topics}})
	if err != nil {
		logrus.Println("Error during LISTEN to websocket:", err)
		return &Websocket{}, err
	}

	// Our main loop for the client
	// We send our relevant packets here
	go func() error {
		for {
			select {
			case <-time.After(time.Duration(5) * time.Millisecond * 1000):
				// Send an echo packet every second
				err := conn.WriteJSON(twitchWSOutgoingMessage{Type: "PING"})
				if err != nil {
					logrus.Println("Error during writing to websocket:", err)
					return err
				}
			}
		}
	}()

	return &Websocket{conn}, nil
}

func (ws Websocket) SendMessage(message string) error {
	err := ws.conn.WriteMessage(gorillaWs.TextMessage, []byte(message))
	if err != nil {
		logrus.Println("Error during writing to websocket:", err)
		return err
	}
	return nil
}

func receiveHandler(conn *gorillaWs.Conn, receiveChan chan []byte) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), strconv.Itoa(gorillaWs.CloseNormalClosure)) || strings.Contains(err.Error(), strconv.Itoa(gorillaWs.ClosePolicyViolation)) {
				return
			}
			logrus.Println("Error in receive:", err)
			return
		}
		receiveChan <- msg
		//fmt.Printf("inside receiveHandler: %s", msg)
	}
}

func (ws Websocket) Stop() {
	// We received a SIGINT (Ctrl + C). Terminate gracefully...
	logrus.Println("Received SIGINT interrupt signal. Closing all pending connections")

	// Close our websocket connection
	//err := conn.WriteJSON(twitchWSOutgoingMessage{Type: "DISCONNECT"})
	err := ws.conn.WriteMessage(gorillaWs.CloseMessage, gorillaWs.FormatCloseMessage(gorillaWs.CloseNormalClosure, ""))
	if err != nil {
		logrus.Println("Error during closing websocket: ", err)
		return
	}
	logrus.Println("Timeout in closing receiving channel. Exiting....")
	ws.conn.Close()
	return
}
