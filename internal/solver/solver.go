// == internal/solver/solver.go ==

// Package solver provides multiplier generation algorithms.
//
// NOTE: Implementation covers multiple experimental algorithms for multiplier generation,
// as formal requirements are still pending (see README_QUESTIONS.md).
// Algorithms are provided "as is" for different theoretical scenarios.
package solver

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"strings"
)

const (
	MinRTP        = 0.0
	MaxRTP        = 1.0
	MinMultiplier = 1.0
	MaxMultiplier = 10000.0
	Delta         = 1e-9
)

type Config struct {
	RTP        float64
	NoCheckRTP bool // for test only - disables rtp validation
	Algorithm  string

	// algorithms tunings

	MinMultiplier float64 // currently not used anywhere
	MaxMultiplier float64 // used by algorithm v3
	K             float64 // used by algorithm v3 only
}

func (c Config) Validate() error {
	var errs []error

	if !c.NoCheckRTP && !(MinRTP < c.RTP && c.RTP <= MaxRTP) {
		errs = append(errs, fmt.Errorf("rtp value is incorrect: must be in (%g, %g], got %g", MinRTP, MaxRTP, c.RTP))
	}

	if !(MinMultiplier <= c.MinMultiplier && c.MinMultiplier <= MaxMultiplier) {
		errs = append(errs, fmt.Errorf("min multiplier value is incorrect: must be in [%g, %g], got %g", MinMultiplier, MaxMultiplier, c.MinMultiplier))
	}

	if !(MinMultiplier <= c.MaxMultiplier && c.MaxMultiplier <= MaxMultiplier) {
		errs = append(errs, fmt.Errorf("max multiplier value is incorrect: must be in [%g, %g], got %g", MinMultiplier, MaxMultiplier, c.MaxMultiplier))
	}

	if !(c.MinMultiplier <= c.MaxMultiplier) {
		errs = append(errs, fmt.Errorf("min multiplier value must be <= max multiplier value, got %g and %g", c.MinMultiplier, c.MaxMultiplier))
	}

	// TODO: remove or uncomment. I still don't know how harsh I am (now New has a fallback to stub)
	// algoNotFound := true
	// for _, possible := range Algorithms {
	// 	if strings.EqualFold(possible.Name, c.Algorithm) {
	// 		algoNotFound = false
	// 		break
	// 	}
	// }
	// if algoNotFound {
	// 	errs = append(errs, fmt.Errorf("unknown algorithm: %s", c.Algorithm))
	// }

	if c.K <= 0 {
		errs = append(errs, fmt.Errorf("k value is incorrect: must be > 0, got %g", c.K))
	}

	return errors.Join(errs...)
}

type solveFunc func(*Config) float64

type Algorithm struct {
	Name        string
	Description string
	solve       solveFunc
}

var note = fmt.Sprintf("\nNOTE: ANY algorithm is powerless against a sequence consisting only of values equal to %g."+
	"\n  In this case, the RTP is ALWAYS 0, regardless of the multiplier", MaxMultiplier)

var Algorithms = []Algorithm{
	{"v1", fmt.Sprintf("with probability=rtp returns %g, otherwise %g.", MaxMultiplier, MinMultiplier) +
		"\nEnsures convergence for sequence length > 10000 for the transform=x case for an arbitrary sequence." +
		note,
		func(cfg *Config) float64 {
			if p := rand.Float64(); p < cfg.RTP {
				return MaxMultiplier
			}
			return MinMultiplier
		},
	},
	{"v2", fmt.Sprintf("with probability=rtp returns %g, otherwise %g.", MinMultiplier+Delta, MinMultiplier) +
		"\nEnsures convergence on sequence length > 10000 for the transform=x and transform=x*m case" +
		"\nassuming the client is rational",
		func(cfg *Config) float64 {
			if p := rand.Float64(); p < cfg.RTP {
				return 1 + Delta
			}
			return 1
		},
	},
	{"v3", fmt.Sprintf("with probability=rtp/m returns m (computed multiplier), otherwise %g."+
		"\nThe probability that the multiplier will be set to the determined value is defined by exp(-k*x).", MinMultiplier) +
		"\nEnsures convergence on sequence length > 1e8 for the case transform=x*m assuming the client is rational." +
		"\nProvides a random distribution of the multiplier across the entire acceptable range." +
		fmt.Sprintf("\nNOTE: Distribution starts at %g; max multiplier value is may be used to cap tail", MinMultiplier),
		func(cfg *Config) float64 {
			x := rand.Float64()
			d := cfg.MaxMultiplier - MinMultiplier
			m := MinMultiplier + math.Exp(-cfg.K*x)*d

			prob := cfg.RTP / m
			if prob > 1.0 {
				prob = 1.0
			}

			if p := rand.Float64(); p < prob {
				return m
			}
			return MinMultiplier
		},
	},
	{"stub", fmt.Sprintf("always returns strictly %g", MinMultiplier),
		func(cfg *Config) float64 { return MinMultiplier },
	},
}

type Solver struct {
	cfg   Config
	solve solveFunc
}

func (s *Solver) Solve() float64 {
	return s.solve(&s.cfg)
}

func New(cfg Config) (*Solver, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var solve solveFunc
	for _, a := range Algorithms {
		if strings.EqualFold(a.Name, cfg.Algorithm) {
			solve = a.solve
			break
		}
	}

	// fallback to stub
	if solve == nil {
		n := len(Algorithms)
		a := Algorithms[n-1]
		log.Printf("instead of the unknown %q algorithm, %q algorithm will be used", cfg.Algorithm, a.Name)
		cfg.Algorithm = a.Name
		solve = a.solve
	}

	return &Solver{
		cfg:   cfg,
		solve: solve,
	}, nil
}
