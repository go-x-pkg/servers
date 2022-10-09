package servers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-x-pkg/dumpctx"
	"github.com/go-x-pkg/fnspath"
)

type ServerUNIX struct {
	ServerBase `json:",inline" yaml:",inline" bson:",inline"`

	Address        string      `yaml:"addr"`
	SocketFileMode os.FileMode `yaml:"socketFileMode"`
}

func (s *ServerUNIX) Base() *ServerBase { return &s.ServerBase }

func (s *ServerUNIX) Addr() string { return s.Address }

func (s *ServerUNIX) validate() error {
	if err := s.ServerBase.validate(); err != nil {
		return err
	}

	if v := s.Addr(); v != "" {
		dir := filepath.Dir(v)
		if exists, err := fnspath.IsExists(dir); err != nil {
			return fmt.Errorf("unix socket parent dir %q existence check failed: %w", dir, err)
		} else if !exists {
			return fmt.Errorf("error (:path %q): %w", v, ErrUnixSocketParentDirNotExists)
		}
	} else {
		return ErrUnixSocketPathNotProvided
	}

	return nil
}

func (s *ServerUNIX) defaultize() error {
	if err := s.ServerBase.defaultize(); err != nil {
		return err
	}

	if s.SocketFileMode == 0 {
		s.SocketFileMode = defaultUNIXSocketFileMode | os.ModeSocket
	}

	return nil
}

func (s *ServerUNIX) Dump(ctx *dumpctx.Ctx, w io.Writer) {
	fmt.Fprintf(w, "%saddr: %s\n", ctx.Indent(), s.Addr())
	fmt.Fprintf(w, "%ssocketFileMode: %03o | %s\n", ctx.Indent(), s.SocketFileMode, s.SocketFileMode)

	s.ServerBase.Dump(ctx, w)
}
