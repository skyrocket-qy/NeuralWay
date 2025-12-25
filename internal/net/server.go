package net

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// ClientConn represents a connected client.
type ClientConn struct {
	ID       uint32
	conn     *websocket.Conn
	server   *NetServer
	send     chan *Message
	isClosed bool
	mu       sync.RWMutex
}

// NetServer is a WebSocket game server.
type NetServer struct {
	clients      map[uint32]*ClientConn
	nextClientID uint32
	onConnect    func(clientID uint32)
	onDisconnect func(clientID uint32)
	onMessage    func(clientID uint32, msg *Message)
	upgrader     websocket.Upgrader
	mu           sync.RWMutex
}

// NewNetServer creates a new network server.
func NewNetServer() *NetServer {
	return &NetServer{
		clients: make(map[uint32]*ClientConn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// OnConnect sets the callback for new connections.
func (s *NetServer) OnConnect(handler func(clientID uint32)) {
	s.onConnect = handler
}

// OnDisconnect sets the callback for disconnections.
func (s *NetServer) OnDisconnect(handler func(clientID uint32)) {
	s.onDisconnect = handler
}

// OnMessage sets the callback for received messages.
func (s *NetServer) OnMessage(handler func(clientID uint32, msg *Message)) {
	s.onMessage = handler
}

// Start begins listening for WebSocket connections.
func (s *NetServer) Start(addr string) error {
	http.HandleFunc("/ws", s.handleWebSocket)
	log.Printf("NetServer starting on %s", addr)

	return http.ListenAndServe(addr, nil)
}

// handleWebSocket handles new WebSocket connections.
func (s *NetServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)

		return
	}

	clientID := atomic.AddUint32(&s.nextClientID, 1)
	client := &ClientConn{
		ID:     clientID,
		conn:   conn,
		server: s,
		send:   make(chan *Message, 100),
	}

	s.mu.Lock()
	s.clients[clientID] = client
	s.mu.Unlock()

	// Send client their ID
	connectMsg := &Message{Type: MsgConnect, ClientID: clientID}
	client.send <- connectMsg

	if s.onConnect != nil {
		s.onConnect(clientID)
	}

	go client.readPump()
	go client.writePump()
}

// Broadcast sends a message to all connected clients.
func (s *NetServer) Broadcast(msg *Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		client.Send(msg)
	}
}

// SendTo sends a message to a specific client.
func (s *NetServer) SendTo(clientID uint32, msg *Message) {
	s.mu.RLock()
	client, ok := s.clients[clientID]
	s.mu.RUnlock()

	if ok {
		client.Send(msg)
	}
}

// GetClientCount returns the number of connected clients.
func (s *NetServer) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.clients)
}

// GetClientIDs returns all connected client IDs.
func (s *NetServer) GetClientIDs() []uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]uint32, 0, len(s.clients))
	for id := range s.clients {
		ids = append(ids, id)
	}

	return ids
}

// DisconnectClient disconnects a specific client.
func (s *NetServer) DisconnectClient(clientID uint32) {
	s.mu.Lock()

	client, ok := s.clients[clientID]
	if ok {
		delete(s.clients, clientID)
	}

	s.mu.Unlock()

	if ok {
		client.Close()

		if s.onDisconnect != nil {
			s.onDisconnect(clientID)
		}
	}
}

// Send queues a message to be sent to this client.
func (c *ClientConn) Send(msg *Message) {
	c.mu.RLock()
	closed := c.isClosed
	c.mu.RUnlock()

	if closed {
		return
	}

	select {
	case c.send <- msg:
	default:
		// Queue full, drop message
	}
}

// Close closes the client connection.
func (c *ClientConn) Close() {
	c.mu.Lock()

	if c.isClosed {
		c.mu.Unlock()

		return
	}

	c.isClosed = true
	c.mu.Unlock()

	close(c.send)
	c.conn.Close()
}

// readPump reads messages from the client.
func (c *ClientConn) readPump() {
	defer func() {
		c.server.DisconnectClient(c.ID)
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		msg, err := Decode(data)
		if err != nil {
			continue
		}

		msg.ClientID = c.ID

		// Handle ping
		if msg.Type == MsgPing {
			c.Send(NewPongMessage(msg.Sequence))

			continue
		}

		if c.server.onMessage != nil {
			c.server.onMessage(c.ID, msg)
		}
	}
}

// writePump sends messages to the client.
func (c *ClientConn) writePump() {
	for msg := range c.send {
		data := Encode(msg)
		if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			return
		}
	}
}
