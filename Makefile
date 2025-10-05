.PHONY: run
run:
	- go run github.com/arvaliullin/metrics-collection-service/cmd/server

.PHONY: test
test:
	- go test ./...

.PHONY: fmt
fmt:
	- go fmt ./...
