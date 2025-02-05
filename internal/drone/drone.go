package drone

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/azaurus1/swarm/internal/routing"
)

type Drone struct {
	Id                string
	X                 float64
	Y                 float64
	VX                float64
	VY                float64
	TransmissionRange float64
	AODVListener      routing.AODVListener
	DataChan          chan []byte
	PathDiscoveryTime time.Duration
}

func (d *Drone) Start(wg *sync.WaitGroup, radioChan chan []byte) {
	defer wg.Done()

	//
	d.PathDiscoveryTime = 30 * time.Second

	// map routing table
	routingTableEntries := make(map[string]routing.RoutingTableEntry)
	d.AODVListener.RoutingTable.Entries = routingTableEntries

	// create map for rreqs
	recRREQ := make(map[string]time.Time)
	d.AODVListener.ReceivedRREQs = recRREQ

	// create map for rreps
	recRREP := make(map[string]time.Time)
	d.AODVListener.ReceivedRREPs = recRREP

	// make mutex for table
	mu := sync.Mutex{}
	d.AODVListener.RoutingTable.Mutex = &mu

	// hello ticker
	// helloTicker := time.NewTicker(1000 * time.Millisecond)
	// expiry ticke
	expirationTicker := time.NewTicker(1000 * time.Millisecond)

	done := make(chan bool)

	if d.Id == "" {
		log.Println("Drone ID is empty at start!")
		return
	}

	// data in - dataChan (this is data from radio/air)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range d.DataChan {
			log.Printf("drone %s > message received: %s", d.Id, msg)

			// unmarshall
			var aMsg routing.AODVMessage
			json.Unmarshal(msg, &aMsg)

			switch aMsg.Type {
			case 1:
				// RREQ
				// First create or update a route to the PREVIOUS hop without a valid seq num

				rreqKey := fmt.Sprintf("%s-%s", aMsg.OriginatorId, aMsg.RREQID)

				hopCount := aMsg.HopCount + 1

				if entry, exists := d.AODVListener.RoutingTable.Entries[rreqKey]; exists {
					if entry.SequenceNumber <= aMsg.OriginatorSequenceNum {
						// valid, update
						log.Println("Valid, Updating")
						d.AODVListener.RoutingTable.Entries[aMsg.Source] = routing.RoutingTableEntry{
							ID:             aMsg.Source,
							SequenceNumber: aMsg.OriginatorSequenceNum,
							NextHop:        aMsg.OriginatorId,
							HopCount:       hopCount,
							Expiration:     time.Now().Add(30 * time.Second),
						}
					}
				} else {
					// doesnt exist, create
					log.Println("Doesnt exist, creating")
					d.AODVListener.RoutingTable.Entries[aMsg.Source] = routing.RoutingTableEntry{
						ID:             aMsg.Source,
						SequenceNumber: aMsg.OriginatorSequenceNum,
						NextHop:        aMsg.Source,
						HopCount:       hopCount,
						Expiration:     time.Now().Add(30 * time.Second),
					}
				}

				if timestamp, exists := d.AODVListener.ReceivedRREQs[rreqKey]; exists {
					if time.Since(timestamp) < d.PathDiscoveryTime {
						log.Println("Silently discarding this RREQ")
						continue
					}
				}

				d.AODVListener.ReceivedRREQs[rreqKey] = time.Now()

				// Generate an RREP (RFC3561 6.6)
				if d.Id == aMsg.DestinationId {
					// sending RREP
					log.Println("I am the destination for this message")

					repMsg := routing.AODVMessage{
						Source:                 d.Id,
						Type:                   2,
						HopCount:               0,
						DestinationId:          aMsg.DestinationId,
						DestinationSequenceNum: aMsg.DestinationSequenceNum,
						OriginatorId:           aMsg.OriginatorId,
						OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
					}

					data, _ := json.Marshal(repMsg)

					radioChan <- data

				} else if _, exists := d.AODVListener.RoutingTable.Entries[aMsg.DestinationId]; exists {
					log.Println("Route exists in the routing table")
					// we have a route to the destination, we can send the RREP
					repMsg := routing.AODVMessage{
						Source:                 d.Id,
						Type:                   2,
						HopCount:               hopCount,
						DestinationId:          aMsg.DestinationId,
						DestinationSequenceNum: aMsg.DestinationSequenceNum,
						OriginatorId:           aMsg.OriginatorId,
						OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
					}

					data, _ := json.Marshal(repMsg)

					radioChan <- data
				} else {
					log.Println("Repeating RREQ")
					reqMsg := routing.AODVMessage{
						Source:                 d.Id,
						Type:                   1,
						RREQID:                 aMsg.RREQID,
						OriginatorId:           aMsg.OriginatorId,
						OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
						DestinationId:          aMsg.DestinationId,
						DestinationSequenceNum: aMsg.DestinationSequenceNum,
					}

					data, _ := json.Marshal(reqMsg)

					radioChan <- data
				}

				// for _, e := range d.AODVListener.RoutingTable.Entries {
				// 	log.Println(e.ToString())
				// }

			case 2:
				// Handling RREP
				// First, search for the previous hop
				rrepKey := fmt.Sprintf("%s-%s", aMsg.OriginatorId, aMsg.RREQID)

				// if entry, exists := d.AODVListener.RoutingTable.Entries[rrepKey]; exists {
				// 	if entry.SequenceNumber <= aMsg.OriginatorSequenceNum {
				// 		// valid, update
				// 		log.Println("Valid, Updating")
				// 		d.AODVListener.RoutingTable.Entries[aMsg.Source] = routing.RoutingTableEntry{
				// 			ID:             aMsg.Source,
				// 			SequenceNumber: aMsg.OriginatorSequenceNum,
				// 			NextHop:        aMsg.OriginatorId,
				// 			HopCount:       hopCount,
				// 			Expiration:     time.Now().Add(30 * time.Second),
				// 		}
				// 	}
				// } else {
				// 	// doesnt exist, create
				// 	log.Println("Doesnt exist, creating: ", aMsg.Source, aMsg.HopCount)
				// 	d.AODVListener.RoutingTable.Entries[aMsg.Source] = routing.RoutingTableEntry{
				// 		ID:             aMsg.Source,
				// 		SequenceNumber: aMsg.OriginatorSequenceNum,
				// 		NextHop:        aMsg.Source,
				// 		HopCount:       hopCount,
				// 		Expiration:     time.Now().Add(30 * time.Second),
				// 	}

				// }

				if timestamp, exists := d.AODVListener.ReceivedRREPs[rrepKey]; exists {
					if time.Since(timestamp) < d.PathDiscoveryTime {
						log.Println("Silently discarding this RREP")
						continue
					}
				}

				d.AODVListener.ReceivedRREPs[rrepKey] = time.Now()

				hopCount := aMsg.HopCount + 1

				if d.Id == aMsg.OriginatorId {
					log.Println("I am the originator of this RREP")

					d.AODVListener.RoutingTable.Entries[aMsg.DestinationId] = routing.RoutingTableEntry{
						ID:             aMsg.DestinationId,
						SequenceNumber: aMsg.OriginatorSequenceNum,
						NextHop:        aMsg.Source,
						HopCount:       hopCount,
						Expiration:     time.Now().Add(30 * time.Second),
					}

					for _, e := range d.AODVListener.RoutingTable.Entries {
						log.Println(e.ToString())
					}

					continue

				} else {
					// Repeat
					repMsg := routing.AODVMessage{
						Source:                 d.Id,
						Type:                   2,
						HopCount:               hopCount,
						DestinationId:          aMsg.DestinationId,
						DestinationSequenceNum: aMsg.DestinationSequenceNum,
						OriginatorId:           aMsg.OriginatorId,
						OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
					}

					data, _ := json.Marshal(repMsg)

					radioChan <- data
				}

				for _, e := range d.AODVListener.RoutingTable.Entries {
					log.Println(e.ToString())
				}

			}

		}
	}()

	// send a HELLO for neighbour discovery
	// hellos should actually be RREP with TTL of 1
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	defer helloTicker.Stop()
	// 	req := routing.AODVMessage{
	// 		Type: 1,
	// 	}

	// 	data, err := json.Marshal(req)
	// 	if err != nil {
	// 		log.Println("couldn't marshall hello message")
	// 	}

	// 	for {
	// 		select {
	// 		case <-done:
	// 			return
	// 		case <-helloTicker.C:
	// 			radioChan <- data
	// 		}
	// 	}

	// }()

	wg.Add(1)
	go func() {
		// handling expired neighbours
		defer wg.Done()
		defer expirationTicker.Stop()

		for {
			select {
			case <-done:
				return
			case <-expirationTicker.C:
				d.AODVListener.CheckExpiredNeighbours()
			}

		}

	}()

	if d.Id == "1" {
		RREQ := routing.AODVMessage{
			Source:                "1",
			Type:                  1,
			RREQID:                "1738",
			DestinationId:         "3",
			OriginatorId:          "1",
			OriginatorSequenceNum: 1,
			UnknownSequenceNum:    true,
		}

		data, _ := json.Marshal(RREQ)

		radioChan <- data

	}
}

func (d *Drone) ToString() string {
	s := fmt.Sprintf("%s,%f,%f,%f", d.Id, d.X, d.Y, d.TransmissionRange)

	return s
}

func (d *Drone) UpdateLocation(delta time.Duration, lBound, rBound, tBound, bBound float64) {
	d.X += delta.Seconds() * d.VX
	d.Y += delta.Seconds() * d.VY

	if d.X <= lBound || d.X >= rBound {
		d.VX *= -1
	}

	if d.Y <= bBound || d.Y >= tBound {
		d.VY *= -1
	}
	// log.Printf("New position for %s: (%f,%f)", d.Id, d.X, d.Y)
}
