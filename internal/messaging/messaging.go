package messaging

type TransportLayer struct{}

type Message struct {
	RecpientID string `json:"recipient_id"`
	SenderID   string `json:"sender_id"`
	Data       []byte `json:"data"`
}
