package servers

import (
	"context"
	"time"

	"github.com/go-x-pkg/log"
)

type logGRPC struct {
	logger   log.FnT
	levelMin log.Level
}
type args struct {
	fnShutdownTimeout func() time.Duration

	fnLog log.FnT

	fnLogHTTPError log.FnT

	logGRPC

	ctx context.Context
}

func (cfg *args) defaultize() {
	cfg.fnShutdownTimeout = defaultFnShutdownTimeout
	cfg.fnLog = defaultFnLog
	cfg.fnLogHTTPError = defaultFnLog
	cfg.logGRPC.logger = defaultFnLog
	cfg.logGRPC.levelMin = log.Quiet
	cfg.ctx = context.TODO()
}

type Arg func(*args)

func FnShutdownTimeout(v func() time.Duration) Arg {
	return func(cfg *args) { cfg.fnShutdownTimeout = v }
}

func FnLog(v log.FnT) Arg {
	return func(cfg *args) { cfg.fnLog = v }
}

func FnLogHTTPError(v log.FnT) Arg {
	return func(cfg *args) { cfg.fnLogHTTPError = v }
}

func LogGRPC(logger log.FnT, levelMin log.Level) Arg {
	return func(cfg *args) {
		cfg.logGRPC.logger = logger
		cfg.logGRPC.levelMin = levelMin
	}
}

func Context(v context.Context) Arg {
	return func(cfg *args) { cfg.ctx = v }
}
