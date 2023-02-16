package servers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/go-x-pkg/dumpctx"
	"github.com/go-x-pkg/fnspath"
)

type ServerINET struct {
	ServerBase `json:",inline" yaml:",inline" bson:",inline"`

	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	TLS struct {
		Enable                   bool       `yaml:"enable"`
		CertFile                 string     `yaml:"certFile"`
		KeyFile                  string     `yaml:"keyFile"`
		MinVersion               versionTLS `yaml:"minVersion"`
		MaxVersion               versionTLS `yaml:"maxVersion"`
		PreferServerCipherSuites *bool      `yaml:"preferServerCipherSuites"`
	} `yaml:"tls"`
}

func (s *ServerINET) Base() *ServerBase { return &s.ServerBase }

func (s *ServerINET) setPort(v int) { s.Port = v }

func (s *ServerINET) Addr() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}

func (s *ServerINET) newTLSConfig() (*tls.Config, error) {
	if !s.TLS.Enable && !s.getClientAuthTLS().Enable {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion:               uint16(s.TLS.MinVersion),
		MaxVersion:               uint16(s.TLS.MaxVersion),
		PreferServerCipherSuites: *s.TLS.PreferServerCipherSuites,
	}

	if s.TLS.Enable {
		cert, err := tls.LoadX509KeyPair(s.TLS.CertFile, s.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("error load x509 key pair (:cert %q :key %q): %w",
				s.TLS.CertFile, s.TLS.KeyFile, err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if s.getClientAuthTLS().Enable {
		tlsConfig.ClientAuth = tls.ClientAuthType(s.getClientAuthTLS().AuthType)

		if s.getClientAuthTLS().TrustedCA != "" {
			if caCertPool, err := loadCACertPool(
				s.getClientAuthTLS().TrustedCA); err != nil {
				return nil, err
			} else {
				tlsConfig.ClientCAs = caCertPool
			}
		}
	}
	return tlsConfig, nil
}

func (s *ServerINET) interpolate(interpolateFn func(string) string) {
	if interpolateFn == nil {
		return
	}

	s.TLS.CertFile = interpolateFn(s.TLS.CertFile)
	s.TLS.KeyFile = interpolateFn(s.TLS.KeyFile)
	s.getClientAuthTLS().TrustedCA = interpolateFn(
		s.getClientAuthTLS().TrustedCA)
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

		if s.TLS.PreferServerCipherSuites == nil {
			p := true
			s.TLS.PreferServerCipherSuites = &p
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
			fmt.Fprintf(w, "%scertFile: %s\n", ctx.Indent(), s.TLS.CertFile)
			fmt.Fprintf(w, "%skeyFile: %s\n", ctx.Indent(), s.TLS.KeyFile)
			fmt.Fprintf(w, "%sminVersion: %s\n", ctx.Indent(), s.TLS.MinVersion)
			fmt.Fprintf(w, "%smaxVersion: %s\n", ctx.Indent(), s.TLS.MaxVersion)
			fmt.Fprintf(w, "%sPreferServerCipherSuites: %t\n", ctx.Indent(), *s.TLS.PreferServerCipherSuites)

			if !*s.TLS.PreferServerCipherSuites {
				fmt.Fprintf(w, "%sWARNING: PreferServerCipherSuites is false. %s\n",
					ctx.Indent(), "Set to true for avoid potentinal security risk!")
			}
		})
	}

	s.ServerBase.Dump(ctx, w)
}

func loadCACertPool(caCertPath string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("error read TrustedCA (:cert %q): %w",
			caCertPath, err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("error load TrustedCA (:cert %q)",
			caCertPath)
	}
	return caCertPool, nil
}
