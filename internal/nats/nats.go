package nats

import (
	"fmt"
	"universe-chat/internal/config"

	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	conn      *nats.Conn
	jetStream nats.JetStreamContext
}

func NewClient(config *config.NatsConfig) (*NatsClient, error) {
	natsClient := &NatsClient{}
	conn, err := nats.Connect(fmt.Sprintf("%v:%v", config.Host, config.Port))
	if err != nil {
		return nil, err
	}

	context, err := conn.JetStream()
	if err != nil {
		return nil, err
	}

	natsClient.conn = conn
	natsClient.jetStream = context
	natsClient.jetStream.AddStream(&nats.StreamConfig{Name: "MSET", Storage: nats.FileStorage, Subjects: []string{"chat.message.*"}})
	return natsClient, nil
}

func (client *NatsClient) AddStream(streamConfig *nats.StreamConfig) error {
	_, err := client.jetStream.AddStream(streamConfig)
	if err != nil {
		return err
	}

	return nil
}

func (client *NatsClient) Subscribe(channel string, f func(seq uint64, data []byte)) (interface{}, error) {

	subs, err := client.jetStream.Subscribe(fmt.Sprintf("chat.message.%s", channel), func(msg *nats.Msg) {
		meta, _ := msg.Metadata()
		msg.Ack()
		f(meta.Sequence.Stream, msg.Data)
	})
	return subs, err
}

func (client *NatsClient) Send(channel string, data []byte) error {
	_, err := client.jetStream.Publish(fmt.Sprintf("chat.message.%s", channel), data)
	return err
}
