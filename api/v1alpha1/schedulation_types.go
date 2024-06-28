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
	// Type of the resource to control. Can be `Deployment` or `StatefulSet`
	Type ResourceType `json:"type,omitempty"`

	// Number of replicas of the resource to maintain for the scheduled period
	ReplicaCount int32 `json:"replicaCount,omitempty"`

	// Namespace of the resource
	Namespace string `json:"namespace,omitempty"`

	// Name of the resource
	Name string `json:"name,omitempty"`

	// Processing order of the resource
	Order int32 `json:"order,omitempty"`
}

// ResourceType defines the type of the resource to control
// +kubebuilder:validation:Enum=Deployment;StatefulSet
// +kubebuilder:default=Deployment
type ResourceType string

const (
	ResourceTypeDeployment ResourceType = "Deployment"

	ResourceTypeStatefulSet ResourceType = "StatefulSet"
)

// SchedulationStatus defines the observed state of Schedulation
type SchedulationStatus struct {
	//TODO: rimuovere
	// Status of the schedulation. Can be `Running`, `Executed`, `Error`, `Waiting`
	SchedulationExecutionStatus SchedulationExecutionStatus `json:"executionStatus,omitempty"`

	// Last execution time
	// +optional
	LastExecutionTime *metav1.Time `json:"lastExecutionTime,omitempty"`

	// Schedulation conditions
	Conditions []metav1.Condition `json:"conditions"`
}

const (
	// ConditionTypeStarted is set when the schedulation is started
	CondititionTypeStarted = "Started"

	// ConditionTypeExecuted is set when the schedulation is executed
	ConditionTypeExecuted = "Executed"

	// ConditionTypeError is set when the schedulation has an error
	ConditionTypeError = "Error"
)

// SetDefaultConditionsIfNotSet sets the default conditions if not set
func (schedulationStatus *SchedulationStatus) SetDefaultConditionsIfNotSet() {
	if schedulationStatus.Conditions == nil {
		schedulationStatus.Conditions = []metav1.Condition{}
	}

	// Check fi startd condition is set, if not set it
	if schedulationStatus.GetStartedCondition() == nil {
		schedulationStatus.SetStartedCondition(metav1.ConditionFalse, "NotStarted", "The schedulation is not started")
	}

	// Check if executed condition is set, if not set it
	if schedulationStatus.GetExecutedCondition() == nil {
		schedulationStatus.SetExecutedCondition(metav1.ConditionFalse, "NotExecuted", "The schedulation is not executed")
	}

	// Check if error condition is set, if not set it
	if schedulationStatus.GetErrorCondition() == nil {
		schedulationStatus.SetErrorCondition(metav1.ConditionFalse, "NoError", "The schedulation has no error")
	}
}

// SetStartedCondition sets the started condition
func (schedulationStatus *SchedulationStatus) SetStartedCondition(status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:    CondititionTypeStarted,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	setOrAddCondition(schedulationStatus, condition)
}

// GetStartedCondition gets the started condition
func (schedulationStatus *SchedulationStatus) GetStartedCondition() *metav1.Condition {
	return schedulationStatus.getCondtionByType(CondititionTypeStarted)
}

// SetExecutedCondition sets the executed condition
func (schedulationStatus *SchedulationStatus) SetExecutedCondition(status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:    ConditionTypeExecuted,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	setOrAddCondition(schedulationStatus, condition)
}

// GetExecutedCondition gets the executed condition
func (schedulationStatus *SchedulationStatus) GetExecutedCondition() *metav1.Condition {
	return schedulationStatus.getCondtionByType(ConditionTypeExecuted)
}

// SetErrorCondition sets the error condition
func (schedulationStatus *SchedulationStatus) SetErrorCondition(status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:    ConditionTypeError,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	setOrAddCondition(schedulationStatus, condition)
}

// GetErrorCondition gets the error condition
func (schedulationStatus *SchedulationStatus) GetErrorCondition() *metav1.Condition {
	return schedulationStatus.getCondtionByType(ConditionTypeError)
}

// getCondtionByType gets a condition by type
func (schedulationStatus *SchedulationStatus) getCondtionByType(conditionType string) *metav1.Condition {
	for _, c := range schedulationStatus.Conditions {
		if c.Type == conditionType {
			return &c
		}
	}

	return nil
}

// setOrAddCondition sets or adds a condition to the SchedulationStatus
func setOrAddCondition(schedulationStatus *SchedulationStatus, condition metav1.Condition) {
	condition.LastTransitionTime = metav1.Now()

	for i, c := range schedulationStatus.Conditions {
		if c.Type == condition.Type {
			schedulationStatus.Conditions[i] = condition
			return
		}
	}

	schedulationStatus.Conditions = append(schedulationStatus.Conditions, condition)
}

// TODO: rimuovere la parte sotto
// SchedulationExecutionStatus defines the possible status of a schedulation
// +kubebuilder:validation:Enum=Running;Executed;Error;Waiting
// +kubebuilder:default=Waiting
type SchedulationExecutionStatus string

const (
	// The schedulation is running
	SchedulationExecutionStatusRunning SchedulationExecutionStatus = "Running"

	// The schedulation has been executed
	SchedulationExecutionStatusExecuted SchedulationExecutionStatus = "Executed"

	// The schedulation has an error
	SchedulationExecutionStatusError SchedulationExecutionStatus = "Error"

	// The schedulation is waiting to be executed
	SchedulationExecutionStatusWaiting SchedulationExecutionStatus = "Waiting"
)

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
