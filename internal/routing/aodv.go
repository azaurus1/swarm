package routing

import "net"

type AODVListener struct {
	Listener net.UDPConn
}
type RouteRequest struct{}
type RouteReply struct{}
type RouteError struct{}
