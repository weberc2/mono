package mm

import (
	"encoding/json"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

// Slice is just a Go slice that marshals to JSON intuitively.
type Slice[T any] []T

// MarshalJSON implements the `json.Marshaler` interface.
func (s *Slice[T]) MarshalJSON() ([]byte, error) {
	if *s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]T(*s))
}

// Schema implements the `huma.SchemaProvider` interface.
func (s *Slice[T]) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:  "array",
		Items: r.Schema(reflect.TypeOf([0]T{}).Elem(), true, ""),
	}
}
