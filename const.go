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

	// Potential slowloris attack GO-S2112.
	defaultReadHeaderTimeout = 3 * time.Second

	defaultTLSPreferServerCipherSuites = true

	defaultVersionTLS = versionTLS13

	defaultClientAuthTypeTLS = clientAuthTypeTLSNoClientCert
)

var (
	defaultFnShutdownTimeout = func() time.Duration { return 5 * time.Second }
	defaultFnLog             = log.LogfStd
)
