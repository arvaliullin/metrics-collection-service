.PHONY: build
build:
	mkdir -p bin
	go build -o bin/agent github.com/arvaliullin/metrics-collection-service/cmd/agent
	go build -o bin/server github.com/arvaliullin/metrics-collection-service/cmd/server

.PHONY: staticlint
staticlint:
	mkdir -p bin && go build -o bin/staticlint ./cmd/staticlint && go vet -vettool=./bin/staticlint ./...

.PHONY: run
run:
	- go run github.com/arvaliullin/metrics-collection-service/cmd/server

.PHONY: test
test:
	- go test ./...

.PHONY: bench
bench:
	go test -bench=. -benchmem ./internal/repository/memory/

.PHONY: bench-profile-base
bench-profile-base:
	mkdir -p bin/profiles
	go test -bench=BenchmarkRepository_BatchUpdate_1000 -benchmem -benchtime=10s -count=3 -memprofile=bin/profiles/base.pprof ./internal/repository/memory/
	@echo "Benchmark memory profile saved to bin/profiles/base.pprof"

.PHONY: bench-profile-result
bench-profile-result:
	mkdir -p bin/profiles
	go test -bench=BenchmarkRepository_BatchUpdate_1000 -benchmem -benchtime=10s -count=3 -memprofile=bin/profiles/result.pprof ./internal/repository/memory/
	@echo "Benchmark memory profile saved to bin/profiles/result.pprof"

.PHONY: bench-profile-diff
bench-profile-diff:
	@if [ ! -f bin/profiles/base.pprof ] || [ ! -f bin/profiles/result.pprof ]; then \
		echo "Error: Both bin/profiles/base.pprof and bin/profiles/result.pprof must exist"; \
		exit 1; \
	fi
	go tool pprof -top -diff_base=bin/profiles/base.pprof bin/profiles/result.pprof

.PHONY: bench-profile-view
bench-profile-view:
	@if [ ! -f bin/profiles/base.pprof ] || [ ! -f bin/profiles/result.pprof ]; then \
		echo "Error: Both bin/profiles/base.pprof and bin/profiles/result.pprof must exist"; \
		exit 1; \
	fi
	go tool pprof -http=":9090" -diff_base=bin/profiles/base.pprof bin/profiles/result.pprof

.PHONY: install-deps
install-deps:
	- go install go.uber.org/mock/mockgen@latest
	- go install github.com/pressly/goose/v3/cmd/goose@latest
	- go install github.com/swaggo/swag/cmd/swag@latest
	- go install -v golang.org/x/tools/cmd/godoc@latest
	- go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	- go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: generate-mocks
generate-mocks:
	go generate ./...

.PHONY: swag
swag:
	swag init -g cmd/server/main.go

.PHONY: fmt
fmt:
	- go fmt ./...

.PHONY: agent
agent:
	- go run github.com/arvaliullin/metrics-collection-service/cmd/agent

.PHONY: curl
curl:
	- curl -v -X POST http://localhost:8080/update/gauge/Alloc/1435720.000000

.PHONY: curl-err
curl-err:
	- curl  -X POST http://localhost:8080/update/undef/someMetric/527

.PHONY: curl-get
curl-get:
	- curl  -X GET http://localhost:8080/value/counter/PollCount


.PHONY: up
up:
	- docker compose up --build -d

.PHONY: pprof
pprof:
	go tool pprof -http=":9090" -seconds=600 http://localhost:8080/debug/pprof/profile

.PHONY: godoc
godoc:
	godoc -http=:6060 -play

.PHONY: down
down:
	- docker compose down -v

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: prune
prune: down
	- docker image prune -f
	- docker container prune -f
	- docker volume prune -f
	- docker network prune -f
	- docker system prune -a --volumes -f

.PHONY: logs
logs:
	- docker compose logs

.PHONY: migration-create
migration-create:
	goose -dir migrations -s create create_metrics sql

.PHONY: psql
psql:
	- PGPASSWORD=postgres_password psql -h localhost -p 5432 -U postgres_user -d postgres_db
