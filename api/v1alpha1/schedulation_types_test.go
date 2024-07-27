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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("SetDefaultConditionsIfNotSet", func() {
	It("should set default conditions if not set", func() {
		schedulationStatus := &SchedulationStatus{}
		schedulationStatus.SetDefaultConditionsIfNotSet()

		Expect(schedulationStatus.Conditions).To(HaveLen(3))

		startedCondition := schedulationStatus.GetStartedCondition()
		Expect(startedCondition).NotTo(BeNil())
		Expect(startedCondition.Type).To(Equal(CondititionTypeStarted))
		Expect(startedCondition.Status).To(Equal(metav1.ConditionFalse))
		Expect(startedCondition.Reason).To(Equal(StartedConditionNotStartedReason))
		Expect(startedCondition.Message).To(Equal(StartedConditionNotStartedMessage))

		executedCondition := schedulationStatus.GetExecutedCondition()
		Expect(executedCondition).NotTo(BeNil())
		Expect(executedCondition.Type).To(Equal(ConditionTypeExecuted))
		Expect(executedCondition.Status).To(Equal(metav1.ConditionFalse))
		Expect(executedCondition.Reason).To(Equal(ExecutedConditionNotExecutedReason))
		Expect(executedCondition.Message).To(Equal(ExecutedConditionNotExecutedMessage))

		errorCondition := schedulationStatus.GetErrorCondition()
		Expect(errorCondition).NotTo(BeNil())
		Expect(errorCondition.Type).To(Equal(ConditionTypeError))
		Expect(errorCondition.Status).To(Equal(metav1.ConditionFalse))
		Expect(errorCondition.Reason).To(Equal(ErrorConditionNoErrorReason))
		Expect(errorCondition.Message).To(Equal(ErrorConditionNoErrorMessage))
	})
})

var _ = Describe("SetStartedCondition", func() {
	It("should set the started condition", func() {
		schedulationStatus := &SchedulationStatus{}
		schedulationStatus.SetStartedCondition(metav1.ConditionTrue, "Started", "The schedulation is started")

		startedCondition := schedulationStatus.GetStartedCondition()
		Expect(startedCondition).NotTo(BeNil())
		Expect(startedCondition.Type).To(Equal(CondititionTypeStarted))
		Expect(startedCondition.Status).To(Equal(metav1.ConditionTrue))
		Expect(startedCondition.Reason).To(Equal("Started"))
		Expect(startedCondition.Message).To(Equal("The schedulation is started"))
	})

	It("should update the started condition", func() {
		schedulationStatus := &SchedulationStatus{}
		schedulationStatus.SetStartedCondition(metav1.ConditionTrue, "Started", "The schedulation is started")
		schedulationStatus.SetStartedCondition(metav1.ConditionFalse, "NotStarted", "The schedulation is not started")

		startedCondition := schedulationStatus.GetStartedCondition()
		Expect(startedCondition).NotTo(BeNil())
		Expect(startedCondition.Type).To(Equal(CondititionTypeStarted))
		Expect(startedCondition.Status).To(Equal(metav1.ConditionFalse))
		Expect(startedCondition.Reason).To(Equal("NotStarted"))
		Expect(startedCondition.Message).To(Equal("The schedulation is not started"))
	})
})

var _ = Describe("GetStartedCondition", func() {
	It("should return the started condition", func() {
		schedulationStatus := &SchedulationStatus{
			Conditions: []metav1.Condition{
				{
					Type:    CondititionTypeStarted,
					Status:  metav1.ConditionTrue,
					Reason:  StartedConditionStartedReason,
					Message: StartedConditionStartedMessage,
				},
			},
		}

		startedCondition := schedulationStatus.GetStartedCondition()
		Expect(startedCondition).NotTo(BeNil())
		Expect(startedCondition.Type).To(Equal(CondititionTypeStarted))
		Expect(startedCondition.Status).To(Equal(metav1.ConditionTrue))
		Expect(startedCondition.Reason).To(Equal(StartedConditionStartedReason))
		Expect(startedCondition.Message).To(Equal(StartedConditionStartedMessage))
	})

	It("should return nil if the started condition is not set", func() {
		schedulationStatus := &SchedulationStatus{
			Conditions: []metav1.Condition{
				{
					Type:    ConditionTypeExecuted,
					Status:  metav1.ConditionTrue,
					Reason:  ExecutedConditionExecutedReason,
					Message: ExecutedConditionExecutedMessage,
				},
				{
					Type:    ConditionTypeError,
					Status:  metav1.ConditionFalse,
					Reason:  ErrorConditionNoErrorReason,
					Message: ErrorConditionNoErrorMessage,
				},
			},
		}

		startedCondition := schedulationStatus.GetStartedCondition()
		Expect(startedCondition).To(BeNil())
	})
})

