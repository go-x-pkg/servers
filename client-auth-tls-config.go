package servers

import (
	"fmt"
	"io"

	"github.com/go-x-pkg/dumpctx"
)

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
	AuthType clientAuthTypeTLS `json:"authType" yaml:"authType" bson:"authType"`
	// CARoot certificate for clients certificates. Optional.
	TrustedCA string `json:"caCert" yaml:"trustedCa" bson:"caCert"`
	// If set, server will verifie Common Name of certificate given by client has in this list.
	// Otherwise server return Unauthtorized response.
	ClientCommonNames []string `json:"clientCommonNames" yaml:"clientCommonNames" bson:"clientCommonNames"`
}

func (c *ClientAuthTLSConfig) defaultize() {
	if c.AuthType == clientAuthTypeTLSUnknown {
		c.AuthType = defaultClientAuthTypeTLS
	}
}

func (c *ClientAuthTLSConfig) dump(ctx *dumpctx.Ctx, w io.Writer) {
	fmt.Fprintf(w, "%stls:\n", ctx.Indent())
	ctx.Wrap(func() {
		fmt.Fprintf(w, "%senable: %t\n", ctx.Indent(), c.Enable)
		fmt.Fprintf(w, "%sauthType: %s\n", ctx.Indent(), c.AuthType.orDefault())
		fmt.Fprintf(w, "%strustedCA: %q\n", ctx.Indent(), c.TrustedCA)
		fmt.Fprintf(w, "%sclientCommonNames: %s\n", ctx.Indent(), c.ClientCommonNames)
	})
}
