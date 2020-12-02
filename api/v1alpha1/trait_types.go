/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/oam-kubernetes-runtime/pkg/oam"
	"github.com/openkruise/kruise-api/apps/v1alpha1"
)

var _ oam.Trait = &Trait{}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// A DefinitionReference refers to a CustomResourceDefinition by name.
type DefinitionReference struct {
	// Name of the referenced CustomResourceDefinition.
	Name string `json:"name"`

	// Version indicate which version should be used if CRD has multiple versions
	// by default it will use the first one if not specified
	Version string `json:"version,omitempty"`
}

// TraitSpec defines the desired state of Trait
type TraitSpec struct {
	// ReplicaCount of the workload this trait applies to.
	ReplicaCount int32

	// CloneSetSpec defined the desird state of CloneSet
	CloneSetSpec v1alpha1.CloneSetSpec

	v1alpha1.AdvancedStatefulSetTemplateSpec

	v1alpha1.UnitedDeploymentSpec

	v1alpha1.SidecarSetSpec

	v1alpha1.DaemonSetSpec

	// WorkloadReference to the workload this trait applies to.
	WorkloadReference runtimev1alpha1.TypedReference `json:"workloadRef"`

	// Revision indicates whether a trait is aware of component revision
	// +optional
	RevisionEnabled bool `json:"revisionEnabled,omitempty"`
}

type PatchConfigMap string

// TraitStatus defines the observed state of Trait
type TraitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// PatchConfigMap to the configmap that trait patch to.
	PatchConfigMap PatchConfigMap

	//
	runtimev1alpha1.ConditionedStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// Trait is the Schema for the traits API
type Trait struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TraitSpec   `json:"spec,omitempty"`
	Status TraitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TraitList contains a list of Trait
type TraitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Trait `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Trait{}, &TraitList{})
}
