package gateway

import "github.com/gorilla/websocket"

// Predefined gateway channels
const (
	// The TD channel carries 'Thing Description' documents
	TDChannelID = "TD"
	// The notification channel carries Thing status updates
	NotificationChannelID = "notification"
	// The action channel carries Thing action commands
	ActionChannelID = "action"
	// The plugin channel carries plugin registration messages
	PluginChannelID = "plugin"
	// The test channel carries test messages
	TestChannelID = "test"
)

// Publish a message over the websocket connection to a channel topic
func Publish(conn *websocket.Conn, channel string, message []byte) error {
	// message is simply encoded with topic:message
	data := channel + ":" + string(message)
	err := conn.WriteMessage(websocket.TextMessage, []byte(data))
	return err
}
