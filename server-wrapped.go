package servers

import (
	"encoding/json"
	"fmt"

	"github.com/go-x-pkg/isnil"
)

// Struct wrapped typed server.
// Server + kind to unmarshal and build typed json/yaml/bson.
// see https://stackoverflow.com/a/35584188/723095.
type ServerWrapped struct {
	Server `json:",inline" yaml:",inline" bson:",inline"`
}

func (sw *ServerWrapped) unmarshal(fn func(interface{}) error) error {
	wk := WithKind{}

	if err := fn(&wk); err != nil {
		return fmt.Errorf("error unmarshal server-head-kind: %w", err)
	}

	server := wk.Kind().NewServer()
	if server == nil {
		return fmt.Errorf("server-kind(%q): %w", wk.Kind(), ErrUnmarshalUnknownKind)
	}

	sw.Server = server

	if err := fn(sw.Server); err != nil {
		return fmt.Errorf("error unmarshal server: %w", err)
	}

	return nil
}

func (sw *ServerWrapped) MarshalJSON() ([]byte, error) {
	return json.Marshal(sw.Server)
}

func (sw *ServerWrapped) MarshalYAML() (interface{}, error) {
	return sw.Server, nil
}

func (sw *ServerWrapped) UnmarshalJSON(data []byte) error {
	return sw.unmarshal(func(sw interface{}) error { return json.Unmarshal(data, sw) })
}

func (sw *ServerWrapped) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return sw.unmarshal(unmarshal)
}

func serverEnsureWrapped(s Server) *ServerWrapped {
	if isnil.IsNil(s) {
		return nil
	}

	sw, ok := s.(*ServerWrapped)
	if !ok {
		sw = &ServerWrapped{Server: s}
	}

	return sw
}
