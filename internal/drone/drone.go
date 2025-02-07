package drone

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/azaurus1/swarm/internal/messaging"
	"github.com/azaurus1/swarm/internal/routing"
)

type Drone struct {
	Id                string
	X                 float64
	Y                 float64
	VX                float64
	VY                float64
	TransmissionRange float64
	AODVListener      *routing.AODVListener
	DataChan          chan []byte
	PathDiscoveryTime time.Duration
	TransportLayer    *messaging.TransportLayer
	SequenceNumber    int
}

type DroneMessage struct {
	Source      string                `json:"source"`
	Type        string                `json:"type"`
	AODVPayload routing.AODVMessage   `json:"aodv_payload"`
	DataPayload messaging.DataMessage `json:"data_payload"`
}

func (d *Drone) Start(wg *sync.WaitGroup, radioChan chan []byte) {
	defer wg.Done()

	//
	d.PathDiscoveryTime = 30 * time.Second

	d.AODVListener = &routing.AODVListener{
		RoutingTable: routing.RoutingTable{
			Entries: make(map[string]routing.RoutingTableEntry),
			Mutex:   &sync.Mutex{},
		},
		ReceivedRREQs: make(map[string]time.Time),
		ReceivedRREPs: make(map[string]time.Time),
	}

	d.TransportLayer = &messaging.TransportLayer{
		ReceivedMessages: make(map[string]time.Time),
	}

	// hello ticker
	helloTicker := time.NewTicker(1000 * time.Millisecond)
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
			var droneMsg DroneMessage

			err := json.Unmarshal(msg, &droneMsg)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			switch droneMsg.Type {
			case "AODV":
				aMsg := droneMsg.AODVPayload

				if aMsg.TTL < aMsg.HopCount+1 {
					// log.Println("TTL expired, discarding message")
				}

				// log.Printf("Drone %s routing table: ", d.Id)

				// for _, e := range d.AODVListener.RoutingTable.Entries {
				// 	log.Println(e.ToString())
				// }

				switch aMsg.Type {
				case 1:
					// RREQ
					// First create or update a route to the PREVIOUS hop without a valid seq num

					rreqKey := fmt.Sprintf("%s-%s", aMsg.OriginatorId, aMsg.RREQID)

					hopCount := aMsg.HopCount + 1

					if d.Id == aMsg.OriginatorId {
						log.Println("Ignoring because I am the originator")
						continue
					}

					if entry, exists := d.AODVListener.RoutingTable.Entries[aMsg.OriginatorId]; exists {
						if entry.SequenceNumber <= aMsg.OriginatorSequenceNum && aMsg.HopCount < d.AODVListener.RoutingTable.Entries[aMsg.OriginatorId].HopCount {
							// valid, update
							log.Printf("%s: Valid, Updating", d.Id)
							d.AODVListener.RoutingTable.Entries[aMsg.OriginatorId] = routing.RoutingTableEntry{
								ID:             aMsg.OriginatorId,
								SequenceNumber: aMsg.OriginatorSequenceNum,
								NextHop:        aMsg.Source,
								HopCount:       hopCount,
								Expiration:     time.Now().Add(30 * time.Second),
							}
						}
					} else {
						// doesnt exist, create
						log.Printf("%s: Doesnt exist, creating", d.Id)
						d.AODVListener.RoutingTable.Entries[aMsg.OriginatorId] = routing.RoutingTableEntry{
							ID:             aMsg.OriginatorId,
							SequenceNumber: aMsg.OriginatorSequenceNum,
							NextHop:        aMsg.Source,
							HopCount:       hopCount,
							Expiration:     time.Now().Add(30 * time.Second),
						}

					}

					if timestamp, exists := d.AODVListener.ReceivedRREQs[rreqKey]; exists {
						if time.Since(timestamp) < d.PathDiscoveryTime {
							// log.Println("Silently discarding this RREQ")
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
							HopCount:               1,
							DestinationId:          aMsg.DestinationId,
							DestinationSequenceNum: aMsg.OriginatorSequenceNum + 1,
							OriginatorId:           aMsg.OriginatorId,
							OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
						}

						repDMsg := DroneMessage{
							Source:      repMsg.Source,
							Type:        "AODV",
							AODVPayload: repMsg,
						}

						data, _ := json.Marshal(repDMsg)

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

						repDMsg := DroneMessage{
							Source:      repMsg.Source,
							Type:        "AODV",
							AODVPayload: repMsg,
						}

						data, _ := json.Marshal(repDMsg)

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
							HopCount:               hopCount,
						}

						reqDMsg := DroneMessage{
							Source:      reqMsg.Source,
							Type:        "AODV",
							AODVPayload: reqMsg,
						}

						data, _ := json.Marshal(reqDMsg)

						radioChan <- data
					}

					// for _, e := range d.AODVListener.RoutingTable.Entries {
					// 	log.Println(e.ToString())
					// }

				case 2:
					// Handling RREP
					// First, search for the previous hop
					rrepKey := fmt.Sprintf("%s-%s", aMsg.OriginatorId, aMsg.RREQID)

					if d.Id == aMsg.DestinationId {
						log.Println("Discarding RREP")
						continue
					}

					// Instead of looking up aMsg.Source, look up the route for the destination:
					if entry, exists := d.AODVListener.RoutingTable.Entries[aMsg.DestinationId]; exists {
						if entry.SequenceNumber <= aMsg.DestinationSequenceNum && aMsg.HopCount < d.AODVListener.RoutingTable.Entries[aMsg.DestinationId].HopCount {
							// Valid update: update the route for the destination
							log.Println("Valid, Updating")
							d.AODVListener.RoutingTable.Entries[aMsg.DestinationId] = routing.RoutingTableEntry{
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
						d.AODVListener.RoutingTable.Entries[aMsg.DestinationId] = routing.RoutingTableEntry{
							ID:             aMsg.DestinationId,
							SequenceNumber: aMsg.DestinationSequenceNum,
							NextHop:        aMsg.Source,
							HopCount:       aMsg.HopCount,
							Expiration:     time.Now().Add(30 * time.Second),
						}
					}

					if timestamp, exists := d.AODVListener.ReceivedRREPs[rrepKey]; exists {
						if time.Since(timestamp) < d.PathDiscoveryTime {
							// log.Println("Silently discarding this RREP")
							continue
						}
					}

					d.AODVListener.ReceivedRREPs[rrepKey] = time.Now()

					// Increment hop count for forwarding purposes
					hopCount := aMsg.HopCount + 1

					if d.Id == aMsg.OriginatorId {
						log.Println("I am the originator of this RREP")
						// For the originator, install/update the route for the destination.
						d.AODVListener.RoutingTable.Entries[aMsg.DestinationId] = routing.RoutingTableEntry{
							ID:             aMsg.DestinationId,
							SequenceNumber: aMsg.DestinationSequenceNum,
							NextHop:        aMsg.Source,
							HopCount:       aMsg.HopCount,
							Expiration:     time.Now().Add(30 * time.Second),
						}

						continue
					} else {
						// Repeat: forward the RREP with an incremented hop count
						log.Printf("Drone %s repeating rrep", d.Id)
						repMsg := routing.AODVMessage{
							Source:                 d.Id,
							Type:                   2,
							HopCount:               hopCount,
							DestinationId:          aMsg.DestinationId,
							DestinationSequenceNum: aMsg.DestinationSequenceNum,
							OriginatorId:           aMsg.OriginatorId,
							OriginatorSequenceNum:  aMsg.OriginatorSequenceNum,
						}

						repDMsg := DroneMessage{
							Source:      repMsg.Source,
							Type:        "AODV",
							AODVPayload: repMsg,
						}

						data, _ := json.Marshal(repDMsg)
						radioChan <- data
					}

				}
			case "DATA":

				// log.Println("data message received")

				dMsg := droneMsg.DataPayload

				if d.Id != dMsg.RecipientID {
					if _, exists := d.TransportLayer.ReceivedMessages[dMsg.Checksum]; !exists {
						// add to received messages map
						d.TransportLayer.ReceivedMessages[dMsg.Checksum] = time.Now()
					} else {
						// log.Printf("%s - I have received this message, ignoring..", d.Id)
						continue
					}

					routeExists := d.AODVListener.CheckForRoute(dMsg.RecipientID)

					if routeExists {
						// get next hop, then send message
						// propagate message
						droneMsg.Source = d.Id

						dData, err := json.Marshal(droneMsg)
						if err != nil {
							log.Println("error marshalling data message for rebroadcast")
						}

						radioChan <- dData
					} else {
						// send RREQ
						rreq := DroneMessage{
							Source: "1",
							Type:   "AODV",
							AODVPayload: routing.AODVMessage{
								Source:                d.Id,
								Type:                  1,
								RREQID:                "1738",
								DestinationId:         dMsg.RecipientID,
								OriginatorId:          dMsg.SenderID,
								OriginatorSequenceNum: d.SequenceNumber,
								UnknownSequenceNum:    true,
							},
						}

						data, _ := json.Marshal(rreq)

						radioChan <- data
					}

				} else {
					log.Printf("%s - I have received a data message", d.Id)
				}

			}

		}
	}()

	// send a HELLO for neighbour discovery
	// hellos should actually be RREP with TTL of 1
	wg.Add(1)
	go func() {
		// handling expired neighbours
		defer wg.Done()
		defer helloTicker.Stop()

		for {
			select {
			case <-done:
				return
			case <-helloTicker.C:
				helloMsg := routing.AODVMessage{
					Source:                 d.Id,
					Type:                   2,
					HopCount:               1,
					DestinationId:          d.Id,
					DestinationSequenceNum: 1,
					OriginatorId:           d.Id,
					OriginatorSequenceNum:  1,
					TTL:                    1,
				}

				helloDMsg := DroneMessage{
					Source:      helloMsg.Source,
					Type:        "AODV",
					AODVPayload: helloMsg,
				}

				data, _ := json.Marshal(helloDMsg)

				radioChan <- data
			}

		}

	}()

	// Expiration ticker
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

	// sending a RREQ
	if d.Id == "1" {

		reqDMsg := DroneMessage{
			Source: "1",
			Type:   "AODV",
			AODVPayload: routing.AODVMessage{
				Source:                "1",
				Type:                  1,
				RREQID:                "1738",
				DestinationId:         "5",
				OriginatorId:          "1",
				OriginatorSequenceNum: d.SequenceNumber,
				UnknownSequenceNum:    true,
			},
		}

		data, err := json.Marshal(reqDMsg)
		if err != nil {
			log.Println("error marshalling drone message: ", err)
		}

		radioChan <- data
		d.SequenceNumber++

	}

	wg.Add(1)
	go func() {
		// handling expired neighbours
		defer wg.Done()
		// sending a DATA

		time.Sleep(3 * time.Second)

		if d.Id == "1" {

			reqDMsg := DroneMessage{
				Source: "1",
				Type:   "DATA",
				DataPayload: messaging.DataMessage{
					Checksum:    "1738",
					RecipientID: "5",
					SenderID:    "1",
					Data:        []byte("Hello"),
				},
			}

			data, err := json.Marshal(reqDMsg)
			if err != nil {
				log.Println("error marshalling drone message: ", err)
			}

			radioChan <- data
			d.SequenceNumber++

		}
	}()
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
