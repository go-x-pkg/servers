package servers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	versionTLS10Str = "tls-1.0"
	versionTLS11Str = "tls-1.1"
	versionTLS12Str = "tls-1.2"
	versionTLS13Str = "tls-1.3"
)

type versionTLS uint8

const (
	versionTLSUnknown versionTLS = iota
	versionTLS10
	versionTLS11
	versionTLS12
	versionTLS13
)

func (v versionTLS) String() string {
	switch v {
	case versionTLSUnknown:
		return "unknown"
	case versionTLS10:
		return versionTLS10Str
	case versionTLS11:
		return versionTLS11Str
	case versionTLS12:
		return versionTLS12Str
	case versionTLS13:
		return versionTLS13Str
	default:
		panic("undefined versionTLS")
	}
}

// CryptoTLSVersion return native tls.Version* from crypto/tls package.
func (v versionTLS) CryptoTLSVersion() uint16 {
	switch v {
	case versionTLSUnknown:
		return defaultVersionTLS.CryptoTLSVersion()
	case versionTLS10:
		return tls.VersionTLS10
	case versionTLS11:
		return tls.VersionTLS11
	case versionTLS12:
		return tls.VersionTLS12
	case versionTLS13:
		return tls.VersionTLS13
	default:
		return defaultVersionTLS.CryptoTLSVersion()
	}
}

func (v versionTLS) orDefault() versionTLS {
	if v == versionTLSUnknown {
		return defaultVersionTLS
	}

	return v
}

func (v *versionTLS) unmarshal(fn func(interface{}) error) error {
	var raw string

	if err := fn(&raw); err != nil {
		return fmt.Errorf("error unmarshal tls version: %w", err)
	}

	*v = newVersionTLS(raw)
	if *v == versionTLSUnknown {
		return fmt.Errorf("%s: %w", raw, ErrUnknownVersionTLS)
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
	switch strings.ToLower(raw) {
	case versionTLS10Str, "tls 1.0", "versiontls10":
		return versionTLS10
	case versionTLS11Str, "tls 1.1", "versiontls11":
		return versionTLS11
	case versionTLS12Str, "tls 1.2", "versiontls12":
		return versionTLS12
	case versionTLS13Str, "tls 1.3", "versiontls13":
		return versionTLS13
	default:
		return versionTLSUnknown
	}
}
