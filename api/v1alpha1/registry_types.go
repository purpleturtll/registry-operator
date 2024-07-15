/*
Copyright 2024 registry-operator authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageType string

const StorageTypeInMemory StorageType = "inmemory"

type Storage struct {
	// +kubebuilder:default="inmemory"
	// +kubebuilder:validation:Enum=inmemory
	Type StorageType `json:"type"`
}

// RegistrySpec defines the desired state of Registry.
type RegistrySpec struct {
	// +kubebuilder:default={"type": "inmemory"}
	// +kubebuilder:validation:Required
	Storage Storage `json:"storage"`
}

// +kubebuilder:validation:Enum=Pending;Running;Deleting
type RegistryPhase string

const (
	RegistryPhasePending  RegistryPhase = "Pending"
	RegistryPhaseRunning  RegistryPhase = "Running"
	RegistryPhaseDeleting RegistryPhase = "Deleting"
)

// RegistryStatus defines the observed state of Registry.
type RegistryStatus struct {
	// +kubebuilder:default="Pending"
	Phase RegistryPhase `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuildre:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="The current phase of the registry"
// Registry is the Schema for the registries API.
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:default={"storage": { "type": "inmemory"}}
	Spec RegistrySpec `json:"spec"`
	// +kubebuilder:default={phase:Pending}
	Status RegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// RegistryList contains a list of Registry.
type RegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Registry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Registry{}, &RegistryList{})
}
