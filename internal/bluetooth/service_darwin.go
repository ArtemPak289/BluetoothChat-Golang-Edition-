//go:build darwin

package bluetooth

import (
	"errors"

	bt "tinygo.org/x/bluetooth"
)

// ErrPeripheralUnsupported is returned when GATT peripheral mode is unavailable.
var ErrPeripheralUnsupported = errors.New("ble peripheral mode is not supported on macOS in tinygo-org/bluetooth; use Linux or Windows to advertise, or connect outbound from this machine")

// ChatService is a stub on platforms without peripheral GATT support.
type ChatService struct{}

// NewChatService creates a no-op chat service placeholder.
func NewChatService(_ *bt.Adapter, _ func(peerKey string, payload []byte)) *ChatService {
	return &ChatService{}
}

// Register reports that peripheral mode is unavailable.
func (s *ChatService) Register() error {
	return ErrPeripheralUnsupported
}

// SetPeerKey is a no-op on unsupported platforms.
func (s *ChatService) SetPeerKey(string) {}

// ClearPeerKey is a no-op on unsupported platforms.
func (s *ChatService) ClearPeerKey(string) {}

// Send reports that peripheral mode is unavailable.
func (s *ChatService) Send([]byte) error {
	return ErrPeripheralUnsupported
}
