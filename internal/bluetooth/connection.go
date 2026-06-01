package bluetooth

import (
	"fmt"
	"sync"

	bt "tinygo.org/x/bluetooth"
)

// OnMessageFunc is invoked when a complete framed payload arrives from a peer.
type OnMessageFunc func(peerKey string, payload []byte)

// CentralLink sends and receives length-prefixed chat frames as a GATT central.
type CentralLink struct {
	key    string
	device bt.Device
	rx     bt.DeviceCharacteristic
	tx     bt.DeviceCharacteristic
	framer FrameEncoder
}

// Send delivers a framed payload to the peripheral RX characteristic.
func (l *CentralLink) Send(data []byte) error {
	frame, err := EncodeFrame(data)
	if err != nil {
		return err
	}
	return WriteChunks(frame, 20, func(part []byte) error {
		_, err := l.rx.WriteWithoutResponse(part)
		return err
	})
}

// ConnectionManager tracks active central and inbound peripheral peers.
type ConnectionManager struct {
	adapter   *bt.Adapter
	onMessage OnMessageFunc

	mu           sync.RWMutex
	central      map[string]*CentralLink
	inboundKey   string
	inboundReady bool
}

// NewConnectionManager creates a connection registry.
func NewConnectionManager(adapter *bt.Adapter, onMessage OnMessageFunc) *ConnectionManager {
	return &ConnectionManager{
		adapter:   adapter,
		onMessage: onMessage,
		central:   make(map[string]*CentralLink),
	}
}

// Connect establishes a central connection to a discovered peer and sets up notifications.
func (m *ConnectionManager) Connect(peer DiscoveredPeer) (string, error) {
	addr, err := parseAddress(peer.Address)
	if err != nil {
		return "", err
	}

	device, err := m.adapter.Connect(addr, bt.ConnectionParams{})
	if err != nil {
		return "", fmt.Errorf("connect: %w", err)
	}

	services, err := device.DiscoverServices([]bt.UUID{ChatServiceUUID})
	if err != nil {
		_ = device.Disconnect()
		return "", fmt.Errorf("discover services: %w", err)
	}
	service := services[0]

	chars, err := service.DiscoverCharacteristics([]bt.UUID{ChatRXUUID, ChatTXUUID})
	if err != nil {
		_ = device.Disconnect()
		return "", fmt.Errorf("discover characteristics: %w", err)
	}

	key := peerKey(peer)
	link := &CentralLink{
		key:    key,
		device: device,
		rx:     chars[0],
		tx:     chars[1],
	}

	err = link.tx.EnableNotifications(func(buf []byte) {
		m.handleRX(key, &link.framer, buf)
	})
	if err != nil {
		_ = device.Disconnect()
		return "", fmt.Errorf("enable notifications: %w", err)
	}

	m.mu.Lock()
	m.central[key] = link
	m.mu.Unlock()
	return key, nil
}

// Disconnect closes a central connection by peer key.
func (m *ConnectionManager) Disconnect(peerKey string) error {
	m.mu.Lock()
	link, ok := m.central[peerKey]
	if ok {
		delete(m.central, peerKey)
	}
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("peer %q is not connected", peerKey)
	}
	return link.device.Disconnect()
}

// Connected returns peer keys for active central links and any inbound peripheral peer.
func (m *ConnectionManager) Connected() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.central)+1)
	for k := range m.central {
		out = append(out, k)
	}
	if m.inboundReady && m.inboundKey != "" {
		found := false
		for _, k := range out {
			if k == m.inboundKey {
				found = true
				break
			}
		}
		if !found {
			out = append(out, m.inboundKey)
		}
	}
	return out
}

// GetCentral returns a central link for outbound sends.
func (m *ConnectionManager) GetCentral(peerKey string) (*CentralLink, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	link, ok := m.central[peerKey]
	return link, ok
}

// RegisterInbound marks a peripheral-initiated connection for inbound messaging.
func (m *ConnectionManager) RegisterInbound(device bt.Device) string {
	key := device.Address.String()
	m.mu.Lock()
	m.inboundKey = key
	m.inboundReady = true
	m.mu.Unlock()
	return key
}

// UnregisterInbound clears the inbound peripheral peer.
func (m *ConnectionManager) UnregisterInbound(key string) {
	m.mu.Lock()
	if m.inboundKey == key {
		m.inboundKey = ""
		m.inboundReady = false
	}
	m.mu.Unlock()
}

func (m *ConnectionManager) handleRX(peerKey string, framer *FrameEncoder, buf []byte) {
	frames, err := framer.Append(buf)
	if err != nil {
		return
	}
	for _, frame := range frames {
		if m.onMessage != nil {
			m.onMessage(peerKey, frame)
		}
	}
}

// HandlePeripheralRX processes writes to the peripheral RX characteristic.
func (m *ConnectionManager) HandlePeripheralRX(peerKey string, framer *FrameEncoder, value []byte) {
	m.handleRX(peerKey, framer, value)
}

func peerKey(peer DiscoveredPeer) string {
	if peer.Name != "" && peer.Name != peer.Address {
		return peer.Name
	}
	return peer.Address
}

func parseAddress(s string) (bt.Address, error) {
	var addr bt.Address
	addr.Set(s)
	return addr, nil
}
