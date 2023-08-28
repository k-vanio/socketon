package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	manager := NewManager()
	go manager.Start()
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}

		client := NewClient(manager, conn)
		assert.NotNil(t, client)
		assert.NotNil(t, client.manager)
		assert.NotNil(t, client.conn)
		assert.NotNil(t, client.send)
		assert.NotNil(t, client.actions)

	}))
	defer testServer.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+testServer.URL[4:], nil)
	assert.Nil(t, err)
	defer conn.Close()

	<-time.After(3 * time.Millisecond)
}
