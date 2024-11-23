package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Object[T any] interface {
	*T
	runtime.Object
	DeepCopyInto(*T)
}

type List[T any, P Object[T]] struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []T `json:"items"`
}

func (in *List[T, P]) DeepCopyObject() runtime.Object {
	out := List[T, P]{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]T, len(in.Items))
		for i := range in.Items {
			P(&in.Items[i]).DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
