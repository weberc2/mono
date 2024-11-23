package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type UserList List[User, *User]

func (users *UserList) DeepCopyObject() runtime.Object {
	return (*List[User, *User])(users).DeepCopyObject()
}

type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	UserName          string `json:"userName"`
}

func (in *User) DeepCopyInto(out *User) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.UserName = in.UserName
}

func (in *User) DeepCopyObject() runtime.Object {
	var out User
	in.DeepCopyInto(&out)
	return &out
}
