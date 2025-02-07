package routing

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/azaurus1/swarm/internal/types"
)

type AODVListener struct {
	RoutingTable  RoutingTable
	ReceivedRREQs map[string]time.Time
	ReceivedRREPs map[string]time.Time
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

func NewAODVListener() *AODVListener {
	return &AODVListener{
		RoutingTable: RoutingTable{
			Entries: make(map[string]RoutingTableEntry),
			Mutex:   &sync.Mutex{},
		},
		ReceivedRREQs: make(map[string]time.Time),
		ReceivedRREPs: make(map[string]time.Time),
	}
}

func (a *AODVListener) HandleAODVMessage(droneId string, pathDiscoveryTime time.Duration, aMsg types.AODVMessage, radioChan chan []byte) {
	if aMsg.Type == 1 {
		log.Printf("Processing RREQ from %s", aMsg.OriginatorId)
		rreqKey := fmt.Sprintf("%s-%s", aMsg.OriginatorId, aMsg.RREQID)

		hopCount := aMsg.HopCount + 1

		if droneId == aMsg.OriginatorId {
			log.Println("Ignoring because I am the originator")
			return
		}

		if entry, exists := a.RoutingTable.Entries[aMsg.OriginatorId]; exists {
			if entry.SequenceNumber <= aMsg.OriginatorSequenceNum && aMsg.HopCount < a.RoutingTable.Entries[aMsg.OriginatorId].HopCount {
				// valid, update
				log.Printf("%s: Valid, Updating", droneId)
				a.RoutingTable.Entries[aMsg.OriginatorId] = RoutingTableEntry{
					ID:             aMsg.OriginatorId,
					SequenceNumber: aMsg.OriginatorSequenceNum,
					NextHop:        aMsg.Source,
					HopCount:       hopCount,
					Expiration:     time.Now().Add(30 * time.Second),
				}
			}
		} else {
			// doesnt exist, create
			log.Printf("%s: Doesnt exist, creating", droneId)
			a.RoutingTable.Entries[aMsg.OriginatorId] = RoutingTableEntry{
				ID:             aMsg.OriginatorId,
				SequenceNumber: aMsg.OriginatorSequenceNum,
				NextHop:        aMsg.Source,
				HopCount:       hopCount,
				Expiration:     time.Now().Add(30 * time.Second),
			}

		}

		if timestamp, exists := a.ReceivedRREQs[rreqKey]; exists {
			if time.Since(timestamp) < pathDiscoveryTime {
				// log.Println("Silently discarding this RREQ")
				return
			}
		}

		a.ReceivedRREQs[rreqKey] = time.Now()

		// Generate an RREP (RFC3561 6.6)
		if droneId == aMsg.DestinationId {
			// sending RREP
			log.Println("I am the destination for this message")

			repMsg := types.AODVMessage{
				Source:                 droneId,
				Type:                   2,
				HopCount:               1,
				DestinationId:          aMsg.DestinationId,
				DestinationSequenceNum: aMsg.OriginatorSequenceNum + 1,
				OriginatorId:           aMsg.OriginatorId,
				OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
			}

			repDMsg := types.DroneMessage{
				Source:      repMsg.Source,
				Type:        "AODV",
				AODVPayload: repMsg,
			}

			data, _ := json.Marshal(repDMsg)

			radioChan <- data

		} else if _, exists := a.RoutingTable.Entries[aMsg.DestinationId]; exists {
			log.Println("Route exists in the routing table")
			// we have a route to the destination, we can send the RREP
			repMsg := types.AODVMessage{
				Source:                 droneId,
				Type:                   2,
				HopCount:               hopCount,
				DestinationId:          aMsg.DestinationId,
				DestinationSequenceNum: aMsg.DestinationSequenceNum,
				OriginatorId:           aMsg.OriginatorId,
				OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
			}

			repDMsg := types.DroneMessage{
				Source:      repMsg.Source,
				Type:        "AODV",
				AODVPayload: repMsg,
			}

			data, _ := json.Marshal(repDMsg)

			radioChan <- data
		} else {
			log.Println("Repeating RREQ")
			reqMsg := types.AODVMessage{
				Source:                 droneId,
				Type:                   1,
				RREQID:                 aMsg.RREQID,
				OriginatorId:           aMsg.OriginatorId,
				OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
				DestinationId:          aMsg.DestinationId,
				DestinationSequenceNum: aMsg.DestinationSequenceNum,
				HopCount:               hopCount,
			}

			reqDMsg := types.DroneMessage{
				Source:      reqMsg.Source,
				Type:        "AODV",
				AODVPayload: reqMsg,
			}

			data, _ := json.Marshal(reqDMsg)

			radioChan <- data
		}

	} else if aMsg.Type == 2 {
		log.Printf("Processing RREP from %s", aMsg.OriginatorId)
		rrepKey := fmt.Sprintf("%s-%s", aMsg.OriginatorId, aMsg.RREQID)

		if droneId == aMsg.DestinationId {
			log.Println("Discarding RREP")
			return
		}

		// Instead of looking up aMsg.Source, look up the route for the destination:
		if entry, exists := a.RoutingTable.Entries[aMsg.DestinationId]; exists {
			if entry.SequenceNumber <= aMsg.DestinationSequenceNum && aMsg.HopCount < a.RoutingTable.Entries[aMsg.DestinationId].HopCount {
				// Valid update: update the route for the destination
				log.Println("Valid, Updating")
				a.RoutingTable.Entries[aMsg.DestinationId] = RoutingTableEntry{
					ID:             aMsg.DestinationId,
					SequenceNumber: aMsg.DestinationSequenceNum,
					NextHop:        aMsg.Source, // the neighbor from which we received the RREP
					HopCount:       aMsg.HopCount,
					Expiration:     time.Now().Add(30 * time.Second),
				}
			}
		} else {
			// Doesn't exist, so create a route entry for the destination
			log.Println("Doesnt exist, creating")
			log.Println(aMsg)
			a.RoutingTable.Entries[aMsg.DestinationId] = RoutingTableEntry{
				ID:             aMsg.DestinationId,
				SequenceNumber: aMsg.DestinationSequenceNum,
				NextHop:        aMsg.Source,
				HopCount:       aMsg.HopCount,
				Expiration:     time.Now().Add(30 * time.Second),
			}
		}

		if timestamp, exists := a.ReceivedRREPs[rrepKey]; exists {
			if time.Since(timestamp) < pathDiscoveryTime {
				// log.Println("Silently discarding this RREP")
				return
			}
		}

		a.ReceivedRREPs[rrepKey] = time.Now()

		// Increment hop count for forwarding purposes
		hopCount := aMsg.HopCount + 1

		if droneId == aMsg.OriginatorId {
			log.Println("I am the originator of this RREP")
			// For the originator, install/update the route for the destination.
			a.RoutingTable.Entries[aMsg.DestinationId] = RoutingTableEntry{
				ID:             aMsg.DestinationId,
				SequenceNumber: aMsg.DestinationSequenceNum,
				NextHop:        aMsg.Source,
				HopCount:       aMsg.HopCount,
				Expiration:     time.Now().Add(30 * time.Second),
			}

			return
		} else {
			// Repeat: forward the RREP with an incremented hop count
			log.Printf("Drone %s repeating rrep", droneId)
			repMsg := types.AODVMessage{
				Source:                 droneId,
				Type:                   2,
				HopCount:               hopCount,
				DestinationId:          aMsg.DestinationId,
				DestinationSequenceNum: aMsg.DestinationSequenceNum,
				OriginatorId:           aMsg.OriginatorId,
				OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
			}

			repDMsg := types.DroneMessage{
				Source:      repMsg.Source,
				Type:        "AODV",
				AODVPayload: repMsg,
			}

			data, _ := json.Marshal(repDMsg)
			radioChan <- data
		}
	}
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
