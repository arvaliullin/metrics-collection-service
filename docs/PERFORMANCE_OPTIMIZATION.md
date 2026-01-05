# Отчет по оптимизации производительности

## Базовый профиль (base.pprof)

#### Результат выполнения бенчмарка

```bash
go test -bench=BenchmarkRepository_BatchUpdate_1000 -benchmem -benchtime=10s -count=3 -memprofile=bin/profiles/base.pprof ./internal/repository/memory/
goos: linux
goarch: amd64
pkg: github.com/arvaliullin/metrics-collection-service/internal/repository/memory
cpu: Intel(R) Core(TM) i5-4460  CPU @ 3.20GHz
BenchmarkRepository_BatchUpdate_1000-4            238098             47842 ns/op            4001 B/op        500 allocs/op
BenchmarkRepository_BatchUpdate_1000-4            243535             49459 ns/op            4001 B/op        500 allocs/op
BenchmarkRepository_BatchUpdate_1000-4            244258             48349 ns/op            4001 B/op        500 allocs/op
PASS
ok      github.com/arvaliullin/metrics-collection-service/internal/repository/memory    36.762s
```

## Профиль после оптимизаций (result.pprof)

#### Результат выполнения бенчмарка

```bash
go test -bench=BenchmarkRepository_BatchUpdate_1000 -benchmem -benchtime=10s -count=3 -memprofile=bin/profiles/result.pprof ./internal/repository/memory/
goos: linux
goarch: amd64
pkg: github.com/arvaliullin/metrics-collection-service/internal/repository/memory
cpu: Intel(R) Core(TM) i5-4460  CPU @ 3.20GHz
BenchmarkRepository_BatchUpdate_1000-4            504950             22997 ns/op               0 B/op          0 allocs/op
BenchmarkRepository_BatchUpdate_1000-4            518244             22892 ns/op               0 B/op          0 allocs/op
BenchmarkRepository_BatchUpdate_1000-4            527956             22897 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/arvaliullin/metrics-collection-service/internal/repository/memory    36.279s
```

### Сравнение профилей

```bash
go tool pprof -top -diff_base=bin/profiles/base.pprof bin/profiles/result.pprof

File: memory.test
Build ID: 81a1062e08d8f275a83ca3ef97709296b5dba0dc
Type: alloc_space
Time: 2026-01-06 02:07:49 +05
Showing nodes accounting for -2.84GB, 99.90% of 2.84GB total
Dropped 3 nodes (cum <= 0.01GB)
      flat  flat%   sum%        cum   cum%
   -2.84GB 99.90% 99.90%    -2.84GB 99.90%  github.com/arvaliullin/metrics-collection-service/internal/repository/memory.(*Repository).updateCounter
         0     0% 99.90%    -2.83GB 99.88%  github.com/arvaliullin/metrics-collection-service/internal/repository/memory.(*Repository).BatchUpdate
         0     0% 99.90%    -2.83GB 99.86%  github.com/arvaliullin/metrics-collection-service/internal/repository/memory_test.BenchmarkRepository_BatchUpdate_1000
         0     0% 99.90%    -2.83GB 99.81%  testing.(*B).launch
         0     0% 99.90%    -2.83GB 99.86%  testing.(*B).runN
```
