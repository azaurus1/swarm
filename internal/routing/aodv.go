package routing

import (
	"errors"
	"log"
	"sync"
	"time"
)

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
	Mutex   *sync.Mutex
}

type RoutingTableEntry struct {
	ID             string
	SequenceNumber int
	TTL            int
	Expiration     time.Time
}

// check for routing table entries that are past expiration, delete them if they are
func (a *AODVListener) CheckExpiredNeighbours() error {
	for _, entry := range a.RoutingTable.Entries {
		if entry.Expiration.Before(time.Now()) {
			a.RoutingTable.Mutex.Lock()
			log.Println("expired entry found, deleting...")
			delete(a.RoutingTable.Entries, entry.ID)
			a.RoutingTable.Mutex.Unlock()
		}
	}

	return nil
}

// handle the HELLO messages
func (a *AODVListener) HandleHello(msg AODVMessage) error {
	// We have received a HELLO, add this Source to our neighbours and reset the link timer
	n := RoutingTableEntry{
		ID:             msg.Source,
		SequenceNumber: msg.SequenceNumber,
		TTL:            2,
		Expiration:     time.Now().Add(2 * time.Second),
	}

	a.RoutingTable.Mutex.Lock()
	if n.SequenceNumber < a.RoutingTable.Entries[n.ID].SequenceNumber {
		// we cant update
		return errors.New("sequence number is lower than current")
	}

	// lock the table
	a.RoutingTable.Entries[n.ID] = n
	a.RoutingTable.Mutex.Unlock()

	return nil
}

// handle the RREQ
func (a *AODVListener) HandleRouteRequest(msg AODVMessage) error {
	return nil
}

// handle the RREP
func (a *AODVListener) HandleRouteReply(msg AODVMessage) error {
	return nil
}
