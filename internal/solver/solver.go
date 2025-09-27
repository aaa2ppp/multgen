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
	MaxValue = 10000.0
)

type Config struct {
	RTP            float64 // required
	InputRTP       float64 //
	IgnoreInputRTP bool    // for test only
	Algorithm      string  //
	Alpha          float64 //
	AddDelta       bool    // добавить дельту к заначению мультипликатора
}

func (c Config) Validate() error {
	var errs []error

	if !(0 < c.RTP && c.RTP <= 1.0) {
		errs = append(errs, fmt.Errorf("rtp value is incorrect: must be in (0, 1], got %g", c.RTP))
	}

	if !(c.Alpha >= 1) {
		errs = append(errs, fmt.Errorf("alpha must be >= 1, got %g", c.Alpha))
	}

	return errors.Join(errs...)
}

type algoFunc func(*Config) float64

type Algorithm struct {
	Name        string
	Description string
	fn          algoFunc
}

func pareto1() float64 {
	u := rand.Float64()
	m := 1 / (1 - u)
	if m > MaxValue {
		m = MaxValue
	}
	return m
}

func paretoAlpha(alpha float64) float64 {
	u := rand.Float64()
	m := math.Pow(1-u, -1/alpha)
	if m > MaxValue {
		m = MaxValue
	}
	return m
}

// Algorithms выбора мультипликатора
var Algorithms = []Algorithm{
	{
		"pareto1",
		"честный (при любых x, матожидание RTP=1), но плохо сходится при больших x",
		func(_ *Config) float64 { return pareto1() },
	},
	{
		"paretoA",
		`"загоняем" игрока в x=1 (RTP падает с ростом x, при alpha > 1)`,
		func(cfg *Config) float64 { return paretoAlpha(cfg.Alpha) },
	},
	{
		"max",
		fmt.Sprintf("всегда возвращает %g", MaxValue),
		func(_ *Config) float64 { return MaxValue },
	},
	{
		"min",
		"всегда возвращает 1",
		func(_ *Config) float64 { return 1 },
	},
}

type Solver struct {
	cfg    Config
	algoFn algoFunc
}

func New(cfg Config) (*Solver, error) {
	if !cfg.IgnoreInputRTP {
		cfg.RTP = cfg.InputRTP
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var algoFn algoFunc
	for _, algo := range Algorithms {
		if strings.EqualFold(algo.Name, cfg.Algorithm) {
			algoFn = algo.fn
			break
		}
	}

	if algoFn == nil {
		a := defaultAlgorithm()
		log.Printf("instead of the unknown %q algorithm, %q algorithm will be used", cfg.Algorithm, a.Name)
		cfg.Algorithm = a.Name
		algoFn = a.fn
	}

	return &Solver{
		cfg:    cfg,
		algoFn: algoFn,
	}, nil
}

func (s *Solver) Solve() float64 {

	// забираем свою долю
	p := rand.Float64()
	if p > s.cfg.RTP {
		return 1
	}

	multiplier := s.algoFn(&s.cfg)

	if s.cfg.AddDelta {
		multiplier = math.Nextafter(multiplier, multiplier+1)
	}

	return multiplier
}
