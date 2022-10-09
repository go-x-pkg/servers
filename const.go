package servers

import (
	"os"
	"time"

	"github.com/go-x-pkg/log"
)

const (
	defaultUNIXSocketFileMode os.FileMode = 0o666

	// defaultPprofPrefix url prefix of pprof.
	defaultPprofPrefix = "/debug/pprof"
)

var (
	defaultFnShutdownTimeout = func() time.Duration { return 5 * time.Second }
	defaultFnLog             = log.LogfStd
)
