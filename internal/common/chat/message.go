package chat

type Message struct {
	Data    interface{} `json:"data"`
	FromUID string      `json:"from_uid"`
	Time    string      `json:"time"`
}
