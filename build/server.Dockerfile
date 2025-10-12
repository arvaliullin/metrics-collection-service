FROM golang:1.24.7-alpine

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v -o /usr/local/bin/server github.com/arvaliullin/metrics-collection-service/cmd/server

CMD ["server"]
