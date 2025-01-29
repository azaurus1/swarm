package routing

import "errors"

type AODVListener struct {
	RoutingTable RoutingTable
}

type AODVMessage struct {
	Type           string  `json:"type"`
	Source         string  `json:"source"`
	SequenceNumber int     `json:"sequence_number"`
	LinkQuality    float64 `json:"link_quality"`
}

type RoutingTable struct {
	Entries map[string]RoutingTableEntry
}

type RoutingTableEntry struct {
	ID             string
	SequenceNumber int
	TTL            int
}

func (a *AODVListener) HandleHello(msg AODVMessage) error {
	// We have received a HELLO, add this Source to our neighbours and reset the link timer
	n := RoutingTableEntry{
		ID:             msg.Source,
		SequenceNumber: msg.SequenceNumber,
		TTL:            30,
	}

	if n.SequenceNumber < a.RoutingTable.Entries[n.ID].SequenceNumber {
		// we cant update
		return errors.New("sequence number is lower than current")
	}

	a.RoutingTable.Entries[n.ID] = n
	return nil
}
