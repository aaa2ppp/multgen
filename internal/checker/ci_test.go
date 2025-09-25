package checker

import (
	"math"
	"testing"

	"github.com/aaa2ppp/be"
	"gonum.org/v1/gonum/stat"
)

func TestConfidenceInterval(t *testing.T) {
	t.Run("small sample, 95% CI", func(t *testing.T) {
		values := []float64{10.2, 9.8, 10.1, 10.3, 9.9}
		mean, lower, upper, err := ConfidenceInterval(values, 0.95)
		be.Err(t, err, nil)
		be.True(t, mean > 0)
		be.True(t, lower < mean && mean < upper)
		be.True(t, upper-lower > 0)
	})

	t.Run("large sample, 95% CI", func(t *testing.T) {
		// Генерируем 100 значений вокруг 50 с небольшим шумом
		values := make([]float64, 100)
		for i := range values {
			values[i] = 50 + (float64(i%10)-4.5)*0.1 // детерминированный "шум"
		}
		mean, lower, upper, err := ConfidenceInterval(values, 0.95)
		be.Err(t, err, nil)
		be.True(t, math.Abs(mean-50) < 0.1)
		be.True(t, lower < mean && mean < upper)
		be.True(t, upper-lower > 0)
	})

	t.Run("empty input", func(t *testing.T) {
		_, _, _, err := ConfidenceInterval([]float64{}, 0.95)
		be.Err(t, err)
	})

	t.Run("invalid confidence level: >1", func(t *testing.T) {
		_, _, _, err := ConfidenceInterval([]float64{1, 2}, 1.1)
		be.Err(t, err)
	})

	t.Run("invalid confidence level: <=0", func(t *testing.T) {
		_, _, _, err := ConfidenceInterval([]float64{1, 2}, 0.0)
		be.Err(t, err)
	})

	t.Run("single value", func(t *testing.T) {
		mean, lower, upper, err := ConfidenceInterval([]float64{42.0}, 0.95)
		be.Err(t, err, nil)
		be.Equal(t, mean, 42.0)
		be.Equal(t, lower, 42.0)
		be.Equal(t, upper, 42.0)
	})

	t.Run("99% CI wider than 95% CI", func(t *testing.T) {
		values := []float64{1, 2, 3, 4, 5}
		_, lo95, hi95, err := ConfidenceInterval(values, 0.95)
		be.Err(t, err, nil)
		_, lo99, hi99, err := ConfidenceInterval(values, 0.99)
		be.Err(t, err, nil)

		width95 := hi95 - lo95
		width99 := hi99 - lo99
		be.True(t, width99 > width95)
	})
}

// Вспомогательная функция для сравнения float64 с точностью
func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestSmallSampleCases(t *testing.T) {
	// Случай 1: Известный пример из учебника (n=5)
	t.Run("textbook example n=5", func(t *testing.T) {
		// Данные: [1, 2, 3, 4, 5]
		// Среднее = 3, s ≈ 1.5811, SE ≈ 0.7071
		// t_{0.975, 4} ≈ 2.776 → margin ≈ 1.963
		// CI ≈ [1.037, 4.963]
		values := []float64{1, 2, 3, 4, 5}
		mean, lo, hi, err := ConfidenceInterval(values, 0.95)
		be.Err(t, err, nil)
		be.True(t, almostEqual(mean, 3.0, 1e-10))
		be.True(t, almostEqual(lo, 1.037, 0.01))
		be.True(t, almostEqual(hi, 4.963, 0.01))
	})

	// Случай 2: Все значения одинаковые → CI = [x, x]
	t.Run("constant small sample", func(t *testing.T) {
		values := []float64{7.5, 7.5, 7.5, 7.5}
		mean, lo, hi, err := ConfidenceInterval(values, 0.95)
		be.Err(t, err, nil)
		be.Equal(t, mean, 7.5)
		be.Equal(t, lo, 7.5)
		be.Equal(t, hi, 7.5)
	})

	// Случай 3: Очень маленькая выборка (n=2)
	t.Run("n=2 sample", func(t *testing.T) {
		values := []float64{10, 20}
		mean, lo, hi, err := ConfidenceInterval(values, 0.95)
		be.Err(t, err, nil)
		be.True(t, mean == 15)
		be.True(t, lo < 15 && hi > 15)
		// t_{0.975,1} ≈ 12.706 → SE = 5 / sqrt(2) ≈ 3.535 → margin ≈ 44.9 → CI очень широкий
		be.True(t, hi-lo > 80) // грубая проверка ширины
	})

	// Случай 4: 99% CI шире, чем 95% для той же выборки
	t.Run("99% CI wider than 95% (small n)", func(t *testing.T) {
		values := []float64{2.1, 2.3, 1.9, 2.0, 2.2}
		_, lo95, hi95, err := ConfidenceInterval(values, 0.95)
		be.Err(t, err, nil)
		_, lo99, hi99, err := ConfidenceInterval(values, 0.99)
		be.Err(t, err, nil)
		be.True(t, (hi99-lo99) > (hi95-lo95))
	})
}

