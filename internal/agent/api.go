package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type API struct {
	broker   Broker
	upgrader websocket.Upgrader
}

type connectionInfo struct {
	Channel string  `json:"channel"`
	UID     string  `json:"uid"`
	Secret  string  `json:"secret"` // User secret
	LastSeq *uint64 `json:"last_seq,omitempty"`
}

func New(route gin.IRouter, broker Broker) *API {
	api := &API{
		broker: broker,
		upgrader: websocket.Upgrader{
			CheckOrigin:     func(r *http.Request) bool { return true },
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
	route.Handle("GET", "/connect", api.connect)
	return api
}

// TODO this function has to have infinete loop for ws
func (api *API) connect(c *gin.Context) {
	w, r := c.Writer, c.Request

	con, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("error while upgrading to ws connection: %v", err), 500)
		return
	}
	connectionInfo := api.getConnectionInfo(con)
	agent := &Agent{
		conn:   con,
		broker: api.broker,
	}
	agent.startListening(con, connectionInfo)
}

// TODO this2 function has to have infinete loop for ws
func (api *API) getConnectionInfo(con *websocket.Conn) *connectionInfo {
	t, r, err := con.NextReader()
	if err != nil || t == websocket.CloseMessage {
		con.WriteJSON(&Message{Type: errorMessage, Error: "connection close"})
	}
	var info *connectionInfo
	err = json.NewDecoder(r).Decode(&info)
	if err != nil {
		con.WriteJSON(&Message{Type: errorMessage, Error: err.Error()})
	}
	return info
}
