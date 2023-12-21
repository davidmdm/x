package main

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/davidmdm/conf"
	"github.com/davidmdm/x/xcontext"
)

type Config struct{}

func main() {
	var (
		rootTimeout time.Duration
		useCancel   bool
	)

	conf.Var(conf.Environ, &rootTimeout, "ROOT_TIMEOUT")
	conf.Var(conf.Environ, &useCancel, "USE_CANCEL")
	conf.Environ.MustParse()

	ctx, cancel := func() (context.Context, context.CancelFunc) {
		if rootTimeout == 0 {
			return context.Background(), func() {}
		}
		return context.WithTimeout(context.Background(), rootTimeout)
	}()

	defer cancel()

	ctx, cancel = xcontext.WithSignalCancelation(ctx, syscall.SIGINT)
	defer cancel()

	if useCancel {
		cancel()
	}

	<-ctx.Done()

	fmt.Println(context.Cause(ctx))
}
