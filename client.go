package socketon

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

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

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	manager *manager

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// actions
	actions map[string]func(c *Client, data interface{})
}

func NewClient(s *manager, c *websocket.Conn) *Client {
	client := &Client{
		manager: s,
		conn:    c,
		send:    make(chan []byte),
		actions: make(map[string]func(c *Client, data interface{})),
	}

	client.manager.register <- client
	go client.read()
	go client.writer()

	return client
}

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

		data := new(Message)
		if err := json.Unmarshal(message, data); err != nil {
			if action, ok := c.actions[data.Action]; ok {
				action(c, data.Data)
			}
		}
	}
}

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

func (c *Client) On(name string, action func(c *Client, data interface{})) {
	c.actions[name] = action
}

func (c *Client) Emit(to *Client, data Message) {
	if _, ok := c.manager.clients[to]; ok {
		dataSend, err := json.Marshal(data)
		if err == nil {
			to.send <- dataSend
		}
	}
}

func (c *Client) Broadcast(data Message) {
	dataSend, err := json.Marshal(data)
	if err == nil {
		for to := range c.manager.clients {
			if to != c {
				to.send <- dataSend
			}
		}
	}
}
