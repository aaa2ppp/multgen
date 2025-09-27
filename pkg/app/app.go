// == pkg/app/app.go ==

// Package app — публичный фасад для использования приложения из другого модуля.
//
// Зачем существует:
// - Тестовая платформа требует скопировать main.go и запустить его в другом модуле.
// - Go запрещает импортировать internal/ из другого модуля.
// - Значит, нужен публичный пакет — pkg/app — который экспортирует типы и Main из internal/.
//
// Что здесь:
// - Алиасы на internal/config.Config, internal/config.Server, internal/solver.Config.
// - Константы из internal/solver (MinRTP, MinMultiplier и т.д.).
// - Функция Main — просто вызывает internal/cmd/multgen.Main.
//
// Никакой логики — только мост.
//
// Не нравится? — Таковы ограничения Go + требования платформы.
// Патчи принимаются, если сохранят:
// - go run . из корня
// - возможность скопировать main.go на платформу
// - удобную настройку алгоритмов прямо в main.go
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
