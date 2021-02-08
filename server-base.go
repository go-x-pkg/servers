package servers

type WithKind struct {
	Knd Kind `json:"kind" yaml:"kind" bson:"kind"`
}

func (wk *WithKind) Kind() Kind { return wk.Knd }

type ServerBase struct {
	WithKind `json:",inline" yaml:",inline" bson:",inline"`
}
