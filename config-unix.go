package servers

import (
	"fmt"
	"io"
	"os"

	"github.com/go-x-pkg/dumpctx"
)

type ConfigUNIX struct {
	Addr           string      `yaml:"addr"`
	SocketFileMode os.FileMode `yaml:"socket-file-mode"`
}

func (cfg *ConfigUNIX) GetAddr() string { return cfg.Addr }

func (cfg *ConfigUNIX) Dump(ctx *dumpctx.Ctx, w io.Writer) {
	fmt.Fprintf(w, "%saddr: %s\n", ctx.Indent(), cfg.Addr)
	fmt.Fprintf(w, "%ssocket-file-mode: %03o | %s\n", ctx.Indent(), cfg.SocketFileMode, cfg.SocketFileMode)
}

func (cfg *ConfigUNIX) Defaultize() error {
	if cfg.SocketFileMode == 0 {
		cfg.SocketFileMode = DefaultUNIXSocketFileMode | os.ModeSocket
	}

	return nil
}
