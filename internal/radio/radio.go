package radio

import (
	"log"
	"strconv"
	"sync"

	"github.com/azaurus1/swarm/internal/drone"
)

// This is simulating the "air" for the drones

type Radio struct {
	Addr  string
	Nodes map[int]Node
}

type Node struct {
	Id                int
	CurrentPosX       float64
	CurrentPosY       float64
	TransmissionRange float64
}

func (r *Radio) Serve(drones []drone.Drone, wg *sync.WaitGroup, radioChan chan string) {
	defer wg.Done()

	// in - radioChan
	// out - drones[i].DataChan
	go func() {
		for msg := range radioChan {
			log.Printf("radio > message received: %s", msg)

			droneID, err := strconv.Atoi(msg)
			if err != nil || droneID <= 0 || droneID > len(drones) {
				log.Printf("radio > invalid drone ID: %s", msg)
				continue
			}

			drones[droneID-1].DataChan <- "ACK"
		}
	}()

}
