package multgen

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"

	fastapi "github.com/aaa2ppp/multgen/internal/api/fast"
	stdapi "github.com/aaa2ppp/multgen/internal/api/std"
	"github.com/aaa2ppp/multgen/internal/config"
	"github.com/aaa2ppp/multgen/internal/server"
	"github.com/aaa2ppp/multgen/internal/server/dummy"
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
	if cfg.CLIMode {
		exitCode = runAsCLI(os.Stdin, os.Stdout, solver)
	} else {
		exitCode = runAsHTTPServer(cfg.Server, solver)
		log.Printf("exit with code: %d", exitCode)
	}

	os.Exit(exitCode)
}

func runAsHTTPServer(cfg config.Server, solver *solver.Solver) int {
	const (
		readTimeout   = 5 * time.Second
		writeTimeout  = 5 * time.Second
		workerTimeout = 100 * time.Microsecond
	)

	var srv server.Server
	if cfg.DummyHTTP {
		srv = &dummy.Server{
			Solver:        solver,
			ReadTimeout:   readTimeout,
			WriteTimeout:  writeTimeout,
			WorkerTimeout: workerTimeout,
		}
	} else if cfg.FastHTTP {
		srv = &fasthttp.Server{
			Handler:      fastapi.New(solver),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		}
	} else {
		srv = &http.Server{
			Handler:      stdapi.New(solver),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		}
	}

	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Printf("can't start server: %v", err)
		return 1
	}

	_, stopServer := server.Start(srv, listener)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("shutdown by signal: %v", sig)

	if err := stopServer(); err != nil {
		log.Printf("stop server failed: %v", err)
		return 1
	}

	return 0
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