func TestLargeSampleCases(t *testing.T) {
	// Случай 1: Большая выборка вокруг 100, малый шум
	t.Run("large sample around 100", func(t *testing.T) {
		n := 1000
		values := make([]float64, n)
		for i := range values {
			values[i] = 100 + (float64(i%20)-9.5)*0.01 // шум ±0.095, среднее = 100
		}
		mean, lo, hi, err := ConfidenceInterval(values, 0.95)
		be.Err(t, err, nil)
		be.True(t, almostEqual(mean, 100.0, 1e-3))
		be.True(t, lo < 100 && hi > 100)
		// Для n=1000 и малого шума CI должен быть узким
		be.True(t, hi-lo < 0.01)
	})

	// Случай 2: Все значения одинаковые → CI = [x, x]
	t.Run("constant large sample", func(t *testing.T) {
		values := make([]float64, 100)
		for i := range values {
			values[i] = 42.0
		}
		mean, lo, hi, err := ConfidenceInterval(values, 0.99)
		be.Err(t, err, nil)
		be.Equal(t, mean, 42.0)
		be.Equal(t, lo, 42.0)
		be.Equal(t, hi, 42.0)
	})

	// Случай 3: Сравнение с нормальным приближением
	// При больших n t ≈ z, поэтому CI почти совпадает
	t.Run("t vs z negligible for large n", func(t *testing.T) {
		values := []float64{1, 2, 3, 4, 5}
		// Расширим до 100 элементов, повторяя
		large := make([]float64, 100)
		for i := range large {
			large[i] = values[i%5]
		}
		mean, loLarge, hiLarge, err := ConfidenceInterval(large, 0.95)
		be.Err(t, err, nil)

		// Посчитаем вручную через z = 1.96
		stdDev := math.Sqrt(stat.Variance(large, nil))
		se := stdDev / math.Sqrt(100.0)
		marginZ := 1.96 * se
		loZ := mean - marginZ
		hiZ := mean + marginZ

		// Разница между t и z должна быть мала
		be.True(t, math.Abs(loLarge-loZ) < 0.01)
		be.True(t, math.Abs(hiLarge-hiZ) < 0.01)
	})

	// Случай 4: 90% CI уже, чем 95%
	t.Run("90% CI narrower than 95% (large n)", func(t *testing.T) {
		values := make([]float64, 200)
		for i := range values {
			values[i] = 5.0 + math.Sin(float64(i)*0.1) // детерминированный шум
		}
		_, lo90, hi90, err := ConfidenceInterval(values, 0.90)
		be.Err(t, err, nil)
		_, lo95, hi95, err := ConfidenceInterval(values, 0.95)
		be.Err(t, err, nil)
		be.True(t, (hi90-lo90) < (hi95-lo95))
	})
}
