package servers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
)

const (
	strVersionTLS10 = "TLS 1.0"
	strVersionTLS11 = "TLS 1.1"
	strVersionTLS12 = "TLS 1.2"
	strVersionTLS13 = "TLS 1.3"
)

type versionTLS uint16

const versionTLSDefault versionTLS = tls.VersionTLS13

const versionTLSUnknown versionTLS = 0

func (v versionTLS) String() string {
	switch v {
	case tls.VersionTLS10:
		return strVersionTLS10
	case tls.VersionTLS11:
		return strVersionTLS11
	case tls.VersionTLS12:
		return strVersionTLS12
	case tls.VersionTLS13:
		return strVersionTLS13
	case versionTLSUnknown:
		return "unknown"
	default:
		return "undefined"
	}
}

func (v versionTLS) SetedOrDefault() string {
	if v.isDefined() {
		return v.String()
	}
	return versionTLSDefault.String()
}

func (v versionTLS) isDefined() bool {
	switch v {
	case tls.VersionTLS10:
		return true
	case tls.VersionTLS11:
		return true
	case tls.VersionTLS12:
		return true
	case tls.VersionTLS13:
		return true
	default:
		return false
	}
}

func (v versionTLS) setedOrDefault() versionTLS {
	if v.isDefined() {
		return v
	}
	return versionTLSDefault
}

func (v *versionTLS) unmarshal(fn func(interface{}) error) error {
	var raw string

	if err := fn(&raw); err != nil {
		return fmt.Errorf("error unmarshal tls version: %w", err)
	}

	if *v = newVersionTLS(raw); *v == versionTLSUnknown {
		return fmt.Errorf("error unmarshal tls version: %s", versionTLSUnknown)
	}

	return nil
}

func (v versionTLS) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", v.String())), nil
}

func (v versionTLS) MarshalYAML() (interface{}, error) {
	return v.String(), nil
}

func (v *versionTLS) UnmarshalJSON(data []byte) error {
	return v.unmarshal(func(v interface{}) error { return json.Unmarshal(data, v) })
}

func (v *versionTLS) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return v.unmarshal(unmarshal)
}

func newVersionTLS(raw string) versionTLS {
	switch raw {
	case strVersionTLS10:
		return tls.VersionTLS10
	case strVersionTLS11:
		return tls.VersionTLS11
	case strVersionTLS12:
		return tls.VersionTLS12
	case strVersionTLS13:
		return tls.VersionTLS13
	default:
		return versionTLSUnknown
	}
}
