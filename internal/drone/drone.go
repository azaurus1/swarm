package drone

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/azaurus1/swarm/internal/routing"
	"github.com/azaurus1/swarm/internal/udp"
)

type Drone struct {
	Id                int
	X                 float64
	Y                 float64
	VX                float64
	VY                float64
	TransmissionRange float64
	RoutingTable      routing.RoutingTable
	Addr              string
	UDPConn           *net.UDPConn
	RadioConn         *net.UDPConn
}

func (d *Drone) Start(radioAddr string, wg *sync.WaitGroup) {
	defer wg.Done()

	udpAddr, err := net.ResolveUDPAddr("udp", d.Addr)
	if err != nil {
		log.Println("error resolving udp address: ", err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("error listening on udp address: ", err)
		return
	}

	d.UDPConn = conn
	defer d.UDPConn.Close()

	// set up radio connection
	rUdpStr, err := net.ResolveUDPAddr("udp", radioAddr)
	if err != nil {
		log.Println("error resolving udp address: ", err)
		return
	}
	r, err := net.DialUDP("udp", nil, rUdpStr)
	if err != nil {
		log.Println("error dialling udp address: ", err)
		return
	}

	d.RadioConn = r
	defer d.RadioConn.Close()

	log.Println("listening on", d.Addr)

	go udp.Client(context.Background(), d.RadioConn.RemoteAddr().String(), teeReader)

}

func (d *Drone) SendMessage() {
	d.RadioConn.Write([]byte(d.ToString()))
	var buf [512]byte
	_, _, err := d.RadioConn.ReadFromUDP(buf[0:])
	if err != nil {
		log.Println("error reading from UDP: ", err)
	}
	log.Printf("Drone %d> %s", d.Id, string(buf[0:]))
}

func (d *Drone) ToString() string {
	s := fmt.Sprintf("%d,%f,%f,%f", d.Id, d.X, d.Y, d.TransmissionRange)

	return s
}
