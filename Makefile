.PHONY: build
build:
	mkdir -p bin
	go build -o bin/agent github.com/arvaliullin/metrics-collection-service/cmd/agent
	go build -o bin/server github.com/arvaliullin/metrics-collection-service/cmd/server

.PHONY: run
run:
	- go run github.com/arvaliullin/metrics-collection-service/cmd/server

.PHONY: test
test:
	- go test ./...

.PHONY: install-deps
install-deps:
	- go install github.com/golang/mock/mockgen@v1.6.0
	- go install github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: generate-mocks
generate-mocks:
	go generate ./...

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
	- docker-compose up --build -d

.PHONY: down
down:
	- docker-compose down -v

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
	- docker-compose logs

.PHONY: migration-create
migration-create:
	goose -dir migrations -s create create_metrics sql
