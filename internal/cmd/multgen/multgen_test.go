package multgen_test

import (
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/aaa2ppp/be"
	fastapi "github.com/aaa2ppp/multgen/internal/api/fast"
	api "github.com/aaa2ppp/multgen/internal/api/std"
	"github.com/aaa2ppp/multgen/internal/solver"
	"github.com/aaa2ppp/multgen/internal/testutils"
	"github.com/valyala/fasthttp"
)

func BenchmarkHTTPServer(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		for _, prefix := range []string{"without", "with"} {
			b.Run(prefix+" keep alive", func(b *testing.B) {
				url, close := startHTTPServer(b)
				defer close()

				client := newFastHTTPClient(prefix == "with")

				b.ResetTimer()
				b.ReportAllocs()

				testutils.AddRPSMetricToBenchmark(b, func() {
					for i := 0; i < b.N; i++ {
						doGetFast(b, client, url)
					}
				})
			})
		}
	})

	b.Run("parallel", func(b *testing.B) {
		for _, prefix := range []string{"without", "with"} {
			b.Run(prefix+" keep alive", func(b *testing.B) {
				url, close := startHTTPServer(b)
				defer close()

				b.ResetTimer()
				b.ReportAllocs()

				testutils.AddRPSMetricToBenchmark(b, func() {
					b.RunParallel(func(pb *testing.PB) {
						client := newFastHTTPClient(prefix == "with")
						for pb.Next() {
							doGetFast(b, client, url)
						}
					})
				})
			})
		}
	})
}

func startHTTPServer(b *testing.B) (url string, closeServer func()) {
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

	addr := listener.Addr().String()

	url = "http://" + addr + "/get"
	closeServer = func() { server.Close() }

	return url, closeServer
}

func newHTTPClient(keepAlive bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   !keepAlive,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 5 * time.Second,
	}
}

func doGet(b *testing.B, client *http.Client, url string) {
	resp, err := client.Get(url)
	_ = err
	// be.Err(b, err, nil)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func BenchmarkFastHTTPServer(b *testing.B) {

	// TODO: Investigate source of 1 alloc/op (25B) — probably fasthttp.Client internals.

	b.Run("single", func(b *testing.B) {
		for _, prefix := range []string{"without", "with"} {
			b.Run(prefix+" keep alive", func(b *testing.B) {
				url, close := startFastHTTPServer(b)
				defer close()

				client := newFastHTTPClient(prefix == "with")

				b.ResetTimer()
				b.ReportAllocs()

				testutils.AddRPSMetricToBenchmark(b, func() {
					for i := 0; i < b.N; i++ {
						doGetFast(b, client, url)
					}
				})
			})
		}
	})

	b.Run("parallel", func(b *testing.B) {
		for _, prefix := range []string{"without", "with"} {
			b.Run(prefix+" keep alive", func(b *testing.B) {
				url, close := startFastHTTPServer(b)
				defer close()

				b.ResetTimer()
				b.ReportAllocs()

				testutils.AddRPSMetricToBenchmark(b, func() {
					b.RunParallel(func(pb *testing.PB) {
						client := newFastHTTPClient(prefix == "with")
						for pb.Next() {
							doGetFast(b, client, url)
						}
					})
				})
			})
		}
	})
}

func startFastHTTPServer(b *testing.B) (url string, closeServer func()) {
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

	addr := listener.Addr().String()
	url = "http://" + addr + "/get"

	closeServer = func() {
		listener.Close()
		<-serverClosed
	}

	return url, closeServer
}

func newFastHTTPClient(keepAlive bool) *fasthttp.Client {
	return &fasthttp.Client{
		MaxConnsPerHost:               100,
		MaxIdleConnDuration:           90 * time.Second,
		DisableHeaderNamesNormalizing: true,
		ReadTimeout:                   5 * time.Second,
		WriteTimeout:                  5 * time.Second,
	}
}

func doGetFast(b *testing.B, client *fasthttp.Client, url string) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	_ = client.Do(req, resp)
	// тело уже в resp.Body()
}
