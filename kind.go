package servers

import (
	"fmt"
	"strings"
)

type Kind uint8

const (
	KindUnknown Kind = iota
	KindINET
	KindUNIX
)

var kindText = map[Kind]string{
	KindUnknown: "unknown",
	KindINET:    "inet / AF_INET",
	KindUNIX:    "unix / AF_UNIX",
}

func (knd Kind) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", knd.String())), nil
}
func (knd Kind) MarshalYAML() (interface{}, error) { return knd.String(), nil }

func (knd Kind) String() string { return kindText[knd] }

func (knd *Kind) UnmarshalJSON(data []byte) error {
	v := strings.Replace(string(data), "\"", "", -1)
	v = strings.TrimSpace(v)
	*knd = NewKindFromString(v)

	return nil
}

func (knd *Kind) UnmarshalYAML(unmarshal func(interface{}) error) error {
	v := ""
	if err := unmarshal(&v); err != nil {
		return err
	}

	*knd = NewKindFromString(v)

	return nil
}

func NewKindFromString(v string) Kind {
	v = strings.TrimSpace(v)
	v = strings.Replace(v, "/", "", -1)
	v = strings.Replace(v, "\"", "", -1)
	v = strings.Replace(v, "'", "", -1)
	v = strings.Replace(v, "_", "-", -1)
	v = strings.Replace(v, " ", "-", -1)
	v = strings.ToLower(v)

	switch v {
	case "inet":
		return KindINET
	case "unix":
		return KindUNIX
	}

	return KindUnknown
}
