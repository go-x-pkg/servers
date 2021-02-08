package servers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrKindINETAndUNIX = errors.New("got both INET and UNIX kind")

type Kind uint8

const (
	KindEmpty Kind = 0

	KindINET = 1 << (iota - 1)
	KindUNIX

	KindHTTP
	KindGRPC
)

var kindText = map[Kind]string{
	KindEmpty: "unknown",

	KindINET: "inet", // AF_INET
	KindUNIX: "unix", // AF_UNIX

	KindHTTP: "http",
	KindGRPC: "gRPC",
}

// is power of two algorithm
func (knd Kind) IsSingle() bool {
	return !knd.IsEmpty() && (knd&(knd-1) == 0)
}

func (knd Kind) IsEmpty() bool   { return knd == KindEmpty }
func (knd Kind) Has(v Kind) bool { return (knd & v) != 0 }
func (knd *Kind) Set(v Kind)     { *knd |= v }
func (knd *Kind) UnSet(v Kind)   { *knd &= ^v }

func (knd Kind) String() string {
	if knd.IsEmpty() {
		return "~"
	}
	s, _ := knd.MarshalJSON()
	return string(s)
}

func (knd Kind) StringTrySingle() string {
	if knd.IsSingle() {
		return kindText[knd]
	}

	return knd.String()
}

func (knd Kind) validate() error {
	if knd.Has(KindINET) && knd.Has(KindUNIX) {
		return ErrKindINETAndUNIX
	}

	return nil
}

func (knd *Kind) unmarshal(fn func(interface{}) error) error {
	var (
		raw   string
		rawSs []string
	)

	if err1 := fn(&raw); err1 != nil {
		if err2 := fn(&rawSs); err2 != nil {
			return fmt.Errorf("error unmarshal kind (%s): %w", err1, err2)
		}

		*knd = NewKindFromStringSlice(rawSs)
	} else {
		*knd = NewKindFromString(raw)
	}

	return knd.validate()
}

func (knd Kind) MarshalJSON() ([]byte, error) { return json.Marshal(knd.ToStringSlice()) }

func (knd Kind) MarshalYAML() (interface{}, error) { return knd.ToStringSlice(), nil }

func (knd *Kind) UnmarshalJSON(data []byte) error {
	return knd.unmarshal(func(knd interface{}) error { return json.Unmarshal(data, knd) })
}

func (knd *Kind) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return knd.unmarshal(unmarshal)
}

func (knd Kind) ToStringSlice() (vv []string) {
	if knd.Has(KindINET) {
		vv = append(vv, kindText[KindINET])
	}
	if knd.Has(KindUNIX) {
		vv = append(vv, kindText[KindUNIX])
	}
	if knd.Has(KindHTTP) {
		vv = append(vv, kindText[KindHTTP])
	}
	if knd.Has(KindGRPC) {
		vv = append(vv, kindText[KindGRPC])
	}

	return
}

func (knd Kind) NewServer() Server {
	if knd.Has(KindUNIX) {
		return new(ServerUNIX)
	}

	return new(ServerINET)
}

func NewKindFromStringSlice(vv []string) Kind {
	knd := KindEmpty

	for _, v := range vv {
		knd.Set(NewKindFromString(v))
	}

	return knd
}

func NewKindFromString(v string) Kind {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "/", "")
	v = strings.ReplaceAll(v, "\"", "")
	v = strings.ReplaceAll(v, "'", "")
	v = strings.ReplaceAll(v, "_", "-")
	v = strings.ReplaceAll(v, " ", "-")
	v = strings.ToLower(v)

	switch v {
	case "inet":
		return KindINET
	case "unix":
		return KindUNIX
	case "http":
		return KindHTTP
	case "grpc":
		return KindGRPC
	}

	return KindEmpty
}
