package checker

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// ConfidenceInterval возвращает среднее и границы доверительного интервала.
// confidenceLevel — например, 0.95 для 95%.
func ConfidenceInterval(values []float64, confidenceLevel float64) (mean, lower, upper float64, err error) {
	n := len(values)
	if n == 0 {
		return 0, 0, 0, fmt.Errorf("empty input slice")
	}
	if confidenceLevel <= 0 || confidenceLevel >= 1 {
		return 0, 0, 0, fmt.Errorf("confidenceLevel must be between 0 and 1")
	}

	mean = stat.Mean(values, nil)
	if n == 1 {
		return mean, mean, mean, nil
	}

	variance := stat.Variance(values, nil)
	stdDev := math.Sqrt(variance)
	stdErr := stdDev / math.Sqrt(float64(n))

	alpha := 1.0 - confidenceLevel
	p := 1.0 - alpha/2 // например, 0.975 для 95%

	var criticalValue float64
	if n <= 30 {
		// t-распределение
		tDist := distuv.StudentsT{
			Mu:    0,
			Sigma: 1,
			Nu:    float64(n - 1),
		}
		criticalValue = tDist.Quantile(p)
	} else {
		// Нормальное распределение
		normal := distuv.Normal{Mu: 0, Sigma: 1}
		criticalValue = normal.Quantile(p)
		// Или: criticalValue = mathext.NormalQuantile(p) — тоже работает
	}

	margin := criticalValue * stdErr
	lower = mean - margin
	upper = mean + margin

	return mean, lower, upper, nil
}

// Пример использования
// func main() {
// 	values := []float64{10.2, 9.8, 10.1, 10.3, 9.9}
// 	mean, lo, hi, err := ConfidenceInterval(values, 0.95)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("Mean: %.4f\n", mean)
// 	fmt.Printf("95%% CI: [%.4f, %.4f]\n", lo, hi)
// }
