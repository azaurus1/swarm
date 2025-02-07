package control

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/azaurus1/swarm/internal/routing"
	"github.com/azaurus1/swarm/internal/types"
)

type ControlLayer struct {
	ReceivedCommands map[string]time.Time
	Mutex            *sync.Mutex
}

func NewControlLayer() *ControlLayer {
	return &ControlLayer{
		ReceivedCommands: make(map[string]time.Time),
		Mutex:            &sync.Mutex{},
	}
}

func (c *ControlLayer) HandleCommand(droneId string, droneSeqNum int, droneMsg types.DroneMessage, radioChan chan []byte, aodv *routing.AODVListener) {
	cMsg := droneMsg.ControlPayload

	if droneId != cMsg.RecipientID {
		if _, exists := c.ReceivedCommands[cMsg.Checksum]; !exists {
			c.ReceivedCommands[cMsg.Checksum] = time.Now()
		} else {
			return
		}

		routeExists := aodv.CheckForRoute(cMsg.RecipientID)

		if routeExists {
			droneMsg.Source = droneId

			dData, err := json.Marshal(droneMsg)
			if err != nil {
				log.Println("error marshalling control message for rebroadcast ", err)
			}

			radioChan <- dData
		} else {
			// send RREQ
			rreq := types.DroneMessage{
				Source: "1",
				Type:   "AODV",
				AODVPayload: types.AODVMessage{
					Source:                droneId,
					Type:                  1,
					RREQID:                "1738",
					DestinationId:         cMsg.RecipientID,
					OriginatorId:          cMsg.SenderID,
					OriginatorSequenceNum: droneSeqNum,
					UnknownSequenceNum:    true,
				},
			}

			data, _ := json.Marshal(rreq)

			radioChan <- data
		}
	} else {
		log.Println("I have received a command for me")
	}

}
