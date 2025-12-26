package dsmspaces

import "encoding/json"

type Opt[T any] struct {
	v  T
	ok bool
}

func OptSome[T any](v T) (o Opt[T]) {
	o.v = v
	o.ok = true
	return
}

func OptNone[T any]() (o Opt[T]) {
	return
}

func (o Opt[T]) Get() (T, bool) {
	return o.v, o.ok
}

func (o *Opt[T]) UnmarshalText(data []byte) error {
	if len(data) < 1 || string(data) == "null" {
		*o = Opt[T]{}
		return nil
	}
	return json.Unmarshal(data, &o.v)
}

func (o Opt[T]) MarshalText() ([]byte, error) {
	if !o.ok {
		return []byte("null"), nil
	}
	return json.Marshal(o.v)
}
