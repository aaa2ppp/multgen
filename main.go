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

// tune application
var tune = app.Config{
	Server: app.ServerConfig{
		Addr:   "localhost:64333",
		Enable: true,
	},
	Solver: app.SolverConfig{
		Algorithm:     "v1",              // попробуй: v2, v3
		MinMultiplier: app.MinMultiplier, // сейчас не используется
		MaxMultiplier: app.MaxMultiplier, // для v3, попробуй уменьшать (при 1+delta вырождается в v2)
		K:             15,                // для v3, поробуй поднять до 30
	},
}

func main() {
	app.Main(tune)
}
