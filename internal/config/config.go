// == internal/config/config.go ==

package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	s "github.com/aaa2ppp/multgen/internal/solver"
)

type Config struct {
	CLIMode bool
	Server  Server
	Solver  s.Config
}

type Server struct {
	Addr     string
	FastHTTP bool
}

type Solver = s.Config

func algorithmsHelp(msg string, algos []s.Algorithm) string {
	var buf strings.Builder
	buf.WriteString(msg)
	buf.WriteString(":\n")
	for _, a := range algos {
		buf.WriteString(strconv.Quote(a.Name))
		if len(a.Name) < 6 {
			buf.WriteString("\t- ")
		} else {
			buf.WriteString(" - ")
		}
		buf.WriteString(strings.ReplaceAll(a.Description, "\n", "\n    "))
		buf.WriteString(";\n")
	}
	return buf.String()
}

func MustLoad(tune Config) Config {
	server := &tune.Server
	solver := &tune.Solver

	var (
		help = flag.Bool("help", false, "show usage help")

		cliMode = flag.Bool("cli", tune.CLIMode, "cli mode:"+
			"\n- http server does not start;"+
			"\n- read one int N (sequence length) from stdin;"+
			"\n- write N multipliers to stdout")

		// Server flags

		serverAddr = flag.String("http", server.Addr, "http server address")
		fastHTTP   = flag.Bool("fast", server.FastHTTP, "use fasthttp instead of net/http")

		// Solver flags

		rtp       = flag.Float64("rtp", 0, "rtp must be in (0, 1] (required)")
		algorithm = flag.String("algo", solver.Algorithm, algorithmsHelp("algorithm for generating multipliers", s.Algorithms))
		alpha     = flag.Float64("alpha", solver.Alpha, "alpha must be >= 1")
		addDelta  = flag.Bool("d", solver.AddDelta, "add delta to mulipliers")
	)

	flag.Parse()

	if *help {
		fmt.Fprint(os.Stderr, "Usage: multgen [options] -rtp=<value>\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *rtp == 0 {
		fmt.Fprintln(os.Stderr, "rtp is required")
		os.Exit(1)
	}

	tune.CLIMode = *cliMode

	server.Addr = *serverAddr
	server.FastHTTP = *fastHTTP

	solver.InputRTP = *rtp
	solver.Algorithm = *algorithm
	solver.Alpha = *alpha
	solver.AddDelta = *addDelta

	return tune
}
