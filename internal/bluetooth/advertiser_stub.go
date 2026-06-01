//go:build darwin

package bluetooth

import bt "tinygo.org/x/bluetooth"

// Advertiser is a stub on platforms without BLE advertising support.
type Advertiser struct {
	name string
}

// NewAdvertiser creates a placeholder advertiser.
func NewAdvertiser(_ *bt.Adapter, localName string) *Advertiser {
	return &Advertiser{name: localName}
}

// Start reports that advertising is unavailable on this platform.
func (a *Advertiser) Start() error {
	return ErrPeripheralUnsupported
}

// Stop is a no-op on unsupported platforms.
func (a *Advertiser) Stop() error {
	return nil
}
