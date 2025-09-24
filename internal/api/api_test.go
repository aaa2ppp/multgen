// == internal/api/api_test.go ==

package api_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/aaa2ppp/be"
	"github.com/aaa2ppp/multgen/internal/api"
	"github.com/aaa2ppp/multgen/internal/solver"
	"github.com/aaa2ppp/multgen/internal/testutils"
)

var stubSolverCfg = solver.Config{
	RTP:           0.5,
	Algorithm:     "stub",
	MinMultiplier: solver.MinMultiplier,
	MaxMultiplier: solver.MaxMultiplier,
	K:             1,
}

func Test_GetHandler(t *testing.T) {
	// Солвер
	s, err := solver.New(stubSolverCfg)
	be.Err(t, err, nil)

	// Хендлер
	handler := api.New(s)

	// Запрос
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	w := httptest.NewRecorder()

	// Вызов
	handler.ServeHTTP(w, req)

	// Проверки
	be.Equal(t, w.Code, http.StatusOK)
	body := w.Body.String()
	be.Equal(t, body, `{"result":1}`)
	be.Equal(t, w.Header().Get("content-length"), strconv.Itoa(len(body))) // len(`{"result":1}`)
	be.Equal(t, w.Header().Get("content-type"), "application/json")
}

func Benchmark_getHandler(b *testing.B) {
	s, err := solver.New(stubSolverCfg)
	be.Err(b, err, nil)

	handler := api.New(s)
	req := httptest.NewRequest(http.MethodGet, "/get", nil)

	b.ResetTimer()
	b.ReportAllocs()

	testutils.AddRPSMetricToBenchmark(b, func() {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			be.Equal(b, http.StatusOK, w.Code)
		}
	})
}
