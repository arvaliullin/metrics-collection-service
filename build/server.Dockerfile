FROM golang:1.24.7-alpine

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
RUN GOBIN=/usr/local/bin go install github.com/pressly/goose/v3/cmd/goose@v3.18.0

COPY . .

RUN go build -v -o /usr/local/bin/server github.com/arvaliullin/metrics-collection-service/cmd/server

CMD ["server"]
