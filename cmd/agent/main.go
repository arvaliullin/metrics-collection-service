package main

import "github.com/arvaliullin/metrics-collection-service/internal/agent"

func main() {
	app := agent.New()
	app.Run()
}
