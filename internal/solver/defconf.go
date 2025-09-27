package solver

func defaultAlgorithm() Algorithm {
	return Algorithms[len(Algorithms)-1] // min
}

func DefaultConfig() Config {
	return Config{
		RTP:       1,
		InputRTP:  1,
		Algorithm: defaultAlgorithm().Name,
		Alpha:     1,
	}
}
