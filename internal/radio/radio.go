package radio

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/azaurus1/swarm/internal/drone"
	"github.com/azaurus1/swarm/internal/udp"
)

// This is simulating the "air" for the drones

type Radio struct {
	Addr    string
	UDPConn *net.UDPConn
	Nodes   map[int]Node
}

type Node struct {
	Id                int
	CurrentPosX       float64
	CurrentPosY       float64
	TransmissionRange float64
	Addr              string
	Conn              *net.UDPConn
}

func (r *Radio) Serve(drones []drone.Drone, wg *sync.WaitGroup) {
	defer wg.Done()

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// this is where the UDP 'air' sim is
	// we can modify transmission based on factors we set up like transmission range etc.
	// udpAddr, err := net.ResolveUDPAddr("udp", r.Addr)
	// if err != nil {
	// 	log.Println("error resolving udp address: ", err)
	// 	return
	// }

	// conn, err := net.ListenUDP("udp", udpAddr)
	// if err != nil {
	// 	log.Println("error listening on udp address: ", err)
	// 	return
	// }

	// r.UDPConn = conn
	// defer r.UDPConn.Close()

	// set up drones as nodes
	nodes := make(map[int]Node)

	for _, drone := range drones {
		n := Node{
			Id:                drone.Id,
			CurrentPosX:       drone.X,
			CurrentPosY:       drone.Y,
			TransmissionRange: drone.TransmissionRange, // radius of transmission
			Addr:              drone.Addr,              // address string
		}

		// build connections
		udpAddr, err := net.ResolveUDPAddr("udp", n.Addr)
		if err != nil {
			log.Println("error resolving udp address: ", err)
		}
		conn, err := net.DialUDP("udp", nil, udpAddr)

		n.Conn = conn
		nodes[n.Id] = n
	}
	defer func() {
		for _, node := range nodes {
			node.Conn.Close()
		}
	}()

	// log.Println("nodes: ", nodes)

	log.Println("radio listening on", r.Addr)

	go udp.Server(context.Background(), r.Addr)
}
