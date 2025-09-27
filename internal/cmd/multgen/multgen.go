// == internal/cmd/multgen/multgen.go ==

package multgen

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aaa2ppp/multgen/internal/api"
	"github.com/aaa2ppp/multgen/internal/config"
	"github.com/aaa2ppp/multgen/internal/solver"
)

func Main(tune config.Config) {
	cfg := config.MustLoad(tune)
	log.Printf("cfg: %+v", cfg)

	solver, err := solver.New(cfg.Solver)
	if err != nil {
		log.Fatalf("can't create solver: %v", err)
	}

	var exitCode int
	if cfg.Server.Enable {
		exitCode = runAsHTTPServer(cfg.Server, solver)
		log.Printf("exit with code: %d", exitCode)
	} else {
		exitCode = runAsCLI(os.Stdin, os.Stdout, solver)
	}

	os.Exit(exitCode)
}

func runAsHTTPServer(cfg config.Server, s *solver.Solver) int {
	api := api.New(s)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      api,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	done := make(chan int)
	go func() {
		defer close(done)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		s := <-c

		log.Printf("shutdown by signal: %v", s)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			done <- 1
		}
	}()

	log.Printf("http server listens on %v", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("http server fail: %v", err)
		return 1
	}

	return <-done
}

func runAsCLI(in io.Reader, out io.Writer, s *solver.Solver) int {
	var n int
	if _, err := fmt.Fscan(in, &n); err != nil {
		log.Printf("can't read n: %v", err)
		return 1
	}

	w := bufio.NewWriter(out)

	for i := 0; i < n; i++ {
		multiplier := s.Solve()
		b := w.AvailableBuffer()
		b = strconv.AppendFloat(b, multiplier, 'g', -1, 64)
		b = append(b, '\n')
		w.Write(b) // skip the write error check for performance; check it on flush
	}

	if err := w.Flush(); err != nil {
		log.Printf("can't write: %v", err)
		return 1
	}

	return 0
}
