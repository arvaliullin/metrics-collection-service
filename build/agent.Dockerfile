FROM golang:1.24.7-alpine

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/agent github.com/arvaliullin/metrics-collection-service/cmd/agent

CMD ["agent"]
