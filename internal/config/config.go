// == internal/config/config.go ==

package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	minRTP               = 0.0
	maxRTP               = 1.0
	defaultMinMultiplier = 1.0
	defaultMaxMultiplier = 10000.0
	defaultMinSeqValue   = 1.0
	defaultMaxSeqValue   = 10000.0
	defautlServerAddr    = "localhost:64333"
)

var (
	// Solver
	rtp           = flag.Float64("rtp", 0, fmt.Sprintf("rtp must be in (%f ... %f] (required)", minRTP, maxRTP))
	anyRTP        = flag.Bool("anyrtp", false, "disable the rtp check")
	minMultiplier = flag.Float64("minmult", defaultMinMultiplier, "min multiplier")
	maxMultiplier = flag.Float64("maxmult", defaultMaxMultiplier, "max multiplier")
	minSeqValue   = flag.Float64("minseq", defaultMinSeqValue, "min sequence value")
	maxSeqValue   = flag.Float64("maxseq", defaultMaxSeqValue, "max sequence value")
	algorithm     = flag.String("algo", "", "algorithm for generating multipliers")

	// Server
	serverAddr = flag.String("http", defautlServerAddr, "http server addr")
	cliMode    = flag.Bool("cli", false, "cli mode: http server does not start; read one int N (sequence length) from stdin; write N multipliers to stdout")
)

type Config struct {
	Solver *Solver
	Server *Server
}

type Solver struct {
	RTP           float64
	MinMultiplier float64
	MaxMultiplier float64
	MinSeqValue   float64
	MaxSeqValue   float64
	Algorithm     string
}

type Server struct {
	Addr   string
	Enable bool
}

func MustLoad() *Config {
	flag.Parse()

	// rtp must be in (minRTP, maxRTP] i.e. (0, 1.0]
	if !*anyRTP && !(minRTP < *rtp && *rtp <= maxRTP) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *minMultiplier > *maxMultiplier {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *minSeqValue > *maxSeqValue {
		flag.PrintDefaults()
		os.Exit(1)
	}

	return &Config{
		Solver: &Solver{
			RTP:           *rtp,
			MinMultiplier: *minMultiplier,
			MaxMultiplier: *maxMultiplier,
			MinSeqValue:   *minSeqValue,
			MaxSeqValue:   *maxSeqValue,
			Algorithm:     *algorithm,
		},
		Server: &Server{
			Addr:   *serverAddr,
			Enable: !*cliMode,
		},
	}
}
