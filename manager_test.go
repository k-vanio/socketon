package socketon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	s := NewManager()

	assert.NotNil(t, s.broadcast)
	assert.NotNil(t, s.clients)
	assert.NotNil(t, s.register)
	assert.NotNil(t, s.unregister)
}

func TestServerStartAndStop(t *testing.T) {
	manager := NewManager()

	go func() {
		time.Sleep(10 * time.Millisecond)
		manager.Stop()
	}()

	manager.Start()
	assert.Equal(t, false, <-manager.stop)
}
