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

	serverValidator interface{ validate() error }

	serverDefaultizer interface{ defaultize() error }

	serverDumper interface{ Dump(*dumpctx.Ctx, io.Writer) }
)

func runLogPrefix(server Server) string {
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
