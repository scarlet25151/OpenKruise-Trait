package v1alpha1

import runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

func (in *KruiseTrait) SetConditions(c ...runtimev1alpha1.Condition) {
	in.Status.SetConditions(c...)
}

func (in *KruiseTrait) GetCondition(ct runtimev1alpha1.ConditionType) runtimev1alpha1.Condition {
	return in.Status.GetCondition(ct)
}

func (in *KruiseTrait) GetWorkloadReference() runtimev1alpha1.TypedReference {
	return in.Spec.WorkloadReference
}

func (in *KruiseTrait) SetWorkloadReference(r runtimev1alpha1.TypedReference) {
	in.Spec.WorkloadReference = r
}
