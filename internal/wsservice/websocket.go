package wsservice

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type Ws interface {
	SendMessage(message string) error
}

type ws struct {
	conn *websocket.Conn
}

func NewWebsocket(conn *websocket.Conn) ws {
	return ws{
		conn: conn,
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

var done chan interface{}    // Channel to indicate that the receiverHandler is done
var interrupt chan os.Signal // Channel to listen for interrupt signal to terminate gracefully

func receiveHandler(conn *websocket.Conn) {
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

func main(websocketUrl string) {

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT
	done = make(chan interface{})
	interrupt = make(chan os.Signal)

	conn, _, err := websocket.DefaultDialer.Dial(websocketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		case <-time.After(time.Duration(5) * time.Millisecond * 1000):
			// Send an echo packet every second
			err := conn.WriteMessage(websocket.TextMessage, []byte("PING"))
			if err != nil {
				log.Println("Error during writing to websocket:", err)
				return
			}

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}
}
