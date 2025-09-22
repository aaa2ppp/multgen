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
	Server Server
	Solver s.Config
}

type Server struct {
	Addr   string
	Enable bool
}

type Solver = s.Config

func algorithmsHelp(msg string, algos []s.Algorithm) string {
	var buf strings.Builder
	buf.WriteString(msg)
	buf.WriteString(":\n")
	for _, a := range algos {
		buf.WriteString(strconv.Quote(a.Name))
		buf.WriteString(" - ")
		buf.WriteString(strings.ReplaceAll(a.Description, "\n", "\n    "))
		buf.WriteString(";\n")
	}
	return buf.String()
}

func required[T comparable](v T) string {
	var zero T
	if v == zero {
		return " (required)"
	}
	return ""
}

func MustLoad(tune Config) Config {
	var (
		help = flag.Bool("help", false, "show usage help")

		// Server flags

		serverAddr = flag.String("http", tune.Server.Addr, "http server address")
		cliMode    = flag.Bool("cli", !tune.Server.Enable, "cli mode:"+
			"\n- http server does not start;"+
			"\n- read one int N (sequence length) from stdin;"+
			"\n- write N multipliers to stdout")

		// Solver flags

		rtp           = flag.Float64("rtp", tune.Solver.RTP, fmt.Sprintf("rtp must be in (%g, %g]%s", s.MinRTP, s.MaxRTP, required(tune.Solver.RTP)))
		noCheckRTP    = flag.Bool("no-check-rtp", tune.Solver.NoCheckRTP, "disables rtp validation (for testing only)")
		minMultiplier = flag.Float64("min", tune.Solver.MinMultiplier, fmt.Sprintf("tune min multiplier value. must be in [%g, %g] (currently not used anywhere)", s.MinMultiplier, s.MaxMultiplier))
		maxMultiplier = flag.Float64("max", tune.Solver.MaxMultiplier, fmt.Sprintf("tune max multiplier value. must be in [%g, %g]", s.MinMultiplier, s.MaxMultiplier))
		algorithm     = flag.String("algo", tune.Solver.Algorithm, algorithmsHelp("algorithm for generating multipliers", s.Algorithms))
		k             = flag.Float64("k", tune.Solver.K, "k in the exp(-k*x), if applicable; must be > 0")
	)

	flag.Parse()

	if *help {
		fmt.Fprint(os.Stderr, "Usage: multgen [options] -rtp=<value>\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	solverCfg := Solver{
		RTP:           *rtp,
		NoCheckRTP:    *noCheckRTP,
		Algorithm:     *algorithm,
		MinMultiplier: *minMultiplier,
		MaxMultiplier: *maxMultiplier,
		K:             *k,
	}

	if err := solverCfg.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	return Config{
		Solver: solverCfg,
		Server: Server{
			Addr:   *serverAddr,
			Enable: !*cliMode,
		},
	}
}
