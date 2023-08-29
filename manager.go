package socketon

// manager is responsible for managing client connections and broadcasting messages.
type manager struct {
	broadcast  chan []byte      // A channel for broadcasting messages to all clients.
	clients    map[*Client]bool // A map of registered clients.
	register   chan *Client     // A channel for registering clients.
	unregister chan *Client     // A channel for unregister clients.
	stop       chan bool        // A channel to stop the manager's operation.
}

// NewManager creates and returns a new instance of the manager.
func NewManager() *manager {
	return &manager{
		broadcast:  make(chan []byte),
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan bool),
	}
}

// Start starts the manager's main loop for handling client registration, unregister, and broadcasting.
func (m *manager) Start() {
	defer close(m.stop)

end:
	for {
		select {
		case client := <-m.register:
			m.clients[client] = true
		case client := <-m.unregister:
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
			}
		case <-m.stop:
			break end
		}
	}
}

// Stop stops the manager's operation and closes its channels.
func (m *manager) Stop() {
	m.stop <- true
}
