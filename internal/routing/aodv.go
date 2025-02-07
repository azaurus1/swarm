package routing

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type AODVListener struct {
	RoutingTable  RoutingTable
	ReceivedRREQs map[string]time.Time
	ReceivedRREPs map[string]time.Time
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

type RoutingTable struct {
	Entries map[string]RoutingTableEntry
	Mutex   *sync.Mutex
}

type RoutingTableEntry struct {
	ID             string
	SequenceNumber int
	NextHop        string
	HopCount       int
	Expiration     time.Time
}

func (r RoutingTableEntry) ToString() string {
	return fmt.Sprintf(
		"RoutingTableEntry{\n  ID: %s,\n  SequenceNumber: %d,\n  NextHop: %s,\n  HopCount: %d,\n Expiration: %s\n}",
		r.ID, r.SequenceNumber, r.NextHop, r.HopCount, r.Expiration.Format(time.RFC3339),
	)
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

func (a *AODVListener) CheckForRoute(destination string) bool {
	a.RoutingTable.Mutex.Lock()
	defer a.RoutingTable.Mutex.Unlock()

	_, exists := a.RoutingTable.Entries[destination]
	log.Printf("Route exists: %v", exists)
	return exists
}
