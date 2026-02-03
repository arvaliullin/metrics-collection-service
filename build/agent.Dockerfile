FROM golang:1.24.7-alpine

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BUILD_VERSION
ARG BUILD_DATE
ARG BUILD_COMMIT
RUN CGO_ENABLED=0 GOOS=linux go build -v \
	-ldflags "-X main.buildVersion=$BUILD_VERSION -X main.buildDate=$BUILD_DATE -X main.buildCommit=$BUILD_COMMIT" \
	-o /usr/local/bin/agent github.com/arvaliullin/metrics-collection-service/cmd/agent

CMD ["agent"]
