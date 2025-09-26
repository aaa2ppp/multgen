// == cmd/multgen/main.go ==

package main

import (
	"github.com/aaa2ppp/multgen/internal/cmd/multgen"
	"github.com/aaa2ppp/multgen/internal/config"
	"github.com/aaa2ppp/multgen/internal/solver"
)

// tune application
var tune = config.Config{
	Server: config.Server{
		Addr:   "localhost:64333",
		Enable: true,
	},
	Solver: solver.DefaultConfig(),
}

func main() {
	multgen.Main(tune)
}
