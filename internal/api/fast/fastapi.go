package fastapi

import (
	"bytes"
	"strconv"

	"github.com/valyala/fasthttp"

	"github.com/aaa2ppp/multgen/internal/api/buffer"
)

type Solver interface {
	Solve() float64
}

func New(s Solver) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		if !ctx.IsGet() {
			ctx.Error("Method Not Allowed", fasthttp.StatusMethodNotAllowed)
			return
		}

		// No-cache
		ctx.Response.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		ctx.Response.Header.Set("Pragma", "no-cache")
		ctx.Response.Header.Set("Expires", "0")

		path := ctx.Path()
		switch {
		case bytes.Equal(path, []byte("/get")):
			GetHandler(s)(ctx)
		case bytes.Equal(path, []byte("/ping")):
			PingHandler(ctx)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}

func GetHandler(s Solver) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		multiplier := s.Solve()

		// TODO: Can we avoid the buffer pool and write directly to fasthttp's response buffer?

		// Get buffer from pool
		buf := buffer.Get()

		buf = append(buf, `{"result":`...)
		buf = strconv.AppendFloat(buf, multiplier, 'g', -1, 64)
		buf = append(buf, '}')

		ctx.SetContentType("application/json")
		ctx.SetBody(buf) // fasthttp делает copy

		// Return buffer to pool
		buffer.Put(buf)
	}
}

func PingHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetBodyString("pong")
}
