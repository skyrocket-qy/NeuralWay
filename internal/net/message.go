package net

import (
	"encoding/binary"
	"errors"
)

// MessageType identifies the type of network message.
type MessageType uint8

const (
	// MsgConnect is sent when a client connects.
	MsgConnect MessageType = iota
	// MsgDisconnect is sent when a client disconnects.
	MsgDisconnect
	// MsgStateUpdate contains game state data.
	MsgStateUpdate
	// MsgStateDelta contains delta-compressed state.
	MsgStateDelta
	// MsgInput contains player input.
	MsgInput
	// MsgRPC is a remote procedure call.
	MsgRPC
	// MsgPing is for latency measurement.
	MsgPing
	// MsgPong is the response to ping.
	MsgPong
	// MsgAck acknowledges receipt of a message.
	MsgAck
)

// Message represents a network message.
type Message struct {
	Type     MessageType
	Tick     int64
	ClientID uint32
	Sequence uint32
	Payload  []byte
}

// Header size: Type(1) + Tick(8) + ClientID(4) + Sequence(4) + PayloadLen(4) = 21 bytes.
const headerSize = 21

// Encode serializes a message to bytes.
func Encode(msg *Message) []byte {
	data := make([]byte, headerSize+len(msg.Payload))

	data[0] = byte(msg.Type)
	binary.LittleEndian.PutUint64(data[1:9], uint64(msg.Tick))
	binary.LittleEndian.PutUint32(data[9:13], msg.ClientID)
	binary.LittleEndian.PutUint32(data[13:17], msg.Sequence)
	binary.LittleEndian.PutUint32(data[17:21], uint32(len(msg.Payload)))
	copy(data[21:], msg.Payload)

	return data
}

// Decode deserializes bytes to a message.
func Decode(data []byte) (*Message, error) {
	if len(data) < headerSize {
		return nil, errors.New("message too short")
	}

	payloadLen := binary.LittleEndian.Uint32(data[17:21])
	if len(data) < headerSize+int(payloadLen) {
		return nil, errors.New("payload incomplete")
	}

	msg := &Message{
		Type:     MessageType(data[0]),
		Tick:     int64(binary.LittleEndian.Uint64(data[1:9])),
		ClientID: binary.LittleEndian.Uint32(data[9:13]),
		Sequence: binary.LittleEndian.Uint32(data[13:17]),
		Payload:  make([]byte, payloadLen),
	}
	copy(msg.Payload, data[21:21+payloadLen])

	return msg, nil
}

// NewConnectMessage creates a connect message.
func NewConnectMessage() *Message {
	return &Message{Type: MsgConnect}
}

// NewDisconnectMessage creates a disconnect message.
func NewDisconnectMessage(clientID uint32) *Message {
	return &Message{Type: MsgDisconnect, ClientID: clientID}
}

// NewStateMessage creates a state update message.
func NewStateMessage(tick int64, state []byte) *Message {
	return &Message{Type: MsgStateUpdate, Tick: tick, Payload: state}
}

// NewInputMessage creates an input message.
func NewInputMessage(tick int64, clientID uint32, input []byte) *Message {
	return &Message{Type: MsgInput, Tick: tick, ClientID: clientID, Payload: input}
}

// NewRPCMessage creates an RPC message.
func NewRPCMessage(method string, args []byte) *Message {
	// Format: methodLen(1) + method + args
	payload := make([]byte, 1+len(method)+len(args))
	payload[0] = byte(len(method))
	copy(payload[1:], method)
	copy(payload[1+len(method):], args)

	return &Message{Type: MsgRPC, Payload: payload}
}

// ParseRPC extracts method name and args from an RPC message.
func ParseRPC(msg *Message) (string, []byte, error) {
	if msg.Type != MsgRPC {
		return "", nil, errors.New("not an RPC message")
	}

	if len(msg.Payload) < 1 {
		return "", nil, errors.New("invalid RPC payload")
	}

	methodLen := int(msg.Payload[0])
	if len(msg.Payload) < 1+methodLen {
		return "", nil, errors.New("invalid RPC method")
	}

	method := string(msg.Payload[1 : 1+methodLen])
	args := msg.Payload[1+methodLen:]

	return method, args, nil
}

// NewPingMessage creates a ping message.
func NewPingMessage(sequence uint32) *Message {
	return &Message{Type: MsgPing, Sequence: sequence}
}

// NewPongMessage creates a pong response.
func NewPongMessage(sequence uint32) *Message {
	return &Message{Type: MsgPong, Sequence: sequence}
}
