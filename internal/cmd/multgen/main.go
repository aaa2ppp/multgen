// == internal/cmd/multgen/main.go ==

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
	"time"

	"multgen/internal/api"
	"multgen/internal/config"
	"multgen/internal/solver"
)

type Solver = solver.Solver

func Main() {
	cfg := config.MustLoad()
	solver := solver.New(cfg.Solver)

	var exitCode int
	if cfg.Server.Enable {
		exitCode = runAsHTTPServer(cfg.Server, solver)
		log.Printf("exit with code: %d", exitCode)
	} else {
		exitCode = runAsCLI(os.Stdin, os.Stdout, solver)
	}

	os.Exit(exitCode)
}

func runAsHTTPServer(cfg *config.Server, solver Solver) int {
	api := api.New(solver)

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
		signal.Notify(c, os.Interrupt)
		s := <-c

		log.Printf("shutdown by signal: %v", s)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("correct shutdown failed: %v", err)
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

func runAsCLI(in io.Reader, out io.Writer, solver Solver) int {
	var n int
	if _, err := fmt.Fscan(in, &n); err != nil {
		log.Printf("can't read n: %v", err)
		return 1
	}

	w := bufio.NewWriter(out)

	for i := 0; i < n; i++ {
		multiplier := solver.Solve()
		// skip the write error check for performance; check it on flush
		w.WriteString(strconv.FormatFloat(multiplier, 'f', 3, 64))
		w.WriteByte('\n')
	}

	if err := w.Flush(); err != nil {
		log.Printf("can't write: %v", err)
		return 1
	}

	return 0
}
