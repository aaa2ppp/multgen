
# Иследуем тестовую платформу

`-algo=max`
т.е. на выходе получаем среднеарифметическое ввода

```
Submission #146
failed
Submitted: 29.09.2025, 18:41:01
Processed: 29.09.2025, 18:46:06
Test Results:
Test Kind 0: RTP 4947.2227
✗ FAILED
Test Kind 1: RTP 1.1000
✗ FAILED
Test Kind 2: RTP 2.0000
✗ FAILED
Test Kind 3: RTP 10.0000
✗ FAILED
Test Kind 4: RTP 100.0000
✗ FAILED
Test Kind 5: RTP 1000.0000
✗ FAILED
Go Mod: module test go 1.24.2 require github.com/aaa2ppp/multgen v0.0.0-20250929145139-521a4c92cb6e requi...
Main Go: package main import ( "github.com/aaa2ppp/multgen/pkg/app" ) func main() { solver := app.Default...
```

**Выводы**:
- Судя по таймингам на каждый тест 1 мин.
- Первый тест - случайные данные в диаппазоне `[0, 10000]`, остальные - фиксированные значения равные `RTP` (`1.1`, `2`, `10`, `100`, `1000`).

**Учитывая**:
```
$ make bench-http
go test -bench . -benchmem ./internal/api/... ./internal/cmd/multgen/. 
?       github.com/aaa2ppp/multgen/internal/api/buffer  [no test files]
goos: windows
goarch: amd64
pkg: github.com/aaa2ppp/multgen/internal/api/fast
cpu: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz
Benchmark_getHandler-4          10822783               119.9 ns/op         8340574 rps         0 B/op          0 allocs/op
PASS
ok      github.com/aaa2ppp/multgen/internal/api/fast    1.500s
goos: windows
goarch: amd64
pkg: github.com/aaa2ppp/multgen/internal/api/std
cpu: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz
Benchmark_getHandler-4            732702              1523 ns/op            656531 rps      1136 B/op         13 allocs/op
PASS
ok      github.com/aaa2ppp/multgen/internal/api/std     1.991s
goos: windows
goarch: amd64
pkg: github.com/aaa2ppp/multgen/internal/cmd/multgen
cpu: Intel(R) Core(TM) i3-7100 CPU @ 3.90GHz
BenchmarkHTTPServer/std/single/without_keep_alive-4                 2794            429704 ns/op              2328 rps      6553 B/op         75 allocs/op
BenchmarkHTTPServer/std/single/with_keep_alive-4                   18342             64626 ns/op             15482 rps      2448 B/op         25 allocs/op
BenchmarkHTTPServer/std/parallel/without_keep_alive-4               4846            246133 ns/op              4063 rps      6580 B/op         74 allocs/op
BenchmarkHTTPServer/std/parallel/with_keep_alive-4                 39991             25470 ns/op             39263 rps      2449 B/op         24 allocs/op
BenchmarkHTTPServer/fast/single/without_keep_alive-4                2912            425266 ns/op              2353 rps      3267 B/op         37 allocs/op
BenchmarkHTTPServer/fast/single/with_keep_alive-4                  35152             36811 ns/op             27173 rps        25 B/op          1 allocs/op
BenchmarkHTTPServer/fast/parallel/without_keep_alive-4              5538            213671 ns/op              4681 rps      3294 B/op         37 allocs/op
BenchmarkHTTPServer/fast/parallel/with_keep_alive-4                66955             16410 ns/op             60963 rps        26 B/op          1 allocs/op
PASS
ok      github.com/aaa2ppp/multgen/internal/cmd/multgen 13.971s
```

Длина последовательность `27_000*60 = 1_620_000` (на одном ядре). Этого явно не достаточно для сходимости.

Провеки на `bin/check` показывают, что хорошую сходимость при `-algo=pareto1` на случайных данных получаем при `1e8`.

Получить сходимость на `1e4`, как сказано в ТЗ - нереально.
