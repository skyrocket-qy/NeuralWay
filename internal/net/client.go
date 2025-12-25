package net

import (
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// NetClient is a WebSocket client for connecting to a game server.
type NetClient struct {
	conn          *websocket.Conn
	serverAddr    string
	clientID      uint32
	sequence      uint32
	connected     bool
	onMessage     func(*Message)
	onConnect     func()
	onDisconnect  func()
	sendQueue     chan *Message
	mu            sync.RWMutex
	reconnect     bool
	reconnectWait time.Duration
}

// NewNetClient creates a new network client.
func NewNetClient() *NetClient {
	return &NetClient{
		sendQueue:     make(chan *Message, 100),
		reconnect:     true,
		reconnectWait: 2 * time.Second,
	}
}

// OnMessage sets the callback for received messages.
func (c *NetClient) OnMessage(handler func(*Message)) {
	c.onMessage = handler
}

// OnConnect sets the callback for connection established.
func (c *NetClient) OnConnect(handler func()) {
	c.onConnect = handler
}

// OnDisconnect sets the callback for connection lost.
func (c *NetClient) OnDisconnect(handler func()) {
	c.onDisconnect = handler
}

// Connect establishes a connection to the server.
func (c *NetClient) Connect(addr string) error {
	c.serverAddr = addr

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.mu.Unlock()

	// Send connect message
	c.Send(NewConnectMessage())

	// Start read/write goroutines
	go c.readLoop()
	go c.writeLoop()

	if c.onConnect != nil {
		c.onConnect()
	}

	return nil
}

// IsConnected returns true if connected to server.
func (c *NetClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.connected
}

// GetClientID returns the assigned client ID.
func (c *NetClient) GetClientID() uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.clientID
}

// Send queues a message to be sent to the server.
func (c *NetClient) Send(msg *Message) error {
	if !c.IsConnected() {
		return errors.New("not connected")
	}

	c.mu.Lock()
	msg.Sequence = c.sequence
	c.sequence++
	c.mu.Unlock()

	select {
	case c.sendQueue <- msg:
		return nil
	default:
		return errors.New("send queue full")
	}
}

// SendInput sends player input to the server.
func (c *NetClient) SendInput(tick int64, input []byte) error {
	return c.Send(NewInputMessage(tick, c.clientID, input))
}

// SendRPC sends a remote procedure call.
func (c *NetClient) SendRPC(method string, args []byte) error {
	return c.Send(NewRPCMessage(method, args))
}

// Disconnect closes the connection.
func (c *NetClient) Disconnect() {
	c.mu.Lock()
	c.reconnect = false

	c.connected = false
	if c.conn != nil {
		c.conn.Close()
	}

	c.mu.Unlock()

	if c.onDisconnect != nil {
		c.onDisconnect()
	}
}

// readLoop reads messages from the server.
func (c *NetClient) readLoop() {
	for {
		c.mu.RLock()
		conn := c.conn
		connected := c.connected
		c.mu.RUnlock()

		if !connected || conn == nil {
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			c.handleDisconnect()

			return
		}

		msg, err := Decode(data)
		if err != nil {
			continue
		}

		// Handle special messages
		switch msg.Type {
		case MsgConnect:
			c.mu.Lock()
			c.clientID = msg.ClientID
			c.mu.Unlock()
		case MsgPing:
			c.Send(NewPongMessage(msg.Sequence))
		}

		if c.onMessage != nil {
			c.onMessage(msg)
		}
	}
}

// writeLoop sends queued messages to the server.
func (c *NetClient) writeLoop() {
	for msg := range c.sendQueue {
		c.mu.RLock()
		conn := c.conn
		connected := c.connected
		c.mu.RUnlock()

		if !connected || conn == nil {
			return
		}

		data := Encode(msg)
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			c.handleDisconnect()

			return
		}
	}
}

// handleDisconnect handles connection loss.
func (c *NetClient) handleDisconnect() {
	c.mu.Lock()
	wasConnected := c.connected
	c.connected = false
	shouldReconnect := c.reconnect
	addr := c.serverAddr
	c.mu.Unlock()

	if wasConnected && c.onDisconnect != nil {
		c.onDisconnect()
	}

	// Auto-reconnect
	if shouldReconnect && addr != "" {
		go func() {
			time.Sleep(c.reconnectWait)
			c.Connect(addr)
		}()
	}
}
