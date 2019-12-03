package servers

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/go-x-pkg/dumpctx"
)

type ConfigINET struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  struct {
		Enable   bool   `yaml:"enable"`
		CertFile string `yaml:"cert-file"`
		KeyFile  string `yaml:"key-file"`
	} `yaml:"tls"`
}

func (cfg *ConfigINET) SetPort(v int)   { cfg.Port = v }
func (cfg *ConfigINET) GetAddr() string { return cfg.Addr() }

func (cfg *ConfigINET) Addr() string {
	return net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
}

func (cfg *ConfigINET) Dump(ctx *dumpctx.Ctx, w io.Writer) {
	fmt.Fprintf(w, "%shost: %s\n", ctx.Indent(), cfg.Host)
	fmt.Fprintf(w, "%sport: %d\n", ctx.Indent(), cfg.Port)
	if cfg.TLS.Enable {
		fmt.Fprintf(w, "%stls:\n", ctx.Indent())

		ctx.Enter()

		fmt.Fprintf(w, "%senable: %t\n", ctx.Indent(), cfg.TLS.Enable)
		fmt.Fprintf(w, "%scert-file: %s\n", ctx.Indent(), cfg.TLS.CertFile)
		fmt.Fprintf(w, "%skey-file: %s\n", ctx.Indent(), cfg.TLS.KeyFile)

		ctx.Leave()
	}
}

func (cfg *ConfigINET) Defaultize() error {
	return nil
}
