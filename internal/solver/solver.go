// Package solver provides multiplier generation algorithms.
package solver

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"strings"
)

type algoFunc func(*Config) float64

type Algorithm struct {
	Name        string
	Description string
	fn          algoFunc
}

func pareto1(maxValue float64) float64 {
	u := rand.Float64()
	m := 1 / (1 - u)
	if m > maxValue {
		m = maxValue
	}
	return m
}

func paretoAlpha(alpha float64, maxValue float64) float64 {
	u := rand.Float64()
	m := math.Pow(1-u, -1/alpha)
	if m > maxValue {
		m = maxValue
	}
	return m
}

// Algorithms выбора мультипликатора
var Algorithms = []Algorithm{
	{
		"pareto1",
		`"честный" (при любых x, матожидание RTP=1), но плохо сходится при больших x`,
		func(cfg *Config) float64 { return pareto1(cfg.MaxValue) },
	},
	{
		"paretoA",
		`"загоняем" игрока в x=1 (RTP падает с ростом x, при alpha > 1)`,
		func(cfg *Config) float64 { return paretoAlpha(cfg.Alpha, cfg.MaxValue) },
	},
	{
		"const",
		"всегда возвращает константу",
		func(cfg *Config) float64 { return cfg.Value },
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

func defaultAlgorithm() Algorithm {
	return Algorithms[len(Algorithms)-1] // min
}

type Solver struct {
	cfg    Config
	algoFn algoFunc
}

func New(cfg Config) (*Solver, error) {
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
