package servicebus

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// ServeEcho response with an echo of the request
func ServeEcho(response http.ResponseWriter, request *http.Request) {
	// fmt.Printf("Hello world from %s", r.URL.Path[1:])
	// serveWs(hub, w, r)
	logrus.Info("ServeEcho starting")
	c, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		logrus.Printf("ServeEcho upgrade error : %s", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("ServeEcho read error on %s: %s", request.URL, err)
			}
			logrus.Warningf("ServeEcho connection closed: %s", request.URL)
			break
		}
		logrus.Printf("ServeEcho received message: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			logrus.Errorf("ServeEcho write error: %s", err)
			break
		}
	}
}
