package udp

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

func Client(ctx context.Context, address string, reader io.Reader) error {
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	doneChan := make(chan error, 1)

	go func() {
		n, err := io.Copy(conn, reader)
		if err != nil {
			doneChan <- err
			return
		}

		fmt.Printf("packet-written: bytes=%d\n", n)

		buffer := make([]byte, 1000)

		deadline := time.Now().Add(5 * time.Second)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			doneChan <- err
			return
		}

		nRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			doneChan <- err
			return
		}

		fmt.Printf("packet-received: bytes=%d from=%s\n", nRead, addr.String())

		doneChan <- nil

	}()

	select {
	case <-ctx.Done():
		fmt.Println("cancelled")
		err = ctx.Err()
	case err = <-doneChan:
	}

	return nil
}
