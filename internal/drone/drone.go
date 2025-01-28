package drone

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/azaurus1/swarm/internal/routing"
)

type Drone struct {
	Id                int
	X                 float64
	Y                 float64
	VX                float64
	VY                float64
	TransmissionRange float64
	RoutingTable      routing.RoutingTable
	DataChan          chan string
}

func (d *Drone) Start(radioAddr string, wg *sync.WaitGroup, radioChan chan string) {
	defer wg.Done()
	// channels

	// data in - dataChan (this is data from radio/air)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range d.DataChan {
			log.Printf("drone %d > message received: %s", d.Id, msg)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// data out - radioChan
		id := strconv.Itoa(d.Id)
		radioChan <- id
	}()

}

func (d *Drone) ToString() string {
	s := fmt.Sprintf("%d,%f,%f,%f", d.Id, d.X, d.Y, d.TransmissionRange)

	return s
}
