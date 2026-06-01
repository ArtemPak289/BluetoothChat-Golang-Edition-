package chat

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Message is the JSON wire format for chat payloads over BLE.
type Message struct {
	ID        string `json:"id"`
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
	Text      string `json:"text"`
}

// NewMessage builds an outbound chat message with a fresh id and RFC3339 timestamp.
func NewMessage(sender, text string) Message {
	return Message{
		ID:        uuid.NewString(),
		Sender:    sender,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Text:      text,
	}
}

// Encode serializes the message to JSON bytes.
func (m Message) Encode() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("encode message: %w", err)
	}
	return data, nil
}

// DecodeMessage parses JSON bytes into a Message.
func DecodeMessage(data []byte) (Message, error) {
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return Message{}, fmt.Errorf("decode message: %w", err)
	}
	if m.ID == "" || m.Sender == "" || m.Text == "" {
		return Message{}, fmt.Errorf("decode message: missing required fields")
	}
	return m, nil
}
