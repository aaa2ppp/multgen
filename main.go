// == main.go ==

// Этот файл необходим только для запуска приложения на тестовой платформе и
// удовлетворения требования ТЗ:
//
//   - "Сервис должен запускаться командой: `go run . -rtp={значение}`"
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
			Addr:     "localhost:64333",
			FastHTTP: true,
		},
		Solver: solver,
	})
}
