package socketon

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Constants for managing WebSocket communication.
const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Common byte slices.
var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Upgrader is a WebSocket upgrader with specified buffer sizes.
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a WebSocket client.
type Client struct {
	manager *manager                                     // Reference to the manager.
	conn    *websocket.Conn                              // The WebSocket connection.
	send    chan []byte                                  // Buffered channel for outbound messages.
	actions map[string]func(c *Client, data interface{}) // Map of registered actions.
	Log     bool                                         // Indicates whether logging is enabled for this client.
}

// NewClient creates a new Client instance and starts reading and writing goroutines.
func NewClient(s *manager, c *websocket.Conn) *Client {
	client := &Client{
		manager: s,
		conn:    c,
		send:    make(chan []byte),
		actions: make(map[string]func(c *Client, data interface{})),
		Log:     false,
	}

	client.manager.register <- client
	go client.read()
	go client.writer()

	return client
}

// read reads incoming messages from the WebSocket connection.
func (c *Client) read() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		if c.Log {
			log.Println("new message", message)
		}

		data := new(Message)
		err = json.Unmarshal(message, data)
		if err != nil && c.Log {
			log.Println("new message", message)
			return
		}

		if action, ok := c.actions[data.Action]; ok {
			action(c, data.Data)
		}
	}
}

// writer handles outgoing messages to the WebSocket connection.
func (c *Client) writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

// On registers an action with a given name and corresponding function.
func (c *Client) On(name string, action func(c *Client, data interface{})) {
	c.actions[name] = action
}

// Emit sends a message to a specific client.
func (c *Client) Emit(to *Client, data *Message) {
	if _, ok := c.manager.clients[to]; ok {
		dataSend, err := json.Marshal(data)
		if err == nil {
			to.send <- dataSend
		}
	}
}

// Broadcast sends a message to all connected clients except the sender.
func (c *Client) Broadcast(data *Message) {
	dataSend, err := json.Marshal(data)
	if err == nil {
		for to := range c.manager.clients {
			if to != c {
				to.send <- dataSend
			}
		}
	}
}
