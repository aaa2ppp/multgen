// Package app — публичный фасад для использования приложения из другого модуля.
//
// Зачем существует:
//   - Тестовая платформа требует скопировать main.go и запустить его (в другом модуле).
//
// Что здесь:
//   - Алиасы на internal/config.Config, internal/solver.DefaultConfig etc.
//   - Функция Main — просто вызывает internal/cmd/multgen.Main.
//
// Никакой логики — только мост.
package app

import (
	"github.com/aaa2ppp/multgen/internal/cmd/multgen"
	"github.com/aaa2ppp/multgen/internal/config"
	"github.com/aaa2ppp/multgen/internal/solver"
)

type (
	Config       = config.Config
	ServerConfig = config.Server
	SolverConfig = config.Solver
)

func Main(tune Config)                  { multgen.Main(tune) }
func DefaultSolverConfig() SolverConfig { return solver.DefaultConfig() }
