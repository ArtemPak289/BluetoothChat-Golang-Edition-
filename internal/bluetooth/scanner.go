package bluetooth

import (
	"strings"
	"sync"

	bt "tinygo.org/x/bluetooth"
)

// DiscoveredPeer is a BLE device seen during scanning.
type DiscoveredPeer struct {
	Address string
	Name    string
	RSSI    int16
}

// Scanner discovers nearby devices advertising the chat service.
type Scanner struct {
	adapter *bt.Adapter

	mu    sync.RWMutex
	peers map[string]DiscoveredPeer

	stop chan struct{}
	done chan struct{}
}

// NewScanner creates a scanner bound to the default adapter.
func NewScanner(adapter *bt.Adapter) *Scanner {
	return &Scanner{
		adapter: adapter,
		peers:   make(map[string]DiscoveredPeer),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// Start runs scanning in a background goroutine until Stop is called.
func (s *Scanner) Start() error {
	select {
	case <-s.done:
		s.stop = make(chan struct{})
		s.done = make(chan struct{})
	default:
	}

	s.mu.Lock()
	s.peers = make(map[string]DiscoveredPeer)
	s.mu.Unlock()

	go s.run()
	return nil
}

func (s *Scanner) run() {
	defer close(s.done)

	err := s.adapter.Scan(func(_ *bt.Adapter, result bt.ScanResult) {
		if !result.HasServiceUUID(ChatServiceUUID) {
			return
		}
		addr := result.Address.String()
		name := strings.TrimSpace(result.LocalName())
		if name == "" {
			name = addr
		}
		peer := DiscoveredPeer{
			Address: addr,
			Name:    name,
			RSSI:    result.RSSI,
		}
		s.mu.Lock()
		s.peers[addr] = peer
		s.mu.Unlock()
	})
	if err != nil {
		return
	}

	<-s.stop
	_ = s.adapter.StopScan()
}

// Stop ends an active scan. It is safe to call when scanning is not active.
func (s *Scanner) Stop() error {
	select {
	case <-s.done:
		return nil
	default:
	}
	close(s.stop)
	<-s.done
	return nil
}

// List returns a snapshot of discovered peers.
func (s *Scanner) List() []DiscoveredPeer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]DiscoveredPeer, 0, len(s.peers))
	for _, p := range s.peers {
		out = append(out, p)
	}
	return out
}

// Find resolves a peer by address or case-insensitive name.
func (s *Scanner) Find(peer string) (DiscoveredPeer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.peers[peer]; ok {
		return p, true
	}
	lower := strings.ToLower(peer)
	for _, p := range s.peers {
		if strings.EqualFold(p.Address, peer) || strings.ToLower(p.Name) == lower {
			return p, true
		}
	}
	return DiscoveredPeer{}, false
}
