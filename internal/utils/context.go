package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// CtxWithSyscallHandler создает контекст, который отменяется при получении SIGINT или SIGTERM
func CtxWithSyscallHandler() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	go func() {
		<-sigChan
		cancel()
	}()

	return ctx, cancel
}
