package dummy

import (
	"bytes"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/aaa2ppp/multgen/internal/solver"
)

type Server struct {
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	WorkerTimeout time.Duration
	Solver        *solver.Solver
	counts
}

func (s *Server) Serve(listener net.Listener) error {
	ch := make(chan net.Conn)
	defer close(ch)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		s.acceptCountAdd()

		select {
		case ch <- conn:
		default:
			go s.worker(conn, ch)
		}
	}
}

// worker allows to reuse goroutines and buffers between connections.
// Receives the first connection as parameter and processes it.
// After processing it waits during the timeout the next connection from the channel to process.
// Terminates if the next connection is not received.
func (s *Server) worker(conn net.Conn, ch chan net.Conn) {
	s.workerCountAdd()

	ibuf := make([]byte, 1024)

	const responsePrefix = "HTTP/1.1 200 OK\r\n" +
		"Cache-Control: no-cache, no-store, must-revalidate\r\n" +
		"Pragma: no-cache\r\n" +
		"Expires: 0\r\n" +
		"Content-Type: application/json\r\n" +
		"Content-Length: "

	obuf := append(make([]byte, 0, 256), responsePrefix...)
	body := append(make([]byte, 0, 64), `{"result":`...)

	if s.WorkerTimeout <= 0 {
		s.process(conn, ibuf, obuf, body)
		return
	}

	tm := time.NewTimer(0) // create stopped timer (v1.23+)
	for {
		s.process(conn, ibuf, obuf, body)

		ok := false
		tm.Reset(s.WorkerTimeout)

		select {
		case <-tm.C:
		case conn, ok = <-ch:
			tm.Stop()
		}

		if !ok {
			return
		}
	}
}

// process the response by returning a multiplier for *any* GET request.
// Otherwise, the answer is method not allowed. Keep alive connection until
// the client closes connrction.
func (s *Server) process(conn net.Conn, ibuf, obuf, body []byte) {
	s.processCountAdd()

	defer conn.Close()

	for {
		// read GET request (body omit)
		for p := 0; ; {
			n, err := conn.Read(ibuf[p:])
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Printf("read: %v", err)
				return
			}

			if bytes.Contains(ibuf[max(0, p-3):n], []byte("\r\n\r\n")) {
				break
			}

			p += n
			if p == len(ibuf) {
				log.Printf("read: request too long")
				conn.Write([]byte("HTTP/1.1 431 request too long\r\n\r\n"))
				return
			}
		}

		if !bytes.EqualFold(ibuf[:4], []byte("GET ")) {
			conn.Write([]byte("HTTP/1.1 405 method not allowed\r\n\r\n"))
			continue
		}

		s.requestCountAdd()

		// calculate multiplier
		multiplier := s.Solver.Solve()

		{
			// prepare body
			body := strconv.AppendFloat(body, multiplier, 'g', -1, 64)
			body = append(body, '}')

			// write response
			obuf := strconv.AppendInt(obuf, int64(len(body)), 10)
			obuf = append(obuf, "\r\n\r\n"...)
			obuf = append(obuf, body...)
			conn.Write(obuf)
		}
	}
}
