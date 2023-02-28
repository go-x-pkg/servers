package servers

import (
	"errors"
	"fmt"

	xlog "github.com/go-x-pkg/log"
	"go.uber.org/zap"
)

const prefix = "xGRPCLoggerV2"

const (
	infoLevelGRPC = iota
	warnLevelGRPC
	errorLevelGRPC
	fatalLevelGRPC
)

type xGRPCLoggerV2 struct {
	logGRPC
}

func (l xGRPCLoggerV2) Info(args ...interface{}) {
	l.logger(xlog.Info, prefix, zap.Any("info", args))
}

func (l xGRPCLoggerV2) Infoln(args ...interface{}) {
	l.logger(xlog.Info, prefix, zap.Any("info", args))
}

func (l xGRPCLoggerV2) Infof(format string, args ...interface{}) {
	fmtstr := fmt.Sprintf(format, args...)
	l.logger(xlog.Info, prefix, zap.Any("info", fmtstr))
}

func (l xGRPCLoggerV2) Warning(args ...interface{}) {
	l.logger(xlog.Warn, prefix, zap.Any("warning", args))
}

func (l xGRPCLoggerV2) Warningln(args ...interface{}) {
	l.logger(xlog.Warn, prefix, zap.Any("warning", args))
}

func (l xGRPCLoggerV2) Warningf(format string, args ...interface{}) {
	fmtstr := fmt.Sprintf(format, args...)
	l.logger(xlog.Warn, prefix, zap.Any("warning", fmtstr))
}

func (l xGRPCLoggerV2) Error(args ...interface{}) {
	err := errors.New(fmt.Sprint(args...))
	l.logger(xlog.Error, prefix, zap.Error(err))
}

func (l xGRPCLoggerV2) Errorln(args ...interface{}) {
	err := errors.New(fmt.Sprint(args...))
	l.logger(xlog.Error, prefix, zap.Error(err))
}

func (l xGRPCLoggerV2) Errorf(format string, args ...interface{}) {
	err := errors.New(fmt.Sprintf(format, args...))
	l.logger(xlog.Error, prefix, zap.Error(err))
}

func (l xGRPCLoggerV2) Fatal(args ...interface{}) {
	l.logger(xlog.Critical, prefix, zap.Any("fatal", args))
}

func (l xGRPCLoggerV2) Fatalln(args ...interface{}) {
	l.logger(xlog.Critical, prefix, zap.Any("fatal", args))
}

func (l xGRPCLoggerV2) Fatalf(format string, args ...interface{}) {
	fmtstr := fmt.Sprintf(format, args...)
	l.logger(xlog.Critical, prefix, zap.Any("fatal", fmtstr))
}

func (l xGRPCLoggerV2) xLogLevel(levelGRPC int) xlog.Level {
	switch levelGRPC {
	case infoLevelGRPC:
		return xlog.Info
	case warnLevelGRPC:
		return xlog.Warn
	case errorLevelGRPC:
		return xlog.Error
	case fatalLevelGRPC:
		return xlog.Critical
	default:
		return l.levelMin
	}
}

func (l xGRPCLoggerV2) V(level int) bool {
	if l.levelMin <= xlog.Debug {
		return true
	}

	return l.levelMin == l.xLogLevel(level)
}

func newXGRPCLogger(l logGRPC) xGRPCLoggerV2 {
	return xGRPCLoggerV2{l}
}
