package server_test

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/aaa2ppp/be"
	"github.com/valyala/fasthttp"

	fastapi "github.com/aaa2ppp/multgen/internal/api/fast"
	stdapi "github.com/aaa2ppp/multgen/internal/api/std"
	"github.com/aaa2ppp/multgen/internal/server"
	"github.com/aaa2ppp/multgen/internal/server/dummy"
	"github.com/aaa2ppp/multgen/internal/solver"
	"github.com/aaa2ppp/multgen/internal/testutils"
)

type Server interface {
	Serve(net.Listener) error
}

func BenchmarkServer_single(b *testing.B) {
	solver, err := solver.New(solver.DefaultConfig())
	be.Err(b, err, nil)

	tests := []struct {
		name   string
		server Server
	}{
		{
			"http server",
			&http.Server{
				Handler:      stdapi.New(solver),
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
			},
		},
		{
			"fasthttp server",
			&fasthttp.Server{
				Handler:      fastapi.New(solver),
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
			},
		},
		{
			"dummy server",
			&dummy.Server{
				Solver:       solver,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			be.Err(b, err, nil)

			_, stopServer := server.Start(tt.server, listener)

			addr := listener.Addr().String()
			url := "http://" + addr + "/get"
			client := newClient()

			testutils.AddRPSMetricToBenchmark(b, func() {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := client.Get(url); err != nil {
						b.Fatal(err)
					}
				}
			})

			client.Close()
			be.Err(b, stopServer(), nil)
		})
	}
}

func BenchmarkServer_parallel(b *testing.B) {
	solver, err := solver.New(solver.DefaultConfig())
	be.Err(b, err, nil)

	const (
		readTimeout   = 5 * time.Second
		writeTimeout  = 5 * time.Second
		workerTimeout = 100 * time.Microsecond
	)

	tests := []struct {
		name   string
		server Server
	}{
		{
			"http server",
			&http.Server{
				Handler:      stdapi.New(solver),
				ReadTimeout:  readTimeout,
				WriteTimeout: writeTimeout,
			},
		},
		{
			"fasthttp server",
			&fasthttp.Server{
				Handler:      fastapi.New(solver),
				ReadTimeout:  readTimeout,
				WriteTimeout: writeTimeout,
			},
		},
		{
			"dummy server",
			&dummy.Server{
				Solver:        solver,
				ReadTimeout:   readTimeout,
				WriteTimeout:  writeTimeout,
				WorkerTimeout: workerTimeout,
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			be.Err(b, err, nil)

			_, stopServer := server.Start(tt.server, listener)

			addr := listener.Addr().String()
			url := "http://" + addr + "/get"
			client := newClient()

			testutils.AddRPSMetricToBenchmark(b, func() {
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						if err := client.Get(url); err != nil {
							b.Fatal(err)
						}
					}
				})
			})

			client.Close()
			be.Err(b, stopServer(), nil)
		})
	}
}

type client struct {
	client fasthttp.Client
}

func newClient() *client {
	return &client{
		client: fasthttp.Client{
			MaxConnsPerHost:               100,
			MaxIdleConnDuration:           90 * time.Second,
			DisableHeaderNamesNormalizing: true,
			ReadTimeout:                   5 * time.Second,
			WriteTimeout:                  5 * time.Second,
		},
	}
}

func (c *client) Get(url string) error {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	return c.client.Do(req, resp)
}

func (c *client) Close() {
	c.client.CloseIdleConnections()
}
