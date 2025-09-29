package multgen_test

import (
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/aaa2ppp/be"
	"github.com/valyala/fasthttp"

	fastapi "github.com/aaa2ppp/multgen/internal/api/fast"
	api "github.com/aaa2ppp/multgen/internal/api/std"
	"github.com/aaa2ppp/multgen/internal/solver"
	"github.com/aaa2ppp/multgen/internal/testutils"
)

const (
	withKeepAlive    = "with keep alive"
	withoutKeepAlive = "without keep alive"
)

func BenchmarkHTTPServer(b *testing.B) {
	benchs := []struct {
		name        string
		startServer func(*testing.B) (addr string, stopServer func())
	}{
		{
			"std",
			startHTTPServer,
		},
		{
			// TODO: Investigate source of 1 alloc/op (25B) — probably fasthttp.Client or Solver internals.
			"fast",
			startFastHTTPServer,
		},
	}

	for _, tb := range benchs {
		b.Run(tb.name, func(b *testing.B) {

			b.Run("single", func(b *testing.B) {
				for _, name := range []string{withoutKeepAlive, withKeepAlive} {
					b.Run(name, func(b *testing.B) {
						addr, stopServer := tb.startServer(b)
						defer stopServer()

						url := "http://" + addr + "/get"

						// always use a fast client to reduce the pressure on the benchmark
						client := newFastHTTPClient(b, name == withKeepAlive)

						b.ResetTimer()
						b.ReportAllocs()

						testutils.AddRPSMetricToBenchmark(b, func() {
							for i := 0; i < b.N; i++ {
								client.Get(url)
							}
						})
					})
				}
			})

			b.Run("parallel", func(b *testing.B) {
				for _, name := range []string{withoutKeepAlive, withKeepAlive} {
					b.Run(name, func(b *testing.B) {
						addr, stopServer := tb.startServer(b)
						defer stopServer()

						url := "http://" + addr + "/get"

						b.ResetTimer()
						b.ReportAllocs()

						testutils.AddRPSMetricToBenchmark(b, func() {
							b.RunParallel(func(pb *testing.PB) {

								// always use a fast client to reduce the pressure on the benchmark
								client := newFastHTTPClient(b, name == withKeepAlive)

								for pb.Next() {
									client.Get(url)
								}
							})
						})
					})
				}
			})
		})
	}
}

func startHTTPServer(b *testing.B) (addr string, stopServer func()) {
	s, err := solver.New(solver.DefaultConfig())
	be.Err(b, err, nil)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	be.Err(b, err, nil)

	server := &http.Server{
		Handler: api.New(s),
	}

	go func() {
		_ = server.Serve(listener)
	}()

	addr = listener.Addr().String()

	return addr, func() { server.Close() }
}

type httpClient struct {
	b      *testing.B
	client http.Client
}

func newHTTPClient(b *testing.B, keepAlive bool) *httpClient {
	return &httpClient{
		b: b,
		client: http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   !keepAlive,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
			Timeout: 5 * time.Second,
		},
	}
}
func (c *httpClient) Get(url string) {
	resp, _ := c.client.Get(url)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func startFastHTTPServer(b *testing.B) (addr string, closeServer func()) {
	s, err := solver.New(solver.DefaultConfig())
	be.Err(b, err, nil)
	handler := fastapi.New(s)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	be.Err(b, err, nil)

	serverClosed := make(chan struct{})
	go func() {
		defer close(serverClosed)
		_ = fasthttp.Serve(listener, handler)
	}()

	addr = listener.Addr().String()

	return addr, func() {
		listener.Close()
		<-serverClosed
	}
}

type fastHTTPClient struct {
	b         *testing.B
	client    fasthttp.Client
	keepAlive bool
}

func newFastHTTPClient(b *testing.B, keepAlive bool) *fastHTTPClient {
	return &fastHTTPClient{
		b: b,
		client: fasthttp.Client{
			MaxConnsPerHost:               100,
			MaxIdleConnDuration:           90 * time.Second,
			DisableHeaderNamesNormalizing: true,
			ReadTimeout:                   5 * time.Second,
			WriteTimeout:                  5 * time.Second,
		},
		keepAlive: keepAlive,
	}
}

func (c *fastHTTPClient) Get(url string) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	if !c.keepAlive {
		req.Header.Set("Connection", "close")
	}

	_ = c.client.Do(req, resp)
	// тело уже в resp.Body()
}
