package server

import (
	"errors"
	"net"
)

type Server interface {
	Serve(net.Listener) error
}

func Start(server Server, listener net.Listener) (done chan<- error, stopServer func() error) {
	ch := make(chan error)
	go func() {
		defer close(done)
		done <- server.Serve(listener)
	}()

	stopServer = func() error {
		listener.Close()
		if err := <-ch; err != nil && !errors.Is(err, net.ErrClosed) {
			return err
		}
		return nil
	}

	return ch, stopServer
}
