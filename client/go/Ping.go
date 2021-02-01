package client

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Ping the pipeline server on its echo channel
// Establish a connection and send a ping message on /ping
func Ping(host string) bool {
	pingPath := fmt.Sprintf("ws://%s/echo", host)
	logrus.Printf("connecting to %s", pingPath)
	var pingMsg = "ping"

	connection, resp, err := websocket.DefaultDialer.Dial(pingPath, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed sending ping: %s", err)
		if resp != nil {
			msg = fmt.Sprintf("%s: %s (%d)", err, resp.Status, resp.StatusCode)
		}
		logrus.Fatal("Ping: fatal error: ", msg)
		return false
	}
	defer connection.Close()
	err = connection.WriteMessage(websocket.TextMessage, []byte(pingMsg))
	if err != nil {
		logrus.Errorf("Ping: Error writing message: %s", err)
		return false
	}
	_, message, err := connection.ReadMessage()
	if err != nil {
		logrus.Errorf("Ping: Failed reading response from connection: %s", err)
		return false
	}
	logrus.Infof("Ping: Response: %s", message)
	if string(message) != pingMsg {
		logrus.Errorf("Ping: Expected response '%s'. Got '%s' instead.", pingMsg, message)
		return false
	}

	// time.Sleep(time.Second * 3)
	// connection.Close()
	return true
}
