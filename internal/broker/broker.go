package broker

import (
	"encoding/json"
	"fmt"
	"universe-chat/internal/common"
	"universe-chat/internal/common/chat"
)

type Worker interface {
	Subscribe(channel string, f func(seq uint64, data []byte)) (interface{}, error)
	Send(channel string, data []byte) error
}

type MessageBroker struct {
	worker Worker
}

func New(worker Worker) *MessageBroker {
	return &MessageBroker{worker}
}

func (broker *MessageBroker) Subscribe(channel string, userid string, incomeMessage chan *chat.Message) (func(), error) {
	unsubscriber, err := broker.worker.Subscribe(channel, func(seq uint64, data []byte) {
		var m interface{}

		err := json.Unmarshal(data, &m)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(m)
		var message *chat.Message
		err = json.Unmarshal(data, &message)
		if err != nil {
			fmt.Println(err)
			return
		}
		incomeMessage <- message

	})

	unsun := func() {
		fmt.Println("subsribstion closed")
		un := unsubscriber.(common.Unsibscriber)
		un.Unsubscribe()
	}
	return unsun, err
}

func (broker *MessageBroker) Send(channel string, message *chat.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return broker.worker.Send(channel, data)
}
