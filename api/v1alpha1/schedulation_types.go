/*
Copyright 2024.

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

// SchedulationSpec defines the desired state of Schedulation
type SchedulationSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Schedulation suspended
	Suspended bool `json:"suspended,omitempty"`

	// Schedulation must be executed only one time
	OneShot bool `json:"oneShot,omitempty"`

	// Schedulation start hour
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=23
	StartHour int32 `json:"startHour,omitempty"`

	// Schedulation end hour
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=23
	EndHour int32 `json:"endHour,omitempty"`

	// Resources to control
	Resources []ScheduledResource `json:"resources,omitempty"`
}

type ScheduledResource struct {
	// `Deployment` or `StatefulSet`
	Type string `json:"type,omitempty"`

	// Number of replicas of the resource to maintain for the scheduled period
	ReplicaCount int32 `json:"replicaCount,omitempty"`

	// Namespace of the resource
	Namespace string `json:"namespace,omitempty"`

	// Name of the resource
	Name string `json:"name,omitempty"`
}

// SchedulationStatus defines the observed state of Schedulation
type SchedulationStatus struct {
	// Status of the schedulation. Can be `Running`, `Executed`, `Error`, `Waiting`
	CurrentStatus string `json:"currentStatus,omitempty"`

	// Error message, in case of error
	Error string `json:"error,omitempty"`

	// Last execution time
	LastExecutionTime string `json:"lastExecutionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Schedulation is the Schema for the schedulations API
type Schedulation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SchedulationSpec   `json:"spec,omitempty"`
	Status SchedulationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SchedulationList contains a list of Schedulation
type SchedulationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Schedulation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Schedulation{}, &SchedulationList{})
}
