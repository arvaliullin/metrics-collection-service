package main

import "github.com/arvaliullin/metrics-collection-service/internal/agent"

func main() {
	app := agent.New()
	if err := app.Run(); err != nil {
		panic(err)
	}
}
