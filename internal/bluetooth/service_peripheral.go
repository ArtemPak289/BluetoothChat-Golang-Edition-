//go:build linux || windows

package bluetooth

import (
	"fmt"
	"sync"

	bt "tinygo.org/x/bluetooth"
)

// ChatService hosts the chat GATT service as a BLE peripheral.
type ChatService struct {
	adapter *bt.Adapter

	mu      sync.Mutex
	peerKey string
	framer  FrameEncoder

	rxChar *bt.Characteristic
	txChar *bt.Characteristic

	onPayload func(peerKey string, payload []byte)
}

// NewChatService creates an unregistered GATT chat service.
func NewChatService(adapter *bt.Adapter, onPayload func(peerKey string, payload []byte)) *ChatService {
	return &ChatService{
		adapter:   adapter,
		onPayload: onPayload,
	}
}

// Register adds the GATT service to the adapter.
func (s *ChatService) Register() error {
	var rxChar, txChar bt.Characteristic
	err := s.adapter.AddService(&bt.Service{
		UUID: ChatServiceUUID,
		Characteristics: []bt.CharacteristicConfig{
			{
				Handle: &rxChar,
				UUID:   ChatRXUUID,
				Flags:  bt.CharacteristicWritePermission | bt.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(_ bt.Connection, _ int, value []byte) {
					s.mu.Lock()
					key := s.peerKey
					if key == "" {
						key = "inbound"
					}
					s.mu.Unlock()
					frames, err := s.framer.Append(value)
					if err != nil {
						return
					}
					for _, frame := range frames {
						if s.onPayload != nil {
							s.onPayload(key, frame)
						}
					}
				},
			},
			{
				Handle: &txChar,
				UUID:   ChatTXUUID,
				Flags:  bt.CharacteristicNotifyPermission | bt.CharacteristicReadPermission,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("add service: %w", err)
	}
	s.rxChar = &rxChar
	s.txChar = &txChar
	return nil
}

// SetPeerKey sets the logical peer key for the current inbound central connection.
func (s *ChatService) SetPeerKey(key string) {
	s.mu.Lock()
	s.peerKey = key
	s.framer = FrameEncoder{}
	s.mu.Unlock()
}

// ClearPeerKey resets inbound peer state after disconnect.
func (s *ChatService) ClearPeerKey(key string) {
	s.mu.Lock()
	if s.peerKey == key {
		s.peerKey = ""
		s.framer = FrameEncoder{}
	}
	s.mu.Unlock()
}

// Send notifies connected centrals with a framed payload.
func (s *ChatService) Send(data []byte) error {
	if s.txChar == nil {
		return fmt.Errorf("gatt service not registered")
	}
	frame, err := EncodeFrame(data)
	if err != nil {
		return err
	}
	return WriteChunks(frame, 20, func(part []byte) error {
		_, err := s.txChar.Write(part)
		return err
	})
}
