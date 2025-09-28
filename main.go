// == main.go ==

// This file need only to satisfy the TS:
//
// "Сервис должен запускаться командой: `go run . -rtp={значение}`"
//
// Main entry point - cmd/multgen/main.go
package main

import (
	"github.com/aaa2ppp/multgen/pkg/app"
)

func main() {
	solver := app.DefaultSolverConfig()

	// TODO: Don't forget to configure solver!
	solver.Algorithm = "pareto1"

	app.Main(app.Config{
		Server: app.ServerConfig{
			Addr:   "localhost:64333",
		},
		Solver: solver,
	})
}
