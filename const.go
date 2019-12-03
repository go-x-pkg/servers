package servers

import (
	"os"
	"time"

	"github.com/go-x-pkg/log"
)

const (
	DefaultUNIXSocketFileMode os.FileMode = 0666
)

var (
	defaultFnShutdownTimeout = func() time.Duration { return 5 * time.Second }
	defaultFnLog             = log.LogfStd
)
