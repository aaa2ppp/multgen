// == internal/api/api.go ==

package api

import (
	"log"
	"net/http"
	"strconv"
)

type Solver interface {
	Solve() float64
}

func New(s Solver) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /get", getHandler(s))
	mux.Handle("GET /ping", http.HandlerFunc(pong))
	return noCache(mux)
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
		buf := bufPool.Get()

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
		bufPool.Put(buf)
	}
}
