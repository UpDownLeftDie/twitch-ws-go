package ws

import (
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type Ws interface {
	SendMessage(message string) error
}

type ws struct {
	conn      *websocket.Conn
	done      chan interface{} // Channel to indicate that the receiverHandler is done
	interrupt chan os.Signal   // Channel to listen for interrupt signal to terminate gracefully
}

func NewWebsocketClient(websocketUrl string, oauthToken string, topics []string, done chan interface{}, interrupt chan os.Signal) ws {
	conn, _, err := websocket.DefaultDialer.Dial(websocketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveHandler(conn, done)

	err = conn.WriteJSON(twitchWSOutgoingMessage{Type: "LISTEN", Nonce: "twitchPubSub", Data: authMessageData{AuthToken: oauthToken, Topics: topics}})
	if err != nil {
		log.Println("Error during LISTEN to websocket:", err)
		return ws{}
	}

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		case <-time.After(time.Duration(5) * time.Millisecond * 1000 * 60):
			// Send an echo packet every second
			err := conn.WriteJSON(twitchWSOutgoingMessage{Type: "PING"})
			if err != nil {
				log.Println("Error during writing to websocket:", err)
				return ws{}
			}

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return ws{}
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return ws{}
		}
	}
	return ws{
		conn,
		done,
		interrupt,
	}
}

func (w ws) SendMessage(message string) error {
	err := w.conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println("Error during writing to websocket:", err)
		return err
	}
	return nil
}

func receiveHandler(conn *websocket.Conn, done chan interface{}) {
	defer close(done)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}
