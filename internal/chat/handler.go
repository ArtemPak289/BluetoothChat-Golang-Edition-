package chat

import (
	"fmt"
	"sync"
)

// IncomingMessage is a decoded chat message tagged with the originating peer key.
type IncomingMessage struct {
	Peer    string
	Message Message
}

// Handler manages peer transports and dispatches inbound chat messages.
type Handler struct {
	sender string

	mu       sync.RWMutex
	peers    map[string]Transport
	incoming chan IncomingMessage
}

// NewHandler creates a chat handler for the given local display name.
func NewHandler(sender string) *Handler {
	return &Handler{
		sender:   sender,
		peers:    make(map[string]Transport),
		incoming: make(chan IncomingMessage, 32),
	}
}

// Incoming returns a channel of messages received from any peer.
func (h *Handler) Incoming() <-chan IncomingMessage {
	return h.incoming
}

// RegisterPeer associates a peer key with its transport for outbound messages.
func (h *Handler) RegisterPeer(key string, sender Transport) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.peers[key] = sender
}

// UnregisterPeer removes a peer from the outbound registry.
func (h *Handler) UnregisterPeer(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.peers, key)
}

// ConnectedPeers returns the keys of currently registered peers.
func (h *Handler) ConnectedPeers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]string, 0, len(h.peers))
	for k := range h.peers {
		out = append(out, k)
	}
	return out
}

// SendText encodes and delivers a message to a single peer.
func (h *Handler) SendText(peerKey, text string) error {
	h.mu.RLock()
	peer, ok := h.peers[peerKey]
	h.mu.RUnlock()
	if !ok {
		return fmt.Errorf("peer %q is not connected", peerKey)
	}
	msg := NewMessage(h.sender, text)
	data, err := msg.Encode()
	if err != nil {
		return err
	}
	return peer.Send(data)
}

// BroadcastText sends the same message to every connected peer.
func (h *Handler) BroadcastText(text string) error {
	h.mu.RLock()
	keys := make([]string, 0, len(h.peers))
	for k := range h.peers {
		keys = append(keys, k)
	}
	h.mu.RUnlock()

	var firstErr error
	for _, k := range keys {
		if err := h.SendText(k, text); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if len(keys) == 0 {
		return fmt.Errorf("no connected peers")
	}
	return firstErr
}

// HandlePayload decodes an inbound payload and pushes it to the incoming channel.
func (h *Handler) HandlePayload(peerKey string, data []byte) error {
	msg, err := DecodeMessage(data)
	if err != nil {
		return err
	}
	select {
	case h.incoming <- IncomingMessage{Peer: peerKey, Message: msg}:
	default:
		return fmt.Errorf("incoming queue full, drop message from %q", peerKey)
	}
	return nil
}
