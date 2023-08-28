package socketon

type manager struct {
	broadcast chan []byte

	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	stop chan bool
}

func NewManager() *manager {
	return &manager{
		broadcast:  make(chan []byte),
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan bool),
	}
}

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

func (m *manager) Stop() {
	m.stop <- true
}
