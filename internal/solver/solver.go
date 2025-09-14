// == internal/solver/solver.go ==

// Package solver provides multiplier generation algorithms.
//
// NOTE: Current implementation is a stub. Real algorithm development
// is postponed until requirements clarification from README_QUESTIONS.md
package solver

import "multgen/internal/config"

type Solver interface {
	Solve() float64
}

func newDefaultSolver(cfg *config.Solver) Solver {
	return &stubSolver{cfg}
}

// stubSolver always returns the MinMultiplier value (default is 1.0)
type stubSolver struct {
	cfg *config.Solver
}

func (s *stubSolver) Solve() float64 { return s.cfg.MinMultiplier }

func New(cfg *config.Solver) Solver {
	switch cfg.Algorithm {
	case "stub":
		return &stubSolver{cfg}
	default:
		return newDefaultSolver(cfg)
	}
}
