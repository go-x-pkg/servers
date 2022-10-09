package servers

import (
	"fmt"
	"io"

	"github.com/go-x-pkg/dumpctx"
)

type WithKind struct {
	Knd Kind `json:"kind" yaml:"kind" bson:"kind"`
}

func (wk *WithKind) Kind() Kind { return wk.Knd }
func (wk *WithKind) validate() error {
	if wk.Knd.Has(KindUNIX) && wk.Knd.Has(KindINET) {
		return ErrGotBothInetAndUnix
	}

	return nil
}

func (wk *WithKind) defaultize() error {
	if !wk.Knd.Has(KindINET) && !wk.Knd.Has(KindUNIX) {
		wk.Knd.Set(KindINET)
	}

	if !wk.Knd.Has(KindHTTP) && !wk.Knd.Has(KindGRPC) {
		wk.Knd.Set(KindHTTP)
	}

	return nil
}

type WithNetwork struct {
	Net string `json:"network" yaml:"network" bson:"network"`
}

type ServerBase struct {
	WithKind    `json:",inline" yaml:",inline" bson:",inline"`
	WithNetwork `json:",inline" yaml:",inline" bson:",inline"`

	GRPC struct {
		Reflection bool `yaml:"reflection"`
	} `yaml:"grpc"`

	Pprof struct {
		Enable bool   `yaml:"enable"`
		Prefix string `yaml:"prefix"`
	} `yaml:"pprof"`
}

func (s *ServerBase) Network() string {
	if s.Net != "" {
		return s.Net
	}

	if s.Kind().Has(KindUNIX) {
		return "unix"
	}

	return "tcp"
}

func (s *ServerBase) validate() error {
	return s.WithKind.validate()
}

func (s *ServerBase) defaultize() error {
	if err := s.WithKind.defaultize(); err != nil {
		return err
	}

	if s.Pprof.Prefix == "" {
		s.Pprof.Prefix = defaultPprofPrefix
	}

	return nil
}

func (s *ServerBase) Dump(ctx *dumpctx.Ctx, w io.Writer) {
	fmt.Fprintf(w, "%sgrpc:\n", ctx.Indent())

	ctx.Wrap(func() {
		fmt.Fprintf(w, "%sreflection: %t\n", ctx.Indent(), s.GRPC.Reflection)
	})

	fmt.Fprintf(w, "%spprof:\n", ctx.Indent())

	ctx.Wrap(func() {
		fmt.Fprintf(w, "%senable: %t\n", ctx.Indent(), s.Pprof.Enable)
		fmt.Fprintf(w, "%sprefix: %q\n", ctx.Indent(), s.Pprof.Prefix)
	})
}
