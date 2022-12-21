package agent

import (
	"encoding/json"
	"fmt"
	"time"
	"universe-chat/internal/common/chat"

	"github.com/gorilla/websocket"
)

type MessageType int

type Message struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
	Type  MessageType `json:"type,omitempty"`
}

type Unsibscriber interface {
	Unsubscribe() error
}

type Broker interface {
	Subscribe(channel string, userid string, message chan *chat.Message) (func(), error)
	Send(channel string, message *chat.Message) error
}

type Agent struct {
	uns     func()
	uid     string
	broker  Broker
	conn    *websocket.Conn
	channel string
	closed  bool
	done    chan struct{}
}

const (
	chatMessage MessageType = iota
	errorMessage
	historyMessage
)

func (agent *Agent) startListening(conn *websocket.Conn, req *connectionInfo) {
	agent.conn = conn
	agent.channel = req.Channel
	agent.uid = req.UID
	agent.conn.SetCloseHandler(func(code int, test string) error {
		agent.closed = true
		agent.done <- struct{}{}
		return nil
	})

	m := make(chan *chat.Message)
	{
		unsibscriber, err := agent.broker.Subscribe(req.Channel, req.UID, m)
		if err != nil {
			conn.WriteJSON(&Message{Type: errorMessage, Error: fmt.Errorf("error during subscribe to channel %w", err).Error()})
			return
		}

		agent.uns = unsibscriber

	}
	agent.loop(m)
}

func (agent *Agent) loop(mc chan *chat.Message) {
	go func() {
		defer agent.uns()
		defer agent.conn.Close()

		for {
			select {
			case message := <-mc:
				{
					if message.FromUID != agent.uid {
						err := agent.conn.WriteJSON(&Message{Data: message.Data, Type: chatMessage})
						if err != nil {
							fmt.Println(err)
						}
					}
				}
			case <-agent.done:
				return
			default:
				time.After(1 * time.Second)
			}

		}

	}()

	go func() {
		for {
			_, r, err := agent.conn.NextReader()
			if err != nil {
				agent.conn.WriteJSON(&Message{Type: errorMessage})
				return
			}

			var msg chat.Message
			json.NewDecoder(r).Decode(&msg)

			err = agent.broker.Send(agent.channel, &chat.Message{Data: msg.Data, FromUID: msg.FromUID})
			if err != nil {
				return
			}
		}

	}()

}
