package messaging

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/azaurus1/swarm/internal/routing"
	"github.com/azaurus1/swarm/internal/types"
)

type TransportLayer struct {
	ReceivedMessages map[string]time.Time
	Mutex            *sync.Mutex
}

func NewTransportLayer() *TransportLayer {
	return &TransportLayer{
		ReceivedMessages: make(map[string]time.Time),
		Mutex:            &sync.Mutex{},
	}
}

func (t *TransportLayer) HandleDataMessage(droneId string, droneSeqNum int, droneMsg types.DroneMessage, radioChan chan []byte, aodv *routing.AODVListener) {
	dMsg := droneMsg.DataPayload

	if droneId != dMsg.RecipientID {
		if _, exists := t.ReceivedMessages[dMsg.Checksum]; !exists {
			// add to received messages map
			t.ReceivedMessages[dMsg.Checksum] = time.Now()
		} else {
			return
		}

		routeExists := aodv.CheckForRoute(dMsg.RecipientID)

		if routeExists {
			// get next hop, then send message
			// propagate message
			droneMsg.Source = droneId

			dData, err := json.Marshal(droneMsg)
			if err != nil {
				log.Println("error marshalling data message for rebroadcast")
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
					DestinationId:         dMsg.RecipientID,
					OriginatorId:          dMsg.SenderID,
					OriginatorSequenceNum: droneSeqNum,
					UnknownSequenceNum:    true,
				},
			}

			data, _ := json.Marshal(rreq)

			radioChan <- data
		}
	} else {
		log.Printf("%s - I have received a data message", droneId)

	}
}
