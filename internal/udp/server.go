package udp

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

func Server(ctx context.Context, address string) error {
	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		log.Println("error listening for packets: ", err)
		return nil
	}

	defer pc.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, 1024)

	go func() {
		for {
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				log.Println(err)
				return
			}

			fmt.Printf("server > packet-received: bytes=%d from=%s\n", n, addr.String())

			deadline := time.Now().Add(5 * time.Second)
			err = pc.SetWriteDeadline(deadline)
			if err != nil {
				doneChan <- err
				log.Println(err)
				return
			}

			n, err = pc.WriteTo(buffer[:n], addr)
			if err != nil {
				doneChan <- err
				log.Println(err)
				return
			}
			fmt.Printf("server > packet-written: bytes=%d to=%s\n", n, addr.String())
		}
	}()
	select {
	case <-ctx.Done():
		fmt.Println("server cancelled")
		err = ctx.Err()
	case err = <-doneChan:
		log.Println(err)
	}

	return nil
}
