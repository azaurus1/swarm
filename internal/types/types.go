package types

import (
	"time"
)

type DroneMessage struct {
	Source         string         `json:"source"`
	Type           string         `json:"type"`
	AODVPayload    AODVMessage    `json:"aodv_payload"`
	DataPayload    DataMessage    `json:"data_payload"`
	ControlPayload ControlMessage `json:"control_payload"`
}

type DataMessage struct {
	Checksum    string `json:"checksum"`
	RecipientID string `json:"recipient_id"`
	SenderID    string `json:"sender_id"`
	Data        []byte `json:"data"`
}

type AODVMessage struct {
	Source                 string        `json:"source"`
	Type                   int           `json:"aodv_type"`
	HopCount               int           `json:"hop_count"`
	RREQID                 string        `json:"rreq_id"`
	DestinationId          string        `json:"destination_id"`
	DestinationSequenceNum int           `json:"destination_sequence_num"`
	OriginatorId           string        `json:"originator_id"`
	OriginatorSequenceNum  int           `json:"originator_sequence_num"`
	LifeTime               time.Duration `json:"lifetime"`
	UnknownSequenceNum     bool          `json:"unknown_sequence_num"`
	TTL                    int           `json:"ttl"`
}

type ControlMessage struct {
	Checksum    string            `json:"checksum"`
	RecipientID string            `json:"recipient_id"`
	SenderID    string            `json:"sender_id"`
	Command     string            `json:"command"`
	Params      map[string]string `json:"params"`
}
