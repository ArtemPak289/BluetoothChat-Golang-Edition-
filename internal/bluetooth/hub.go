package bluetooth

import (
	"fmt"

	bt "tinygo.org/x/bluetooth"
)

// Hub coordinates scanning, advertising, the GATT chat service, and peer connections.
type Hub struct {
	adapter *bt.Adapter
	name    string

	scanner    *Scanner
	advertiser *Advertiser
	service    *ChatService
	conns      *ConnectionManager
}

// PeerEventFunc is called when an inbound peripheral connection goes up or down.
type PeerEventFunc func(peerKey string, connected bool)

// NewHub wires BLE subsystems for a node with the given display name.
func NewHub(localName string, onPayload func(peerKey string, payload []byte), onPeer PeerEventFunc) (*Hub, error) {
	adapter := bt.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("enable adapter: %w", err)
	}

	conns := NewConnectionManager(adapter, onPayload)
	h := &Hub{
		adapter:    adapter,
		name:       localName,
		scanner:    NewScanner(adapter),
		advertiser: NewAdvertiser(adapter, localName),
		service:    NewChatService(adapter, onPayload),
		conns:      conns,
	}

	adapter.SetConnectHandler(func(device bt.Device, connected bool) {
		key := device.Address.String()
		if connected {
			h.service.SetPeerKey(key)
			h.conns.RegisterInbound(device)
			if onPeer != nil {
				onPeer(key, true)
			}
		} else {
			h.service.ClearPeerKey(key)
			h.conns.UnregisterInbound(key)
			if onPeer != nil {
				onPeer(key, false)
			}
		}
	})

	return h, nil
}

// StartAdvertising registers the GATT service and begins BLE advertising.
func (h *Hub) StartAdvertising() error {
	if err := h.service.Register(); err != nil {
		return err
	}
	return h.advertiser.Start()
}

// StartScan begins discovering nearby chat peers.
func (h *Hub) StartScan() error {
	return h.scanner.Start()
}

// StopScan stops discovery.
func (h *Hub) StopScan() error {
	return h.scanner.Stop()
}

// DiscoveredPeers returns peers found by the last scan.
func (h *Hub) DiscoveredPeers() []DiscoveredPeer {
	return h.scanner.List()
}

// Connect dials a discovered peer as a GATT central.
func (h *Hub) Connect(peer string) (string, error) {
	_ = h.scanner.Stop()
	found, ok := h.scanner.Find(peer)
	if !ok {
		return "", fmt.Errorf("peer %q not found; run scan first", peer)
	}
	return h.conns.Connect(found)
}

// Disconnect closes a connection by peer key.
func (h *Hub) Disconnect(peer string) error {
	if err := h.conns.Disconnect(peer); err == nil {
		return nil
	}
	h.conns.UnregisterInbound(peer)
	h.service.ClearPeerKey(peer)
	return nil
}

// ConnectedPeers lists active peer keys.
func (h *Hub) ConnectedPeers() []string {
	return h.conns.Connected()
}

// SendTo delivers a framed payload to a connected peer (central or inbound peripheral).
func (h *Hub) SendTo(peerKey string, data []byte) error {
	if link, ok := h.conns.GetCentral(peerKey); ok {
		return link.Send(data)
	}
	for _, p := range h.conns.Connected() {
		if p == peerKey {
			return h.service.Send(data)
		}
	}
	return fmt.Errorf("peer %q is not connected", peerKey)
}
