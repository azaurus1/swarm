package drone

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

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
}

func (d *Drone) Start(wg *sync.WaitGroup, radioChan chan []byte) {
	defer wg.Done()
	// map
	routingTableEntries := make(map[string]routing.RoutingTableEntry)

	d.AODVListener.RoutingTable.Entries = routingTableEntries

	// data in - dataChan (this is data from radio/air)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range d.DataChan {
			log.Printf("drone %s > message received: %s", d.Id, msg)

			// unmarshall
			var aMsg routing.AODVMessage
			json.Unmarshal(msg, &aMsg)

			if aMsg.Type == "HELLO" {
				// handle the hello
				d.AODVListener.HandleHello(aMsg)

				// send local
			}

		}
	}()

	// send a HELLO for neighbour discovery
	go func() {
		defer wg.Done()
		req := routing.AODVMessage{
			Type:   "HELLO",
			Source: d.Id,
		}

		data, err := json.Marshal(req)
		if err != nil {
			log.Println("couldn't marshall hello message")
		}

		radioChan <- data

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// data out - radioChan
		radioChan <- []byte(d.Id)
	}()

}

func (d *Drone) ToString() string {
	s := fmt.Sprintf("%s,%f,%f,%f", d.Id, d.X, d.Y, d.TransmissionRange)

	return s
}
