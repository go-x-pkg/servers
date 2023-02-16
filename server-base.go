package servers

import (
	"crypto/tls"
	"fmt"
	"io"
	"time"

	"github.com/go-x-pkg/dumpctx"
)

type WithKind struct {
	Knd Kind `json:"kind" yaml:"kind" bson:"kind"`
}

func (wk *WithKind) Kind() Kind { return wk.Knd }
func (wk *WithKind) validate() error {
	if wk.Knd.Has(KindUNIX) && wk.Knd.Has(KindINET) {
		return ErrGotBothInetAndUnix
	}

	return nil
}

func (wk *WithKind) defaultize() error {
	if !wk.Knd.Has(KindINET) && !wk.Knd.Has(KindUNIX) {
		wk.Knd.Set(KindINET)
	}

	if !wk.Knd.Has(KindHTTP) && !wk.Knd.Has(KindGRPC) {
		wk.Knd.Set(KindHTTP)
	}

	return nil
}

type WithNetwork struct {
	Net string `json:"network" yaml:"network" bson:"network"`
}

type ClientAuthTLSConfig struct {
	// Enable/Disable client auth through mTLS
	Enable bool `json:"enable" yaml:"enable" bson:"enable"`
	// AuthType declares the policy the server will follow for
	// TLS Client Authentication.
	//
	// "NoClientCert" indicates that no client certificate should be requested
	// during the handshake, and if any certificates are sent they will not
	// be verified.
	//
	// "RequestClientCert" indicates that a client certificate should be requested
	// during the handshake, but does not require that the client send any
	// certificates.
	//
	// "RequireAnyClientCert" indicates that a client certificate should be requested
	// during the handshake, and that at least one certificate is required to be
	// sent by the client, but that certificate is not required to be valid.
	//
	// "VerifyClientCertIfGiven" indicates that a client certificate should be requested
	// during the handshake, but does not require that the client sends a
	// certificate. If the client does send a certificate it is required to be
	// valid.
	//
	// "RequireAndVerifyClientCert" indicates that a client certificate should be requested
	// during the handshake, and that at least one valid certificate is required
	// to be sent by the client.
	//
	// If ClientAuthTLS is set true, AuthType must be set.
	AuthType string `json:"authType" yaml:"authType" bson:"authType"`
	// CARoot certificate for clients certificates. Optional.
	TrustedCA string `json:"trustedCA" yaml:"trustedCA" bson:"trustedCA"`
	// If set, server will verifie Common Name of certificate given by client has in this list.
	// Otherwise server return Unauthtorized responce.
	ClientCommonNames []string `json:"clientCommonNames" yaml:"clientCommonNames" bson:"clientCommonNames"`
}

func (c *ClientAuthTLSConfig) getAuthType() tls.ClientAuthType {
	switch c.AuthType {
	case "NoClientCert":
		return tls.NoClientCert
	case "RequestClientCert":
		return tls.RequestClientCert
	case "RequireAnyClientCert":
		return tls.RequireAnyClientCert
	case "VerifyClientCertIfGiven":
		return tls.VerifyClientCertIfGiven
	case "RequireAndVerifyClientCert":
		return tls.RequireAndVerifyClientCert
	default:
		return tls.NoClientCert
	}
}

func (c *ClientAuthTLSConfig) dump(ctx *dumpctx.Ctx, w io.Writer) {
	ctx.Wrap(func() {
		fmt.Fprintf(w, "%sclientAuthTLS:\n", ctx.Indent())
		ctx.Wrap(func() {
			fmt.Fprintf(w, "%snable: %t\n", ctx.Indent(), c.Enable)
			fmt.Fprintf(w, "%sauthType: %s\n", ctx.Indent(), c.AuthType)
			fmt.Fprintf(w, "%strustedCA: %s\n", ctx.Indent(), c.TrustedCA)
			fmt.Fprintf(w, "%sclientCommonNames: %s\n", ctx.Indent(), c.ClientCommonNames)
		})
	})
}

type ServerBase struct {
	WithKind    `json:",inline" yaml:",inline" bson:",inline"`
	WithNetwork `json:",inline" yaml:",inline" bson:",inline"`

	GRPC struct {
		Reflection    bool                `yaml:"reflection"`
		ClientAuthTLS ClientAuthTLSConfig `json:"clientAuthTLS" yaml:"clientAuthTLS" bson:"clientAuthTLS"`
	} `yaml:"grpc"`

	HTTP struct {
		ReadHeaderTimeout time.Duration       `yaml:"readHeaderTimeout"`
		ClientAuthTLS     ClientAuthTLSConfig `json:"clientAuthTLS" yaml:"clientAuthTLS" bson:"clientAuthTLS"`
	} `yaml:"http"`

	Pprof struct {
		Enable bool   `yaml:"enable"`
		Prefix string `yaml:"prefix"`
	} `yaml:"pprof"`
}

func (s *ServerBase) Network() string {
	if s.Net != "" {
		return s.Net
	}

	if s.Kind().Has(KindUNIX) {
		return "unix"
	}

	return "tcp"
}

func (s *ServerBase) getClientAuthConfig() *ClientAuthTLSConfig {
	if s.Knd.Has(KindGRPC) {
		return &s.GRPC.ClientAuthTLS
	}
	return &s.HTTP.ClientAuthTLS
}

func (s *ServerBase) validate() error {
	if s.getClientAuthConfig().getAuthType().String() !=
		s.getClientAuthConfig().AuthType {
		return ErrClientAuthTLSAuthType
	}
	return s.WithKind.validate()
}

func (s *ServerBase) defaultize() error {
	if err := s.WithKind.defaultize(); err != nil {
		return err
	}

	if s.Pprof.Prefix == "" {
		s.Pprof.Prefix = defaultPprofPrefix
	}

	if s.HTTP.ReadHeaderTimeout == 0 {
		s.HTTP.ReadHeaderTimeout = defaultReadHeaderTimeout
	}

	return nil
}

func (s *ServerBase) Dump(ctx *dumpctx.Ctx, w io.Writer) {
	if s.Kind().Has(KindGRPC) {
		fmt.Fprintf(w, "%sgrpc:\n", ctx.Indent())
		ctx.Wrap(func() {
			fmt.Fprintf(w, "%sreflection: %t\n", ctx.Indent(), s.GRPC.Reflection)
			s.getClientAuthConfig().dump(ctx, w)
		})
	}

	if s.Kind().Has(KindHTTP) {
		fmt.Fprintf(w, "%shttp:\n", ctx.Indent())
		ctx.Wrap(func() {
			fmt.Fprintf(w, "%sreadHeaderTimeout: %s\n", ctx.Indent(), s.HTTP.ReadHeaderTimeout)
			s.getClientAuthConfig().dump(ctx, w)
		})
	}

	fmt.Fprintf(w, "%spprof:\n", ctx.Indent())
	ctx.Wrap(func() {
		fmt.Fprintf(w, "%senable: %t\n", ctx.Indent(), s.Pprof.Enable)
		fmt.Fprintf(w, "%sprefix: %q\n", ctx.Indent(), s.Pprof.Prefix)
	})
}
