//go:build dev

package dummy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aaa2ppp/be"
	"github.com/aaa2ppp/multgen/internal/solver"
)

func TestServer_Serve_AcceptsConnections(t *testing.T) {
	solver, err := solver.New(solver.DefaultConfig())
	be.Err(t, err, nil)

	srv := &Server{
		Solver: solver,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	be.Err(t, err, nil)
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Serve(listener)
	}()

	// Establish one connection
	conn, err := net.Dial("tcp", listener.Addr().String())
	be.Err(t, err, nil)
	conn.Close()

	// Give server time to process
	time.Sleep(10 * time.Millisecond)

	be.Equal(t, srv.AcceptCount(), int64(1))
	be.Equal(t, srv.WorkerCount(), int64(1))
	be.Equal(t, srv.ProcessCount(), int64(1))
}

func TestServer_ProcessCount_IncrementsPerRequest(t *testing.T) {
	solver, err := solver.New(solver.DefaultConfig())
	be.Err(t, err, nil)

	srv := &Server{
		Solver: solver,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	be.Err(t, err, nil)
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Serve(listener)
	}()

	// Send 3 requests over one keep-alive connection
	conn, err := net.Dial("tcp", listener.Addr().String())
	be.Err(t, err, nil)
	defer conn.Close()

	for i := 0; i < 3; i++ {
		req := "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: keep-alive\r\n\r\n"
		_, err := conn.Write([]byte(req))
		be.Err(t, err, nil)

		// Read response
		resp, err := httpReadResponse(conn)
		be.Err(t, err, nil)
		be.True(t, strings.HasPrefix(resp.Status, "200 "))
	}

	// Give time for counters to update
	time.Sleep(10 * time.Millisecond)

	be.Equal(t, srv.AcceptCount(), int64(1))
	be.Equal(t, srv.WorkerCount(), int64(1))
	be.Equal(t, srv.ProcessCount(), int64(1))
	be.Equal(t, srv.RequestCount(), int64(3))
}

