//go:build linux || windows

package bluetooth

import (
	"fmt"

	bt "tinygo.org/x/bluetooth"
)

// Advertiser publishes this node as an available chat peer.
type Advertiser struct {
	adapter *bt.Adapter
	adv     *bt.Advertisement
	name    string
}

// NewAdvertiser creates an advertiser for the given local name.
func NewAdvertiser(adapter *bt.Adapter, localName string) *Advertiser {
	return &Advertiser{
		adapter: adapter,
		name:    localName,
	}
}

// Start begins BLE advertising with the chat service UUID.
func (a *Advertiser) Start() error {
	adv := a.adapter.DefaultAdvertisement()
	if err := adv.Configure(bt.AdvertisementOptions{
		LocalName:    a.name,
		ServiceUUIDs: []bt.UUID{ChatServiceUUID},
	}); err != nil {
		return fmt.Errorf("configure advertisement: %w", err)
	}
	if err := adv.Start(); err != nil {
		return fmt.Errorf("start advertisement: %w", err)
	}
	a.adv = adv
	return nil
}

// Stop stops advertising if it was started.
func (a *Advertiser) Stop() error {
	if a.adv == nil {
		return nil
	}
	err := a.adv.Stop()
	a.adv = nil
	return err
}
