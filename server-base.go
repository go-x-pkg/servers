package servers

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
	return s.WithKind.defaultize()
}
