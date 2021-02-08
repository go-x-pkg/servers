package servers

import (
	"context"
	"time"

	"github.com/go-x-pkg/log"
)

type args struct {
	fnShutdownTimeout func() time.Duration

	fnLog log.FnT

	ctx context.Context
}

func (cfg *args) defaultize() {
	cfg.fnShutdownTimeout = defaultFnShutdownTimeout
	cfg.fnLog = defaultFnLog
	cfg.ctx = context.TODO()
}

type arg func(*args)

func FnShutdownTimeout(v func() time.Duration) arg {
	return func(cfg *args) { cfg.fnShutdownTimeout = v }
}

func FnLog(v log.FnT) arg {
	return func(cfg *args) { cfg.fnLog = v }
}

func Context(v context.Context) arg {
	return func(cfg *args) { cfg.ctx = v }
}
