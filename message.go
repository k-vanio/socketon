package websocket

type Message struct {
	Action string      `json:"e"`
	Data   interface{} `json:"d"`
}

func NewMessage(Action string, Data interface{}) *Message {
	return &Message{
		Action: Action,
		Data:   Data,
	}
}
