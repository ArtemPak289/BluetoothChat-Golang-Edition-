package bluetooth

import bt "tinygo.org/x/bluetooth"

// Custom GATT service and characteristics for BLE chat (NUS-style split).
var (
	// ChatServiceUUID identifies the chat GATT service in advertisements and discovery.
	ChatServiceUUID = bt.New16BitUUID(0xFFE0)
	// ChatRXUUID is written by the central (incoming messages to the peripheral).
	ChatRXUUID = bt.New16BitUUID(0xFFE1)
	// ChatTXUUID is used for notifications from the peripheral to the central.
	ChatTXUUID = bt.New16BitUUID(0xFFE2)
)
