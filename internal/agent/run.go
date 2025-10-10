package agent

import (
	"fmt"
	"time"
)

func (a *Agent) Run() error {
	a.cron.Do(func() { fmt.Println(time.Now()) })
	return nil
}
