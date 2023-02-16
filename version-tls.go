package servers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
)

const (
	strVersionTLS10      = "VersionTLS10"
	strVersionTLS11      = "VersionTLS11"
	strVersionTLS12      = "VersionTLS12"
	strVersionTLS13      = "VersionTLS13"
	strVersionTLSDefault = ""
)

type versionTLS uint16

const versionTLSDefault versionTLS = 0
const versionTLSUnknown versionTLS = 65535

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
	case versionTLSDefault:
		return strVersionTLS13
	default:
		return "versionTLSUnknown"
	}
}

func (v *versionTLS) unmarshal(fn func(interface{}) error) error {
	var raw string

	if err := fn(&raw); err != nil {
		return fmt.Errorf("error unmarshal tls version: %w", err)
	}
	*v = newVersionTLS(raw)

	if *v == versionTLSUnknown {
		return fmt.Errorf("unknown tls version: %q", raw)
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
	case strVersionTLSDefault:
		return tls.VersionTLS13
	default:
		return versionTLSUnknown
	}
}
