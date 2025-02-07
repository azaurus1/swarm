package messaging

import "time"

type TransportLayer struct {
	ReceivedMessages map[string]time.Time
}

type DataMessage struct {
	Checksum    string `json:"checksum"`
	RecipientID string `json:"recipient_id"`
	SenderID    string `json:"sender_id"`
	Data        []byte `json:"data"`
}
