package solver

import (
	"errors"
	"fmt"
)

const (
	MaxValue = 10000.0
)

type Config struct {
	RTP       float64 // целевой RTP
	Algorithm string  // алгоритма генерации RTP
	Alpha     float64 // параметр алгоритма paretoAlpha
	AddDelta  bool    // добавить дельту к заначению мультипликатора (имеет смысл для алгоритма min)
	Value     float64
	MaxValue  float64
}

func (c Config) Validate() error {
	var errs []error

	if !(0 < c.RTP && c.RTP <= 1.0) {
		errs = append(errs, fmt.Errorf("rtp value is incorrect: must be in (0, 1], got %g", c.RTP))
	}

	if !(c.Alpha >= 1) {
		errs = append(errs, fmt.Errorf("alpha must be >= 1, got %g", c.Alpha))
	}

	if !(1 <= c.Value && c.Value <= MaxValue) {
		errs = append(errs, fmt.Errorf("value must be in [1, %g], got %g", MaxValue, c.Value))
	}

	if !(1 <= c.MaxValue && c.MaxValue <= MaxValue) {
		errs = append(errs, fmt.Errorf("max value must be in [1, %g], got %g", MaxValue, c.Value))
	}

	return errors.Join(errs...)
}

func DefaultConfig() Config {
	return Config{
		RTP:       1,
		Algorithm: defaultAlgorithm().Name,
		Alpha:     1,
		Value:     1,
		MaxValue:  MaxValue,
	}
}
