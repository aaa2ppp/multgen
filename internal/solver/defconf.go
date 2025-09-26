package solver

func defaultAlgorithm() Algorithm {
	return Algorithms[len(Algorithms)-1]
}

func DefaultConfig() Config {
	return Config{
		RTP:           1.0,
		Algorithm:     defaultAlgorithm().Name,
		MinMultiplier: MinMultiplier,
		MaxMultiplier: MaxMultiplier,
		K:             1,
	}
}
