package servers

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/go-x-pkg/dumpctx"
	"github.com/go-x-pkg/fnspath"
)

type ServerINET struct {
	ServerBase `json:",inline" yaml:",inline" bson:",inline"`

	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	TLS struct {
		Enable   bool   `yaml:"enable"`
		CertFile string `yaml:"cert-file"`
		KeyFile  string `yaml:"key-file"`
	} `yaml:"tls"`

	GRPC struct {
		Reflection bool `yaml:"reflection"`
	} `yaml:"grpc"`
}

func (s *ServerINET) setPort(v int) { s.Port = v }

func (s *ServerINET) Addr() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}

func (s *ServerINET) validate() error {
	if err := s.ServerBase.validate(); err != nil {
		return err
	}

	if s.TLS.Enable {
		if v := s.TLS.CertFile; v != "" {
			if exists, err := fnspath.IsExists(v); err != nil {
				return fmt.Errorf("tls cert-file existence check failed: %w", err)
			} else if !exists {
				return fmt.Errorf("error (:path %q): %w", v, ErrTLSCertFileNotExists)
			}
		} else {
			return ErrTLSCertFilePathNotProvided
		}

		if v := s.TLS.KeyFile; v != "" {
			if exists, err := fnspath.IsExists(v); err != nil {
				return fmt.Errorf("tls key-file existence check failed: %w", err)
			} else if !exists {
				return fmt.Errorf("error (:path %q): %w", v, ErrTLSKeyFileNotExists)
			}
		} else {
			return ErrTLSKeyFilePathNotProvided
		}
	}

	return nil
}

func (s *ServerINET) Dump(ctx *dumpctx.Ctx, w io.Writer) {
	fmt.Fprintf(w, "%shost: %s\n", ctx.Indent(), s.Host)
	fmt.Fprintf(w, "%sport: %d\n", ctx.Indent(), s.Port)
	if s.TLS.Enable {
		fmt.Fprintf(w, "%stls:\n", ctx.Indent())

		ctx.Wrap(func() {
			fmt.Fprintf(w, "%senable: %t\n", ctx.Indent(), s.TLS.Enable)
			fmt.Fprintf(w, "%scert-file: %s\n", ctx.Indent(), s.TLS.CertFile)
			fmt.Fprintf(w, "%skey-file: %s\n", ctx.Indent(), s.TLS.KeyFile)
		})
	}

	fmt.Fprintf(w, "%sgrpc:\n", ctx.Indent())

	ctx.Wrap(func() {
		fmt.Fprintf(w, "%sreflection: %t\n", ctx.Indent(), s.GRPC.Reflection)
	})
}
