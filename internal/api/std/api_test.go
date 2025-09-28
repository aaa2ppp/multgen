// == internal/api/std/api_test.go ==

package api_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/aaa2ppp/be"
	api "github.com/aaa2ppp/multgen/internal/api/std"
	"github.com/aaa2ppp/multgen/internal/solver"
	"github.com/aaa2ppp/multgen/internal/testutils"
)

func Test_GetHandler(t *testing.T) {
	// Солвер
	s, err := solver.New(solver.DefaultConfig())
	be.Err(t, err, nil)

	// Хендлер
	handler := api.New(s)

	// Запрос
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	w := httptest.NewRecorder()

	// Вызов
	handler.ServeHTTP(w, req)

	// Проверки
	be.Equal(be.Require(t), w.Code, http.StatusOK)
	body := w.Body.String()
	be.Equal(t, body, `{"result":1}`)
	be.Equal(t, w.Header().Get("content-length"), strconv.Itoa(len(body))) // len(`{"result":1}`)
	be.Equal(t, w.Header().Get("content-type"), "application/json")
}

func Benchmark_getHandler(b *testing.B) {
	s, err := solver.New(solver.DefaultConfig())
	be.Err(b, err, nil)

	handler := api.New(s)
	req := httptest.NewRequest(http.MethodGet, "/get", nil)

	b.ResetTimer()
	b.ReportAllocs()

	testutils.AddRPSMetricToBenchmark(b, func() {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			// be.Equal(be.Require(b), w.Code, http.StatusOK)
		}
	})
}

func Test_PingHandler(t *testing.T) {
	s, err := solver.New(solver.DefaultConfig())
	be.Err(t, err, nil)
	handler := api.New(s)

	// Проверяем только статус и заголовки. Тело ответа неважно,
	// может быть любым (например, "pong", поздравление с Новым годом и т.д.).

	for _, method := range []string{http.MethodHead, http.MethodGet} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/ping", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			be.Equal(t, w.Code, http.StatusOK)
			be.Equal(t, w.Header().Get("pragma"), "no-cache")
			be.Equal(t, w.Header().Get("expires"), "0")

			cacheControl := w.Header().Get("cache-control")
			be.True(t, strings.Contains(cacheControl, "no-cache"))
			be.True(t, strings.Contains(cacheControl, "no-store"))
			be.True(t, strings.Contains(cacheControl, "must-revalidate"))
		})
	}
}
