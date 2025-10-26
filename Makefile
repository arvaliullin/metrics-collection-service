.PHONY: run
run:
	- ADDRESS=localhost:8082 go run github.com/arvaliullin/metrics-collection-service/cmd/server

.PHONY: test
test:
	- go test ./...

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

.PHONY: prune
prune: down
	- docker image prune -f
	- docker container prune -f
	- docker volume prune -f
	- docker network prune -f
	- docker system prune -a --volumes -f
