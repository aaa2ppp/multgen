
# Исследуем тестовую платформу

`-algo=max`
т.е. на выходе получаем среднеарифметическое последовательности

```
Submission #146
failed
Submitted: 29.09.2025, 18:41:01
Processed: 29.09.2025, 18:46:06
Test Results:
Test Kind 0: RTP 4947.2227 ✗ FAILED
Test Kind 1: RTP 1.1000    ✗ FAILED
Test Kind 2: RTP 2.0000    ✗ FAILED
Test Kind 3: RTP 10.0000   ✗ FAILED
Test Kind 4: RTP 100.0000  ✗ FAILED
Test Kind 5: RTP 1000.0000 ✗ FAILED
```

**Выводы**:
- Судя по таймингам на каждый тест 1 мин.
- Первый тест - случайные данные в диапазоне `[0, 10000]`, остальные - фиксированные значения равные `RTP` (`1.1`, `2`, `10`, `100`, `1000`).

**Учитывая**:
```
BenchmarkHTTPServer/std/single/with_keep_alive-4                   18342             64626 ns/op             15482 rps      2448 B/op         25 allocs/op
BenchmarkHTTPServer/std/parallel/with_keep_alive-4                 39991             25470 ns/op             39263 rps      2449 B/op         24 allocs/op
BenchmarkHTTPServer/fast/single/with_keep_alive-4                  35152             36811 ns/op             27173 rps        25 B/op          1 allocs/op
BenchmarkHTTPServer/fast/parallel/with_keep_alive-4                66955             16410 ns/op             60963 rps        26 B/op          1 allocs/op
```

Длина последовательность `27_000*60 = 1_620_000` (на одном ядре). Этого явно недостаточно для сходимости.

Проверки на `bin/check` показывают, что хорошую сходимость при `-algo=pareto1` на случайных данных получаем при длине последовательности `1e8`.

Получить сходимость на `1e4`, как сказано в ТЗ - нереально.
