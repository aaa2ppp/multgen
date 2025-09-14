// == internal/api/api.go ==

package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Solver interface {
	Solve() float64
}

func New(s Solver) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /get", getHandler(s))
	mux.Handle("/ping", http.HandlerFunc(pong))
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

type getResponse struct {
	Result float64 `json:"result"`
}

func getHandler(s Solver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		multiplier := s.Solve()
		buf, _ := json.Marshal(getResponse{Result: multiplier}) // can't fail for struct literal

		w.Header().Set("content-type", "application/json")
		w.Header().Set("content-length", strconv.Itoa(len(buf)))

		if _, err := w.Write(buf); err != nil {
			logWriteError(r, err)
		}
	}
}