func TestServer_NonGET_Returns405(t *testing.T) {
	solver, err := solver.New(solver.DefaultConfig())
	be.Err(t, err, nil)

	srv := &Server{
		Solver: solver,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	be.Err(t, err, nil)
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Serve(listener)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	be.Err(t, err, nil)
	defer conn.Close()

	req := "POST / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	_, err = conn.Write([]byte(req))
	be.Err(t, err, nil)

	resp, err := httpReadResponse(conn)
	be.Err(t, err, nil)
	be.True(t, strings.HasPrefix(resp.Status, "405 "))
}

func TestServer_WorkerReuse_WithTimeout(t *testing.T) {
	solver, err := solver.New(solver.DefaultConfig())
	be.Err(t, err, nil)

	srv := &Server{
		Solver:        solver,
		WorkerTimeout: 100 * time.Millisecond,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	be.Err(t, err, nil)
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Serve(listener)
	}()

	// First connection
	conn1, err := net.Dial("tcp", listener.Addr().String())
	be.Err(t, err, nil)
	_, err = conn1.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	be.Err(t, err, nil)
	_, err = httpReadResponse(conn1)
	be.Err(t, err, nil)
	conn1.Close()

	// Yield to let worker exit process() and start waiting on ch
	time.Sleep(1 * time.Millisecond)

	// Second connection
	conn2, err := net.Dial("tcp", listener.Addr().String())
	be.Err(t, err, nil)
	_, err = conn2.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	be.Err(t, err, nil)
	_, err = httpReadResponse(conn2)
	be.Err(t, err, nil)
	conn2.Close()

	// Expect 2 accepts, 1 worker (reused), 2 processes
	be.Equal(t, srv.AcceptCount(), int64(2))
	be.Equal(t, srv.WorkerCount(), int64(1))
	be.Equal(t, srv.ProcessCount(), int64(2))
	be.Equal(t, srv.RequestCount(), int64(2))
}

func TestServer_WorkerCreatesNew_WhenTimeoutExceeded(t *testing.T) {
	solver, err := solver.New(solver.DefaultConfig())
	be.Err(t, err, nil)

	srv := &Server{
		Solver:        solver,
		WorkerTimeout: 50 * time.Millisecond,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	be.Err(t, err, nil)
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Serve(listener)
	}()

	// First connection
	conn1, err := net.Dial("tcp", listener.Addr().String())
	be.Err(t, err, nil)
	_, err = conn1.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\nConnection: keep-alive\r\n\r\n"))
	be.Err(t, err, nil)
	_, err = httpReadResponse(conn1)
	be.Err(t, err, nil)
	conn1.Close()

	// Wait longer than WorkerTimeout
	time.Sleep(100 * time.Millisecond)

	// Second connection — should spawn new worker
	conn2, err := net.Dial("tcp", listener.Addr().String())
	be.Err(t, err, nil)
	_, err = conn2.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\nConnection: keep-alive\r\n\r\n"))
	be.Err(t, err, nil)
	_, err = httpReadResponse(conn2)
	be.Err(t, err, nil)
	conn2.Close()

	// Allow final processing
	time.Sleep(20 * time.Millisecond)

	be.Equal(t, srv.AcceptCount(), int64(2))
	be.Equal(t, srv.WorkerCount(), int64(2)) // two separate workers
	be.Equal(t, srv.ProcessCount(), int64(2))
}

// httpReadResponse reads one HTTP response from conn using textproto
func httpReadResponse(conn net.Conn) (*http.Response, error) {
	bufConn := bufio.NewReader(conn)
	tp := textproto.NewReader(bufConn)

	// Read status line
	statusLine, err := tp.ReadLine()
	if err != nil {
		return nil, err
	}

	// Parse status line: "HTTP/1.1 200 OK"
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid status line: %s", statusLine)
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	// Read headers until blank line
	headers := make(http.Header)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			return nil, err
		}
		if line == "" {
			break
		}
		// Simple header parsing (no folding)
		if i := strings.IndexByte(line, ':'); i >= 0 {
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			headers.Add(key, val)
		}
	}

	// Read body if Content-Length present
	var body []byte
	if clStr := headers.Get("Content-Length"); clStr != "" {
		cl, err := strconv.ParseInt(clStr, 10, 64)
		if err != nil {
			return nil, err
		}
		body = make([]byte, cl)
		_, err = io.ReadFull(bufConn, body)
		if err != nil {
			return nil, err
		}
	}

	return &http.Response{
		Status:        parts[1] + " " + strings.Join(parts[2:], " "),
		StatusCode:    statusCode,
		Header:        headers,
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
	}, nil
}

func TestServer_RequestTooLong_Returns431(t *testing.T) {
	solver, err := solver.New(solver.DefaultConfig())
	be.Err(t, err, nil)

	srv := &Server{
		Solver: solver,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	be.Err(t, err, nil)
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Serve(listener)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	be.Err(t, err, nil)
	defer conn.Close()

	// Craft a request with headers longer than BufInputSize (1024)
	longPath := strings.Repeat("A", 1100) // >1024, и плюс "GET /... HTTP/1.1\r\nHost: ...\r\n\r\n"
	req := "GET /" + longPath + " HTTP/1.1\r\nHost: localhost\r\n\r\n"
	_, err = conn.Write([]byte(req))
	be.Err(t, err, nil)

	resp, err := httpReadResponse(conn)
	be.Err(t, err, nil)

	be.True(t, strings.HasPrefix(resp.Status, "431 "))
	be.Equal(t, srv.AcceptCount(), int64(1))
	be.Equal(t, srv.WorkerCount(), int64(1))
	be.Equal(t, srv.ProcessCount(), int64(1))
	be.Equal(t, srv.RequestCount(), int64(0))
}
