package multgen_test

import (
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/aaa2ppp/be"
	"github.com/aaa2ppp/multgen/internal/api"
	"github.com/aaa2ppp/multgen/internal/solver"
	"github.com/aaa2ppp/multgen/internal/testutils"
)

func BenchmarkHTTPServer(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		for _, prefix := range []string{"without", "with"} {
			b.Run(prefix+" keep alive", func(b *testing.B) {
				url, close := startHTTPServer(b)
				defer close()

				client := newHTTPClient(prefix == "with")

				b.ResetTimer()
				b.ReportAllocs()

				testutils.AddRPSMetricToBenchmark(b, func() {
					for i := 0; i < b.N; i++ {
						doGet(b, client, url)
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
						client := newHTTPClient(prefix == "with")
						for pb.Next() {
							doGet(b, client, url)
						}
					})
				})
			})
		}
	})
}

func startHTTPServer(b *testing.B) (url string, close func()) {
	var stubSolverCfg = solver.Config{
		RTP:           0.5,
		Algorithm:     "stub",
		MinMultiplier: solver.MinMultiplier,
		MaxMultiplier: solver.MaxMultiplier,
		K:             1,
	}

	s, err := solver.New(stubSolverCfg)
	be.Err(b, err, nil)

	// Запускаем реальный HTTP-сервер
	server := &http.Server{
		Addr:    "127.0.0.1:0", // случайный порт
		Handler: api.New(s),
	}

	// Запуск в фоне
	listener, err := net.Listen("tcp", server.Addr)
	be.Err(b, err, nil)

	go func() {
		_ = server.Serve(listener)
	}()

	// Ждём, пока сервер запустится
	addr := listener.Addr().String()
	url = "http://" + addr + "/get"

	return url, func() { server.Close() }
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
	be.Err(b, err, nil)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
