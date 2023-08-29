package socketon

// Message represents a message that will be transmitted through a socket communication system.
type Message struct {
	Action string      `json:"a"` // Action associated with the message.
	Data   interface{} `json:"d"` // Data associated with the message.
}

// NewMessage is a constructor function that creates and returns a new instance of the Message structure.
// It accepts an action and data as arguments and returns a pointer to the new message.
func NewMessage(Action string, Data interface{}) *Message {
	return &Message{
		Action: Action,
		Data:   Data,
	}
}
