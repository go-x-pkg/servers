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
)

type clientAuthTypeTLS tls.ClientAuthType

const clientAuthTypeTLSDefault clientAuthTypeTLS = clientAuthTypeTLS(tls.NoClientCert)
const clientAuthTypeTLSUnknown clientAuthTypeTLS = -1

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
	case tls.ClientAuthType(clientAuthTypeTLSUnknown):
		return "unknown"
	default:
		return "undefined"
	}
}

func (c clientAuthTypeTLS) SetedOrDefault() string {
	if c.isDefined() {
		return c.String()
	}
	return versionTLSDefault.String()
}

func (c clientAuthTypeTLS) isDefined() bool {
	switch tls.ClientAuthType(c) {
	case tls.NoClientCert:
		return true
	case tls.RequestClientCert:
		return true
	case tls.RequireAnyClientCert:
		return true
	case tls.VerifyClientCertIfGiven:
		return true
	case tls.RequireAndVerifyClientCert:
		return true
	default:
		return false
	}
}

func (c clientAuthTypeTLS) setedOrDefault() clientAuthTypeTLS {
	if c.isDefined() {
		return c
	}
	return clientAuthTypeTLSDefault
}

func (c *clientAuthTypeTLS) unmarshal(fn func(interface{}) error) error {
	var raw string

	if err := fn(&raw); err != nil {
		return fmt.Errorf("error unmarshal client authType: %w", err)
	}

	if *c = newClientAuthTypeTLS(raw); *c == clientAuthTypeTLSUnknown {
		return fmt.Errorf("error unmarshal client authType: %s", clientAuthTypeTLSUnknown)
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
	default:
		return clientAuthTypeTLSUnknown
	}
}
