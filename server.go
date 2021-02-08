package servers

import (
	"io"

	"github.com/go-x-pkg/dumpctx"
)

type Server interface {
	serverKinder
	serverAddred
}

type (
	serverKinder interface{ Kind() Kind }
	serverAddred interface{ Addr() string }

	serverValidator interface{ validate() error }

	serverDefaultizer interface{ defaultize() error }

	serverDumper interface{ Dump(*dumpctx.Ctx, io.Writer) }
)

func runLogPrefix(s Server) string {
	if s.Kind().IsEmpty() {
		return "!!!"
	} else if s.Kind().Has(KindUNIX) {
		return "UNIX"
	} else if s.Kind().Has(KindINET) {
		if inet, ok := s.(*ServerINET); ok && inet.TLS.Enable {
			return "TLS"
		}

		return "TCP"
	}

	return "???"
}
