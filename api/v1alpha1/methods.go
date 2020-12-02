package v1alpha1

import runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

func (in *Trait) SetConditions(c ...runtimev1alpha1.Condition) {
	in.Status.SetConditions(c...)
}

func (in *Trait) GetCondition(ct runtimev1alpha1.ConditionType) runtimev1alpha1.Condition {
	return in.Status.GetCondition(ct)
}

func (in *Trait) GetWorkloadReference() runtimev1alpha1.TypedReference {
	return in.Spec.WorkloadReference
}

func (in *Trait) SetWorkloadReference(r runtimev1alpha1.TypedReference) {
	in.Spec.WorkloadReference = r
}
