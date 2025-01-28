package radio

import (
	"log"
	"strconv"
	"sync"

	"github.com/azaurus1/swarm/internal/drone"
)

// This is simulating the "air" for the drones

type Radio struct {
	Drones map[int]drone.Drone
}

func (r *Radio) Serve(wg *sync.WaitGroup, radioChan chan []byte) {
	defer wg.Done()

	// in - radioChan
	// out - drones[i].DataChan
	go func() {
		for msg := range radioChan {
			log.Printf("radio > message received: %s", msg)

			droneID, err := strconv.Atoi(string(msg))
			if err != nil || droneID <= 0 || droneID > len(r.Drones) {
				log.Printf("radio > invalid drone ID: %s", msg)
				continue
			}

			for _, d := range r.Drones {
				if droneID == d.Id {
					// ignore same id, obviously they are within their own range
					continue
				}
				inRange := r.calculateTransmission(droneID, d.Id)
				if inRange {
					log.Printf("drone %d is within range of drone %d", d.Id, droneID)
				}

			}

		}
	}()

}

func (r *Radio) calculateTransmission(sourceDroneID int, targetDroneID int) bool {
	// (cX - x)^2 + (cY - y)^2 = transmissionRange^2

	// point is in range if
	// (cX - x)^2 + (cY - y)^2 <= transmissionRange^2
	centerX := r.Drones[sourceDroneID].X
	centerY := r.Drones[sourceDroneID].Y

	dX := centerX - r.Drones[targetDroneID].X
	dY := centerY - r.Drones[targetDroneID].Y
	sqrR := r.Drones[sourceDroneID].TransmissionRange * r.Drones[sourceDroneID].TransmissionRange

	distanceSquared := dX*dX + dY*dY

	// log.Printf("%f + %f <= %f", dX*dX, dY*dY, sqrR)

	if distanceSquared > sqrR {
		return false
	}

	return true

}
