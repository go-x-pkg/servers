package servers

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/go-x-pkg/dumpctx"
	"github.com/go-x-pkg/fnspath"
)

type ConfigWithPort interface {
	SetPort(v int)
}

func configSetPort(any interface{}, port int) bool {
	if cwp, ok := any.(ConfigWithPort); ok {
		cwp.SetPort(port)
		return true
	}
	return false
}

type Configer interface {
	Dump(ctx *dumpctx.Ctx, w io.Writer)
	Defaultize() error
	GetAddr() string
}

type configUnmarshal struct {
	Kind Kind `yaml:"kind"`
}

type Config struct {
	kind Kind

	inner Configer
}

func (cfg *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	that := configUnmarshal{}
	if e := unmarshal(&that); e != nil {
		return e
	}

	switch that.Kind {
	case KindINET, KindUnknown:
		cfg.kind = KindINET
		cfg.inner = &ConfigINET{}
	case KindUNIX:
		cfg.kind = KindUNIX
		cfg.inner = &ConfigUNIX{}
	}

	if e := unmarshal(cfg.inner); e != nil {
		return e
	}

	return nil
}

func (cfg *Config) Validate() error {
	if cfg.kind == KindUNIX {
		that := cfg.inner.(*ConfigUNIX)

		if v := that.Addr; v != "" {
			dir := filepath.Dir(v)
			if exists, err := fnspath.IsExists(dir); err != nil {
				return fmt.Errorf("unix socket parent dir '%s' existence check failed: %w", dir, err)
			} else if !exists {
				return fmt.Errorf("unix socket parent dir '%s' does not exist", dir)
			}
		} else {
			return fmt.Errorf("unix socket path is empty")
		}
	} else if cfg.kind == KindINET {
		that := cfg.inner.(*ConfigINET)

		if that.TLS.Enable {
			if v := that.TLS.CertFile; v != "" {
				if exists, err := fnspath.IsExists(v); err != nil {
					return fmt.Errorf("tls cert-file existence check failed: %w", err)
				} else if !exists {
					return fmt.Errorf("tls cert-file '%s' does not exist", v)
				}
			} else {
				return fmt.Errorf("tls cert-file path is not provided")
			}

			if v := that.TLS.KeyFile; v != "" {
				if exists, err := fnspath.IsExists(v); err != nil {
					return fmt.Errorf("tls key-file existence check failed: %w", err)
				} else if !exists {
					return fmt.Errorf("tls key-file '%s' does not exist", v)
				}
			} else {
				return fmt.Errorf("tls key-file path is not provided")
			}
		}
	}

	return nil
}

func (cfg *Config) RunLogPrefix() string {
	switch cfg.kind {
	case KindUnknown:
		return "!!!"
	case KindINET:
		that := cfg.inner.(*ConfigINET)

		if that.TLS.Enable {
			return "HTTPS"
		}

		return "HTTP"
	case KindUNIX:
		return "UNIX"
	}

	return "???"
}

func (cfg *Config) Dump(ctx *dumpctx.Ctx, w io.Writer) {
	ctx.EmitPrefix(w)

	fmt.Fprintf(w, "kind: %s\n", cfg.kind)

	ctx.Enter()
	defer ctx.Leave()

	cfg.inner.Dump(ctx, w)
}
