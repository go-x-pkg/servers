package servers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
)

const (
	clientAuthTypeTLSNoClientCertStr               = "NoClientCert"
	clientAuthTypeTLSRequestClientCertStr          = "RequestClientCert"
	clientAuthTypeTLSRequireAnyClientCertStr       = "RequireAnyClientCert"
	clientAuthTypeTLSVerifyClientCertIfGivenStr    = "VerifyClientCertIfGiven"
	clientAuthTypeTLSRequireAndVerifyClientCertStr = "RequireAndVerifyClientCert"
)

type clientAuthTypeTLS uint8

const (
	clientAuthTypeTLSUnknown clientAuthTypeTLS = iota
	clientAuthTypeTLSNoClientCert
	clientAuthTypeTLSRequestClientCert
	clientAuthTypeTLSRequireAnyClientCert
	clientAuthTypeTLSVerifyClientCertIfGiven
	clientAuthTypeTLSRequireAndVerifyClientCert
)

func (c clientAuthTypeTLS) String() string {
	switch c {
	case clientAuthTypeTLSUnknown:
		return "unknown"
	case clientAuthTypeTLSNoClientCert:
		return clientAuthTypeTLSNoClientCertStr
	case clientAuthTypeTLSRequestClientCert:
		return clientAuthTypeTLSRequestClientCertStr
	case clientAuthTypeTLSRequireAnyClientCert:
		return clientAuthTypeTLSRequireAnyClientCertStr
	case clientAuthTypeTLSVerifyClientCertIfGiven:
		return clientAuthTypeTLSVerifyClientCertIfGivenStr
	case clientAuthTypeTLSRequireAndVerifyClientCert:
		return clientAuthTypeTLSRequireAndVerifyClientCertStr
	default:
		panic("undefined clientAuthTypeTLS")
	}
}

// CryptoTLSVersion return native tls.Version* from crypto/tls package.
func (c clientAuthTypeTLS) CryptoTLSClientAuthType() tls.ClientAuthType {
	switch c {
	case clientAuthTypeTLSUnknown:
		return defaultClientAuthTypeTLS.CryptoTLSClientAuthType()
	case clientAuthTypeTLSNoClientCert:
		return tls.NoClientCert
	case clientAuthTypeTLSRequestClientCert:
		return tls.RequestClientCert
	case clientAuthTypeTLSRequireAnyClientCert:
		return tls.RequireAnyClientCert
	case clientAuthTypeTLSVerifyClientCertIfGiven:
		return tls.VerifyClientCertIfGiven
	case clientAuthTypeTLSRequireAndVerifyClientCert:
		return tls.RequireAnyClientCert
	default:
		return defaultClientAuthTypeTLS.CryptoTLSClientAuthType()
	}
}

func (c clientAuthTypeTLS) orDefault() clientAuthTypeTLS {
	if c == clientAuthTypeTLSUnknown {
		return defaultClientAuthTypeTLS
	}

	return c
}

func (c *clientAuthTypeTLS) unmarshal(fn func(interface{}) error) error {
	var raw string

	if err := fn(&raw); err != nil {
		return fmt.Errorf("error unmarshal client authType: %w", err)
	}

	*c = newClientAuthTypeTLS(raw)
	if *c == clientAuthTypeTLSUnknown {
		return fmt.Errorf("%s: %w", raw, ErrUnknownClientAuthTypeTLS)
	}

	return nil
}

func (c clientAuthTypeTLS) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", c.String())), nil
}

func (c clientAuthTypeTLS) MarshalYAML() (interface{}, error) {
	return c.String(), nil
}

func (c *clientAuthTypeTLS) UnmarshalJSON(data []byte) error {
	return c.unmarshal(func(c interface{}) error { return json.Unmarshal(data, c) })
}

func (c *clientAuthTypeTLS) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return c.unmarshal(unmarshal)
}

func newClientAuthTypeTLS(raw string) clientAuthTypeTLS {
	switch raw {
	case "no-client-cert", clientAuthTypeTLSNoClientCertStr:
		return clientAuthTypeTLSNoClientCert
	case "request-client-cert", clientAuthTypeTLSRequestClientCertStr:
		return clientAuthTypeTLSRequestClientCert
	case "require-any-client-cert", clientAuthTypeTLSRequireAnyClientCertStr:
		return clientAuthTypeTLSRequireAnyClientCert
	case "verify-client-cert-if-given", clientAuthTypeTLSVerifyClientCertIfGivenStr:
		return clientAuthTypeTLSVerifyClientCertIfGiven
	case "require-and-verify-client-cert", clientAuthTypeTLSRequireAndVerifyClientCertStr:
		return clientAuthTypeTLSRequireAndVerifyClientCert
	default:
		return clientAuthTypeTLSUnknown
	}
}
