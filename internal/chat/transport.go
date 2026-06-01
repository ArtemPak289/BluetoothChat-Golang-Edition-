package chat

// Transport delivers encoded message bytes to a single peer.
type Transport interface {
	Send(data []byte) error
}

// HubTransport adapts a send callback to the Transport interface.
type HubTransport struct {
	SendFn func(data []byte) error
}

// Send delivers encoded bytes through the configured callback.
func (t HubTransport) Send(data []byte) error {
	return t.SendFn(data)
}
