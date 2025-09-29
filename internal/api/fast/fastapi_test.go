package fastapi_test

import (
	"net/http"
	"testing"

	"github.com/aaa2ppp/be"
	"github.com/valyala/fasthttp"

	fastapi "github.com/aaa2ppp/multgen/internal/api/fast"
	"github.com/aaa2ppp/multgen/internal/solver"
	"github.com/aaa2ppp/multgen/internal/testutils"
)

func TestGetHandler(t *testing.T) {
	s, err := solver.New(solver.DefaultConfig())
	be.Err(t, err, nil)

	handler := fastapi.GetHandler(s)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/get")

	handler(ctx)

	be.Equal(be.Require(t), ctx.Response.StatusCode(), http.StatusOK)
	be.Equal(t, string(ctx.Response.Body()), `{"result":1}`)
	be.Equal(t, string(ctx.Response.Header.ContentType()), "application/json")
}

func Benchmark_getHandler(b *testing.B) {
	s, err := solver.New(solver.DefaultConfig())
	be.Err(b, err, nil)

	handler := fastapi.GetHandler(s)
	ctx := &fasthttp.RequestCtx{}

	b.ResetTimer()
	b.ReportAllocs()

	testutils.AddRPSMetricToBenchmark(b, func() {
		for i := 0; i < b.N; i++ {
			ctx.Request.SetRequestURI("/get")
			handler(ctx)
			// be.Equal(be.Require(b), ctx.Response.StatusCode(), http.StatusOK)
		}
	})
}
