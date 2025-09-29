package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/aaa2ppp/multgen/internal/api/buffer"
)

type Solver interface {
	Solve() float64
}

func New(s Solver) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /get", noCache(getHandler(s)))
	mux.Handle("GET /ping", noCache(http.HandlerFunc(pong)))
	return mux
}

func logWriteError(r *http.Request, err error) {
	log.Printf("%s %s: write body failed: %v", r.Method, r.URL.Path, err)
}

func pong(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("pong")); err != nil {
		logWriteError(r, err)
	}
}

func getHandler(s Solver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		multiplier := s.Solve()

		// one Get
		buf := buffer.Get()

		// The response is simple, so we may not use json package. It is for performance reasons.
		buf = append(buf, `{"result":`...)
		buf = strconv.AppendFloat(buf, multiplier, 'g', -1, 64)
		buf = append(buf, '}')

		w.Header().Set("content-type", "application/json")
		w.Header().Set("content-length", strconv.Itoa(len(buf)))

		if _, err := w.Write(buf); err != nil {
			logWriteError(r, err)
		}

		// one Put
		buffer.Put(buf)
	}
}
