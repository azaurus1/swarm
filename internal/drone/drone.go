package drone

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/azaurus1/swarm/internal/control"
	"github.com/azaurus1/swarm/internal/messaging"
	"github.com/azaurus1/swarm/internal/routing"
	"github.com/azaurus1/swarm/internal/types"
)

type Drone struct {
	Id                string
	X                 float64
	Y                 float64
	VX                float64
	VY                float64
	TransmissionRange float64
	SequenceNumber    int
	DataChan          chan []byte
	PathDiscoveryTime time.Duration
	TransportLayer    *messaging.TransportLayer
	AODVListener      *routing.AODVListener
	ContolLayer       *control.ControlLayer
}

func (d *Drone) Start(wg *sync.WaitGroup, radioChan chan []byte) {
	defer wg.Done()

	//
	d.PathDiscoveryTime = 30 * time.Second

	d.AODVListener = routing.NewAODVListener()
	d.TransportLayer = messaging.NewTransportLayer()
	d.ContolLayer = control.NewControlLayer()
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
			var droneMsg types.DroneMessage

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
				d.AODVListener.HandleAODVMessage(d.Id, d.PathDiscoveryTime, aMsg, radioChan)
			case "DATA":
				d.TransportLayer.HandleDataMessage(d.Id, d.SequenceNumber, droneMsg, radioChan, d.AODVListener)
			case "CONTROL":
				d.ContolLayer.HandleCommand(d.Id, d.SequenceNumber, droneMsg, radioChan, d.AODVListener)
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
				helloMsg := types.AODVMessage{
					Source:                 d.Id,
					Type:                   2,
					HopCount:               1,
					DestinationId:          d.Id,
					DestinationSequenceNum: 1,
					OriginatorId:           d.Id,
					OriginatorSequenceNum:  1,
					TTL:                    1,
				}

				helloDMsg := types.DroneMessage{
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

		reqDMsg := types.DroneMessage{
			Source: "1",
			Type:   "AODV",
			AODVPayload: types.AODVMessage{
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

			reqDMsg := types.DroneMessage{
				Source: "1",
				Type:   "DATA",
				DataPayload: types.DataMessage{
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

	wg.Add(1)
	go func() {
		// handling expired neighbours
		defer wg.Done()
		// sending a DATA

		time.Sleep(3 * time.Second)

		if d.Id == "1" {

			reqDMsg := types.DroneMessage{
				Source: "1",
				Type:   "CONTROL",
				ControlPayload: types.ControlMessage{
					Checksum:    "1738",
					RecipientID: "5",
					SenderID:    "1",
					Command:     "move",
					Params: map[string]string{
						"x": "2",
						"y": "1",
					},
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
