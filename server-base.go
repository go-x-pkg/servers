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

type ClientAuthCfgTLS struct {
	// Enable/Disable client auth throw mTLS
	ClientAuthTLS bool `json:"clientAuthTLS" yaml:"clientAuthTLS" bson:"clientAuthTLS"`
	// ClientAuthType declares the policy the server will follow for
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
	// If ClientAuthTLS is set true, ClientAuthType must be set.
	ClientAuthType  string `json:"clientAuthType" yaml:"clientAuthType" bson:"clientAuthType"`
	ClientTrustedCA string `json:"clientTrustedCA" yaml:"clientTrustedCA" bson:"clientTrustedCA"`
	// If set, server will verifie Common Name of certificate given by client has in this list.
	// Otherwise server return Unauthtorized responce.
	ClientsCommonNames []string `json:"clientsCommonNames" yaml:"clientsCommonNames" bson:"clientsCommonNames"`
}

func (c *ClientAuthCfgTLS) getClientAuthType() tls.ClientAuthType {
	switch c.ClientAuthType {
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

type ServerBase struct {
	WithKind    `json:",inline" yaml:",inline" bson:",inline"`
	WithNetwork `json:",inline" yaml:",inline" bson:",inline"`

	GRPC struct {
		Reflection       bool `yaml:"reflection"`
		ClientAuthCfgTLS `json:",inline" yaml:",inline" bson:",inline"`
	} `yaml:"grpc"`

	HTTP struct {
		ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout"`
		ClientAuthCfgTLS  `json:",inline" yaml:",inline" bson:",inline"`
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

func (s *ServerBase) getClientAuthConfig() *ClientAuthCfgTLS {
	if s.Knd.Has(KindGRPC) {
		return &s.GRPC.ClientAuthCfgTLS
	}
	return &s.HTTP.ClientAuthCfgTLS
}

func (s *ServerBase) validate() error {
	if s.getClientAuthConfig().getClientAuthType().String() !=
		s.getClientAuthConfig().ClientAuthType {
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
			fmt.Fprintf(w, "%sclientAuthTLS: %t\n", ctx.Indent(), s.GRPC.ClientAuthTLS)
			fmt.Fprintf(w, "%sclientAuthType: %s\n", ctx.Indent(), s.GRPC.ClientAuthType)
			fmt.Fprintf(w, "%sclientTrustedCA: %s\n", ctx.Indent(), s.GRPC.ClientTrustedCA)
			fmt.Fprintf(w, "%sclientsCommonNames: %s\n", ctx.Indent(), s.GRPC.ClientsCommonNames)
		})
	}

	if s.Kind().Has(KindHTTP) {
		fmt.Fprintf(w, "%shttp:\n", ctx.Indent())
		ctx.Wrap(func() {
			fmt.Fprintf(w, "%sreadHeaderTimeout: %s\n", ctx.Indent(), s.HTTP.ReadHeaderTimeout)
			fmt.Fprintf(w, "%sclientAuthTLS: %t\n", ctx.Indent(), s.HTTP.ClientAuthTLS)
			fmt.Fprintf(w, "%sclientAuthType: %s\n", ctx.Indent(), s.HTTP.ClientAuthType)
			fmt.Fprintf(w, "%sclientTrustedCA: %s\n", ctx.Indent(), s.HTTP.ClientTrustedCA)
			fmt.Fprintf(w, "%sclientsCommonNames: %s\n", ctx.Indent(), s.HTTP.ClientsCommonNames)
		})
	}

	fmt.Fprintf(w, "%spprof:\n", ctx.Indent())
	ctx.Wrap(func() {
		fmt.Fprintf(w, "%senable: %t\n", ctx.Indent(), s.Pprof.Enable)
		fmt.Fprintf(w, "%sprefix: %q\n", ctx.Indent(), s.Pprof.Prefix)
	})
}
