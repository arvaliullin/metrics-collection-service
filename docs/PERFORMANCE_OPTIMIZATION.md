# Отчет по оптимизации производительности

## Базовый профиль (base.pprof)

#### Результат выполнения бенчмарка

```bash
go test -bench=BenchmarkRepository_BatchUpdate_1000 -benchmem -benchtime=10s -count=3 -memprofile=bin/profiles/base.pprof ./internal/repository/memory/
goos: linux
goarch: amd64
pkg: github.com/arvaliullin/metrics-collection-service/internal/repository/memory
cpu: Intel(R) Core(TM) i5-4460  CPU @ 3.20GHz
BenchmarkRepository_BatchUpdate_1000-4            193119             65153 ns/op            8001 B/op       1000 allocs/op
BenchmarkRepository_BatchUpdate_1000-4            199262             65493 ns/op            8001 B/op       1000 allocs/op
BenchmarkRepository_BatchUpdate_1000-4            188800             61628 ns/op            8001 B/op       1000 allocs/op
PASS
ok      github.com/arvaliullin/metrics-collection-service/internal/repository/memory    39.163s
```

## Профиль после оптимизаций (result.pprof)

#### Результат выполнения бенчмарка

```bash
go test -bench=BenchmarkRepository_BatchUpdate_1000 -benchmem -benchtime=10s -count=3 -memprofile=bin/profiles/result.pprof ./internal/repository/memory/
goos: linux
goarch: amd64
pkg: github.com/arvaliullin/metrics-collection-service/internal/repository/memory
cpu: Intel(R) Core(TM) i5-4460  CPU @ 3.20GHz
BenchmarkRepository_BatchUpdate_1000-4            252829             49699 ns/op            4001 B/op        500 allocs/op
BenchmarkRepository_BatchUpdate_1000-4            250468             49995 ns/op            4001 B/op        500 allocs/op
BenchmarkRepository_BatchUpdate_1000-4            232635             50106 ns/op            4001 B/op        500 allocs/op
PASS
ok      github.com/arvaliullin/metrics-collection-service/internal/repository/memory    38.244s
```

## Результаты оптимизации

### Сравнение профилей

```bash
go tool pprof -top -diff_base=bin/profiles/base.pprof bin/profiles/result.pprof

File: memory.test
Build ID: 4796f2ae33c424355acf7b0cab2fe08d28a66671
Type: alloc_space
Time: 2026-01-06 01:11:44 +05
Showing nodes accounting for -1.70GB, 37.37% of 4.55GB total
Dropped 6 nodes (cum <= 0.02GB)
      flat  flat%   sum%        cum   cum%
   -1.70GB 37.37% 37.37%    -1.70GB 37.37%  github.com/arvaliullin/metrics-collection-service/internal/repository/memory.(*Repository).BatchUpdate
         0     0% 37.37%    -1.70GB 37.37%  github.com/arvaliullin/metrics-collection-service/internal/repository/memory_test.BenchmarkRepository_BatchUpdate_1000
         0     0% 37.37%    -1.70GB 37.33%  testing.(*B).launch
         0     0% 37.37%    -1.70GB 37.37%  testing.(*B).runN
```
