package servers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
)

const (
	strNoClientCert               = "NoClientCert"
	strRequestClientCert          = "RequestClientCert"
	strRequireAnyClientCert       = "RequireAnyClientCert"
	strVerifyClientCertIfGiven    = "VerifyClientCertIfGiven"
	strRequireAndVerifyClientCert = "RequireAndVerifyClientCert"
	strClientAuthTypeTLSDefault   = ""
)

type clientAuthTypeTLS tls.ClientAuthType

const clientAuthTypeTLSUnknown clientAuthTypeTLS = 65535

func (c clientAuthTypeTLS) String() string {
	switch tls.ClientAuthType(c) {
	case tls.NoClientCert:
		return strNoClientCert
	case tls.RequestClientCert:
		return strRequestClientCert
	case tls.RequireAnyClientCert:
		return strRequireAnyClientCert
	case tls.VerifyClientCertIfGiven:
		return strVerifyClientCertIfGiven
	case tls.RequireAndVerifyClientCert:
		return strRequireAndVerifyClientCert
	default:
		return "clientAuthTypeTLSUnknown"
	}
}

func (c *clientAuthTypeTLS) unmarshal(fn func(interface{}) error) error {
	var raw string

	if err := fn(&raw); err != nil {
		return fmt.Errorf("error unmarshal client authType: %w", err)
	}
	*c = newClientAuthTypeTLS(raw)

	if *c == clientAuthTypeTLSUnknown {
		return fmt.Errorf("unknown client authType: %q", raw)
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
	case strNoClientCert:
		return clientAuthTypeTLS(tls.NoClientCert)
	case strRequestClientCert:
		return clientAuthTypeTLS(tls.RequestClientCert)
	case strRequireAnyClientCert:
		return clientAuthTypeTLS(tls.RequireAnyClientCert)
	case strVerifyClientCertIfGiven:
		return clientAuthTypeTLS(tls.VerifyClientCertIfGiven)
	case strRequireAndVerifyClientCert:
		return clientAuthTypeTLS(tls.RequireAndVerifyClientCert)
	case strClientAuthTypeTLSDefault:
		return clientAuthTypeTLS(tls.NoClientCert)
	default:
		return clientAuthTypeTLSUnknown
	}
}
