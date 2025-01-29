package radio

import (
	"encoding/json"
	"log"
	"math"
	"sync"

	"github.com/azaurus1/swarm/internal/drone"
	"github.com/azaurus1/swarm/internal/routing"
)

// This is simulating the "air" for the drones

type Radio struct {
	Drones map[string]drone.Drone
}

func (r *Radio) Serve(wg *sync.WaitGroup, radioChan chan []byte) {
	defer wg.Done()

	// in - radioChan
	// out - drones[i].DataChan
	wg.Add(1)
	go func() {
		for msg := range radioChan {

			req := routing.AODVMessage{}

			// unmarshall
			json.Unmarshal(msg, &req)

			for _, d := range r.Drones {
				if req.Source == d.Id {
					// ignore same id, obviously they are within their own range
					continue
				}
				inRange := r.calculateTransmission(req.Source, d.Id)
				if inRange {
					// log.Printf("drone %s is within range of drone %s", d.Id, req.Source)

					q := r.calculateLinkQuality(req.Source, d.Id)

					req.LinkQuality = q

					// marshall to json
					calcMsg, err := json.Marshal(req)
					if err != nil {
						log.Println("error marshalling calculated message: ", err)
					}

					// forward the message
					d.DataChan <- calcMsg
				}

			}

		}
	}()

}

func (r *Radio) calculateTransmission(sourceDroneID string, targetDroneID string) bool {
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

func (r *Radio) calculateLinkQuality(sourceDroneID string, targetDroneID string) float64 {
	dx := r.Drones[targetDroneID].X - r.Drones[sourceDroneID].X
	dy := r.Drones[targetDroneID].Y - r.Drones[sourceDroneID].Y

	d := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))

	q := 1 - (d / r.Drones[sourceDroneID].TransmissionRange)

	if q < 0 {
		q = 0
	}

	return q

}
