package servicebus

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var channels = NewChannelPlumbing()
var upgrader = websocket.Upgrader{} // use default options
const defaultBufferSize = 10
const isPub = "pub"
const isSub = "sub"

// ServeChannel handles new channel connections of capture, processor and consumer clients
func ServeChannel(response http.ResponseWriter, request *http.Request) {
	chID := mux.Vars(request)["ChannelID"]
	pubOrSub := mux.Vars(request)["Stage"]
	logrus.Infof("ServeChannel starting for channel %s/%s", chID, pubOrSub)

	// upgrade the HTTP connection to a websocket connection
	c, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		logrus.Printf("ServeChannel upgrade error : %s", err)
		return
	}

	channel := channels.GetChannel(chID)
	if channel == nil {
		// Create a new channel with worker to process messages
		channel = channels.NewChannel(chID, defaultBufferSize)
		go channelMessageWorker(channel)
	}
	// Keep listening for messages from publishers
	if pubOrSub == isPub {
		// capture connections are closed on exit
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Errorf("ServeChannel on %s/%s read error: %s", chID, pubOrSub, err)
				}
				channels.RemoveConnection(c)
				break
			}
			logrus.Infof("ServeChannel received capture message on channel ID %s", chID)
			channel.ChannelGoChan <- message
		}
	}
}

// read a message from the channel and send it to subscribers
func channelMessageWorker(channel *Channel) {
	for message := range channel.ChannelGoChan {
		ConsumeChannelMessage(channel.ID, message)
	}
}

// ConsumeChannelMessage passes a message to all subscribers
// If a subscriber fails, it is removed from the channel
func ConsumeChannelMessage(channelID string, message []byte) {
	consumers := channels.GetSubscribers(channelID)
	logrus.Infof("Sending message to %d consumers of channel %s", len(consumers), channelID)
	for _, c := range consumers {
		err := SendMessage(c, message)
		if err != nil {
			channels.RemoveConnection(c)
		}
	}
}

// ReceiveMessage from a socket connection
// Wait until message received or timeoutSec has passed, use 0 to wait indefinitely
// Note that if the connection times out, the connectino must be discarded.
func ReceiveMessage(connection *websocket.Conn, timeout time.Duration) ([]byte, error) {
	connection.SetReadDeadline(time.Now().Add(timeout))
	_, message, err := connection.ReadMessage()
	return message, err
}

// SendMessage publishes a message into the channel
func SendMessage(connection *websocket.Conn, message []byte) error {
	w, err := connection.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	w.Write(message)
	w.Close()
	return nil
}
