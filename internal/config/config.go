package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aaa2ppp/multgen/internal/solver"
)

type Config struct {
	CLIMode bool

	// Позволяет проанализировать поведение игрока, используя предопределённый RTP,
	// игнорируя значение флага -rtp
	IgnoreInputRTP bool

	Server Server
	Solver solver.Config
}

type Server struct {
	Addr     string
	FastHTTP bool
}

type Solver = solver.Config

func algorithmsHelp(msg string, algos []solver.Algorithm) string {
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
	var (
		help = flag.Bool("help", false, "show usage help")

		cliMode = flag.Bool("cli", tune.CLIMode, "cli mode:"+
			"\n- http server does not start;"+
			"\n- read one int N (sequence length) from stdin;"+
			"\n- write N multipliers to stdout")

		// Server flags
		serverAddr = flag.String("http", tune.Server.Addr, "http server address")
		fastHTTP   = flag.Bool("fast", tune.Server.FastHTTP, "use fasthttp instead of net/http")

		// Solver flags
		rtp       = flag.Float64("rtp", 0, "rtp must be in (0, 1] (required)")
		algorithm = flag.String("algo", tune.Solver.Algorithm, algorithmsHelp("algorithm for generating multipliers", solver.Algorithms))
		alpha     = flag.Float64("alpha", tune.Solver.Alpha, "alpha must be >= 1")
		addDelta  = flag.Bool("d", tune.Solver.AddDelta, "add delta to mulipliers")
	)

	flag.Parse()

	if *help {
		fmt.Fprint(os.Stderr, "Usage: multgen [options] -rtp=<value>\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *rtp == 0 {
		fmt.Fprintln(os.Stderr, "rtp is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if !tune.IgnoreInputRTP && !(0 < *rtp && *rtp <= 1) {
		fmt.Fprintf(os.Stderr, "rtp must be in (0, 1], got %v\n", *rtp)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if !(*alpha >= 1) {
		fmt.Fprintf(os.Stderr, "alpha must be >= 1, got %v", *alpha)
		flag.PrintDefaults()
		os.Exit(1)
	}

	tune.CLIMode = *cliMode

	tune.Server.Addr = *serverAddr
	tune.Server.FastHTTP = *fastHTTP

	if !tune.IgnoreInputRTP {
		tune.Solver.RTP = *rtp
	}
	tune.Solver.Algorithm = *algorithm
	tune.Solver.Alpha = *alpha
	tune.Solver.AddDelta = *addDelta

	return tune
}
