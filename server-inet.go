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
		Enable                   bool   `yaml:"enable"`
		CertFile                 string `yaml:"certFile"`
		KeyFile                  string `yaml:"keyFile"`
		MinVersion               string `yaml:"minVersion"`
		PreferServerCipherSuites bool   `yaml:"preferServerCipherSuites"`
	} `yaml:"tls"`
}

func (s *ServerINET) Base() *ServerBase { return &s.ServerBase }

func (s *ServerINET) setPort(v int) { s.Port = v }

func (s *ServerINET) Addr() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}

func (s *ServerINET) getMinVersionTLS() uint16 {
	switch s.TLS.MinVersion {
	case "VersionTLS10":
		return tls.VersionTLS10
	case "VersionTLS11":
		return tls.VersionTLS11
	case "VersionTLS12":
		return tls.VersionTLS12
	default:
		return tls.VersionTLS13
	}
}

func (s *ServerINET) newConfigTLS() (*tls.Config, error) {
	if !s.TLS.Enable && !s.getClientAuthConfig().ClientAuthTLS {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion: s.getMinVersionTLS(),
		MaxVersion: 0,
		//nolint:gosec
		PreferServerCipherSuites: s.TLS.PreferServerCipherSuites,
	}

	if s.TLS.Enable {
		cert, err := tls.LoadX509KeyPair(s.TLS.CertFile, s.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("error load x509 key pair (:cert %q :key %q): %w",
				s.TLS.CertFile, s.TLS.KeyFile, err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if s.getClientAuthConfig().ClientAuthTLS {
		tlsConfig.ClientAuth = s.getClientAuthConfig().getClientAuthType()

		if s.getClientAuthConfig().ClientTrustedCA != "" {
			if caCertPool, err := loadCACertPool(
				s.getClientAuthConfig().ClientTrustedCA); err != nil {
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
	s.getClientAuthConfig().ClientTrustedCA = interpolateFn(
		s.getClientAuthConfig().ClientTrustedCA)
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

		if s.getMinVersionTLS() == tls.VersionTLS13 &&
			s.TLS.MinVersion != "VersionTLS13" && s.TLS.MinVersion != "" {
			return fmt.Errorf(
				"error (:unexpected tls minVersion %q), expected %q",
				s.TLS.MinVersion, "VersionTLS1(0|1|2|3)")
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
			fmt.Fprintf(w, "%sPreferServerCipherSuites: %t\n", ctx.Indent(), s.TLS.PreferServerCipherSuites)
		})
	}

	s.ServerBase.Dump(ctx, w)
}

func loadCACertPool(caCertPath string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("error read ClientTrustedCA (:cert %q): %w",
			caCertPath, err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("error load ClientTrustedCA (:cert %q)",
			caCertPath)
	}
	return caCertPool, nil
}
