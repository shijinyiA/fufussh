package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"web-ssh/models"

	"github.com/gorilla/websocket"
)

var defaultUpgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	HandshakeTimeout: 10 * time.Second,
}

func getUpgrader() *websocket.Upgrader {
	return defaultUpgrader
}

func writeWSMessage(ws *websocket.Conn, msg models.WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return ws.WriteMessage(websocket.TextMessage, data)
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
