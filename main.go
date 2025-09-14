// == main.go ==

// This file need only to satisfy the TS:
//
// "Сервис должен запускаться командой: `go run . -rtp={значение}`"
//
// Main entry point - cmd/multgen/main.go
package main

import "multgen/internal/cmd/multgen"

func main() {
	multgen.Main()
}
