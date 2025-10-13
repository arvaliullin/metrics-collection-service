package agent

import "sync"

func (a *Agent) Run() {

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		a.Poll()
	}()

	go func() {
		defer wg.Done()
		a.Report()
	}()

	wg.Wait()
}
