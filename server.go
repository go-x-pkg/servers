package servers

import (
	"io"

	"github.com/go-x-pkg/dumpctx"
)

type Server interface {
	serverBaser
	serverKinder
	serverAddred
	serverNetworker
}

type (
	serverBaser     interface{ Base() *ServerBase }
	serverKinder    interface{ Kind() Kind }
	serverAddred    interface{ Addr() string }
	serverNetworker interface{ Network() string }
	ServerAuther    interface{ ClientAuthTLS() *ClientAuthTLSConfig }

	serverValidator interface{ validate() error }

	serverInterpolator interface{ interpolate(func(string) string) }

	serverDefaultizer interface{ defaultize() error }

	serverDumper interface{ Dump(*dumpctx.Ctx, io.Writer) }
)

func runLogPrefix(server Server) string {
	switch s := server.(type) {
	case *ServerListener:
		server = s.Server
	case *ServerWrapped:
		server = s.Server
	}

	switch s := server.(type) {
	case *ServerUNIX:
		return "UNIX"
	case *ServerINET:
		if s.TLS.Enable {
			return "TLS"
		}

		return "TCP"
	default:
		panic("undefined server")
	}
}
