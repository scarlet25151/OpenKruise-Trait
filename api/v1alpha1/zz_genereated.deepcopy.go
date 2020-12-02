package v1alpha1

import "k8s.io/apimachinery/pkg/runtime"


func (in *Trait) DeepCopyInto(out *Trait)  {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
}

func (in *Trait) DeepCopy() *Trait {
	if in == nil {
		return nil
	}
	out := new(Trait)
	return out
}

func (in *Trait) DeepCopyObject() runtime.Object {
	panic("implement this")
}

func(in *TraitList) DeepCopyInto(out *Trait) {
	panic("implement this")
}

func (in *TraitList) DeepCopyObject() runtime.Object {
	panic("implement this")
}