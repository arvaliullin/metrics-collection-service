.PHONY: run
run:
	- go run github.com/arvaliullin/metrics-collection-service/cmd/server

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
	- curl -v -X POST http://localhost:8080/update/counter/someMetric/527

.PHONY: curl-err
curl-err:
	- curl  -X POST http://localhost:8080/update/undef/someMetric/527