var _ = Describe("SetExecutedCondition", func() {
	It("should set the executed condition", func() {
		schedulationStatus := &SchedulationStatus{}
		schedulationStatus.SetExecutedCondition(metav1.ConditionTrue, "Executed", "The schedulation is executed")

		executedCondition := schedulationStatus.GetExecutedCondition()
		Expect(executedCondition).NotTo(BeNil())
		Expect(executedCondition.Type).To(Equal(ConditionTypeExecuted))
		Expect(executedCondition.Status).To(Equal(metav1.ConditionTrue))
		Expect(executedCondition.Reason).To(Equal("Executed"))
		Expect(executedCondition.Message).To(Equal("The schedulation is executed"))
	})

	It("should update the executed condition", func() {
		schedulationStatus := &SchedulationStatus{}
		schedulationStatus.SetExecutedCondition(metav1.ConditionTrue, "Executed", "The schedulation is executed")
		schedulationStatus.SetExecutedCondition(metav1.ConditionFalse, "NotExecuted", "The schedulation is not executed")

		executedCondition := schedulationStatus.GetExecutedCondition()
		Expect(executedCondition).NotTo(BeNil())
		Expect(executedCondition.Type).To(Equal(ConditionTypeExecuted))
		Expect(executedCondition.Status).To(Equal(metav1.ConditionFalse))
		Expect(executedCondition.Reason).To(Equal("NotExecuted"))
		Expect(executedCondition.Message).To(Equal("The schedulation is not executed"))
	})
})

var _ = Describe("GetExecutedCondition", func() {
	It("should return the executed condition", func() {
		schedulationStatus := &SchedulationStatus{
			Conditions: []metav1.Condition{
				{
					Type:    ConditionTypeExecuted,
					Status:  metav1.ConditionTrue,
					Reason:  ExecutedConditionExecutedReason,
					Message: ExecutedConditionExecutedMessage,
				},
			},
		}

		executedCondition := schedulationStatus.GetExecutedCondition()
		Expect(executedCondition).NotTo(BeNil())
		Expect(executedCondition.Type).To(Equal(ConditionTypeExecuted))
		Expect(executedCondition.Status).To(Equal(metav1.ConditionTrue))
		Expect(executedCondition.Reason).To(Equal(ExecutedConditionExecutedReason))
		Expect(executedCondition.Message).To(Equal(ExecutedConditionExecutedMessage))
	})

	It("should return nil if the executed condition is not set", func() {
		schedulationStatus := &SchedulationStatus{
			Conditions: []metav1.Condition{
				{
					Type:    CondititionTypeStarted,
					Status:  metav1.ConditionTrue,
					Reason:  StartedConditionStartedReason,
					Message: StartedConditionStartedMessage,
				},
				{
					Type:    ConditionTypeError,
					Status:  metav1.ConditionFalse,
					Reason:  ErrorConditionNoErrorReason,
					Message: ErrorConditionNoErrorMessage,
				},
			},
		}

		executedCondition := schedulationStatus.GetExecutedCondition()
		Expect(executedCondition).To(BeNil())
	})
})

var _ = Describe("SetErrorCondition", func() {
	It("should set the error condition", func() {
		schedulationStatus := &SchedulationStatus{}
		schedulationStatus.SetErrorCondition(metav1.ConditionTrue, "Error", "An error occurred")

		errorCondition := schedulationStatus.GetErrorCondition()
		Expect(errorCondition).NotTo(BeNil())
		Expect(errorCondition.Type).To(Equal(ConditionTypeError))
		Expect(errorCondition.Status).To(Equal(metav1.ConditionTrue))
		Expect(errorCondition.Reason).To(Equal("Error"))
		Expect(errorCondition.Message).To(Equal("An error occurred"))
	})

	It("should update the error condition", func() {
		schedulationStatus := &SchedulationStatus{}
		schedulationStatus.SetErrorCondition(metav1.ConditionTrue, "Error", "An error occurred")
		schedulationStatus.SetErrorCondition(metav1.ConditionFalse, "NoError", "No error occurred")

		errorCondition := schedulationStatus.GetErrorCondition()
		Expect(errorCondition).NotTo(BeNil())
		Expect(errorCondition.Type).To(Equal(ConditionTypeError))
		Expect(errorCondition.Status).To(Equal(metav1.ConditionFalse))
		Expect(errorCondition.Reason).To(Equal("NoError"))
		Expect(errorCondition.Message).To(Equal("No error occurred"))
	})
})

var _ = Describe("GetErrorCondition", func() {
	It("should return the error condition", func() {
		schedulationStatus := &SchedulationStatus{
			Conditions: []metav1.Condition{
				{
					Type:    ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  ErrorConditionErrorReason,
					Message: "An error occurred",
				},
			},
		}

		errorCondition := schedulationStatus.GetErrorCondition()
		Expect(errorCondition).NotTo(BeNil())
		Expect(errorCondition.Type).To(Equal(ConditionTypeError))
		Expect(errorCondition.Status).To(Equal(metav1.ConditionTrue))
		Expect(errorCondition.Reason).To(Equal(ErrorConditionErrorReason))
		Expect(errorCondition.Message).To(Equal("An error occurred"))
	})

	It("should return nil if the error condition is not set", func() {
		schedulationStatus := &SchedulationStatus{
			Conditions: []metav1.Condition{
				{
					Type:    CondititionTypeStarted,
					Status:  metav1.ConditionTrue,
					Reason:  StartedConditionStartedReason,
					Message: StartedConditionStartedMessage,
				},
				{
					Type:    ConditionTypeExecuted,
					Status:  metav1.ConditionFalse,
					Reason:  ExecutedConditionNotExecutedReason,
					Message: ExecutedConditionNotExecutedMessage,
				},
			},
		}

		errorCondition := schedulationStatus.GetErrorCondition()
		Expect(errorCondition).To(BeNil())
	})
})
