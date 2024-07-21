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

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crdv1alpha1 "github.com/rru-ninety1/scheduler-operator/api/v1alpha1"
)

func getSchedulationReconciler() *SchedulationReconciler {
	return &SchedulationReconciler{
		Client:   k8sClient,
		Scheme:   k8sClient.Scheme(),
		Recorder: record.NewFakeRecorder(3),
	}
}

var _ = Describe("Schedulation Controller", func() {
	// Context("When reconciling a resource", func() {
	// 	const resourceName = "test-resource"

	// 	ctx := context.Background()

	// 	typeNamespacedName := types.NamespacedName{
	// 		Name:      resourceName,
	// 		Namespace: "default",
	// 	}
	// 	schedulation := &crdv1alpha1.Schedulation{}

	// 	BeforeEach(func() {
	// 		By("Creating the custom resource for the Kind Schedulation")
	// 		err := k8sClient.Get(ctx, typeNamespacedName, schedulation)
	// 		if err != nil && errors.IsNotFound(err) {
	// 			resource := &crdv1alpha1.Schedulation{
	// 				ObjectMeta: metav1.ObjectMeta{
	// 					Name:      resourceName,
	// 					Namespace: "default",
	// 				},
	// 			}
	// 			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
	// 		}
	// 	})

	// 	AfterEach(func() {
	// 		// TODO(user): Cleanup logic after each test, like removing the resource instance.
	// 		resource := &crdv1alpha1.Schedulation{}
	// 		err := k8sClient.Get(ctx, typeNamespacedName, resource)
	// 		Expect(err).NotTo(HaveOccurred())

	// 		By("Cleanup the specific resource instance Schedulation")
	// 		Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
	// 	})

	// 	It("should successfully reconcile the resource", func() {
	// 		By("Reconciling the created resource")
	// 		controllerReconciler := &SchedulationReconciler{
	// 			Client: k8sClient,
	// 			Scheme: k8sClient.Scheme(),
	// 		}

	// 		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
	// 			NamespacedName: typeNamespacedName,
	// 		})
	// 		Expect(err).NotTo(HaveOccurred())
	// 	})
	// })

	Context("When reconciling a resource that does not exist", func() {
		const resourceName = "non-existent-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		It("should not return an error", func() {
			By("Reconciling the non-existent resource")
			controllerReconciler := getSchedulationReconciler()

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When reconciling a one-shot executed schedulation", func() {
		const resourceName = "test-resource"
		const resourceNamespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}

		BeforeEach(func() {
			By("Creating the custom resource for the Kind Schedulation")
			schedulation := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, schedulation)
			if err != nil && errors.IsNotFound(err) {
				schedulation = &crdv1alpha1.Schedulation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: crdv1alpha1.SchedulationSpec{
						OneShot:   true,
						Suspended: false,
					},
				}
				Expect(k8sClient.Create(ctx, schedulation)).To(Succeed())
			}

			schedulation.Status.SetDefaultConditionsIfNotSet()
			schedulation.Status.SetExecutedCondition(metav1.ConditionTrue, crdv1alpha1.ExecutedConditionExecutedReason, crdv1alpha1.ExecutedConditionExecutedMessage)

			By("Set executed condition to true")
			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())
		})

		AfterEach(func() {
			resource := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)

			if !errors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance Schedulation")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should delete the Schedulation, if enough time has passed", func() {
			schedulation := &crdv1alpha1.Schedulation{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, schedulation)).To(Succeed())

			By("Set last execution time to a time that is older than the delete time")
			schedulation.Status.LastExecutionTime = &metav1.Time{
				Time: metav1.Now().Add(-(OneShotExecutedSchedulationDeleteTime + time.Minute)),
			}
			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())

			// Reconcile the Schedulation
			By("Reconciling the created resource")
			controllerReconciler := getSchedulationReconciler()

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			// Check if the Schedulation has been deleted
			By("Checking if the Schedulation has been deleted")
			schedulation = &crdv1alpha1.Schedulation{}
			err = k8sClient.Get(ctx, typeNamespacedName, schedulation)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should not delete the Schedulation, if not enough time has passed", func() {
			schedulation := &crdv1alpha1.Schedulation{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, schedulation)).To(Succeed())

			By("Set last execution time to a time that is not older than the delete time")
			schedulation.Status.LastExecutionTime = &metav1.Time{
				Time: metav1.Now().Time,
			}
			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())

			// Reconcile the Schedulation
			By("Reconciling the created resource")
			controllerReconciler := getSchedulationReconciler()

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(OneShotExecutedSchedulationDeleteTime))

			// Check if the Schedulation has been deleted
			By("Checking if the Schedulation has been deleted")
			schedulation = &crdv1alpha1.Schedulation{}
			err = k8sClient.Get(ctx, typeNamespacedName, schedulation)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When reconciling a suspended schedulation", func() {
		const resourceName = "test-resource"
		const resourceNamespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}

		BeforeEach(func() {
			By("Creating the custom resource for the Kind Schedulation")
			schedulation := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, schedulation)
			if err != nil && errors.IsNotFound(err) {
				schedulation = &crdv1alpha1.Schedulation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: crdv1alpha1.SchedulationSpec{
						Suspended: true,
					},
				}
				Expect(k8sClient.Create(ctx, schedulation)).To(Succeed())
			}

			By("Set default conditions")
			schedulation.Status.SetDefaultConditionsIfNotSet()
			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())
		})

		AfterEach(func() {
			resource := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)

			if !errors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance Schedulation")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should not execute the Schedulation and not requeue it", func() {
			schedulation := &crdv1alpha1.Schedulation{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, schedulation)).To(Succeed())

			// Reconcile the Schedulation
			By("Reconciling the created resource")
			controllerReconciler := getSchedulationReconciler()

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())
			Expect(result.RequeueAfter).To(Equal(time.Duration(0)))

			// Check if the Schedulation has not been executed or started
			By("Checking if the Schedulation has not been executed or started")
			schedulation = &crdv1alpha1.Schedulation{}
			err = k8sClient.Get(ctx, typeNamespacedName, schedulation)
			Expect(err).NotTo(HaveOccurred())
			Expect(schedulation.Status.GetExecutedCondition().Status).To(Equal(metav1.ConditionFalse))
			Expect(schedulation.Status.GetStartedCondition().Status).To(Equal(metav1.ConditionFalse))
		})
	})

	Context("When reconciling a schedulation that is not in execution time", func() {
		const resourceName = "test-resource"
		const resourceNamespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}

		BeforeEach(func() {
			By("Creating the custom resource for the Kind Schedulation")
			schedulation := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, schedulation)
			if err != nil && errors.IsNotFound(err) {
				schedulation = &crdv1alpha1.Schedulation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: crdv1alpha1.SchedulationSpec{
						Suspended: false,
						StartHour: int32(time.Now().Hour()) - 2,
						EndHour:   int32(time.Now().Hour()) - 1,
					},
				}
				Expect(k8sClient.Create(ctx, schedulation)).To(Succeed())
			}

			By("Set set default conditions")
			schedulation.Status.SetDefaultConditionsIfNotSet()
			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())
		})

		AfterEach(func() {
			resource := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)

			if !errors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance Schedulation")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should not execute the Schedulation and requeue it", func() {
			schedulation := &crdv1alpha1.Schedulation{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, schedulation)).To(Succeed())

			// Reconcile the Schedulation
			By("Reconciling the created resource")
			controllerReconciler := getSchedulationReconciler()

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(DefaultRequeueTime))

			// Check if the Schedulation has not been started
			By("Checking if the Schedulation has not been started")
			schedulation = &crdv1alpha1.Schedulation{}
			err = k8sClient.Get(ctx, typeNamespacedName, schedulation)
			Expect(err).NotTo(HaveOccurred())
			Expect(schedulation.Status.GetStartedCondition().Status).To(Equal(metav1.ConditionFalse))
		})

		It("should set the started condition to false, if it's true", func() {
			schedulation := &crdv1alpha1.Schedulation{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, schedulation)).To(Succeed())

			By("Set started condition to true")
			schedulation.Status.SetStartedCondition(metav1.ConditionTrue, crdv1alpha1.StartedConditionStartedReason, crdv1alpha1.StartedConditionStartedMessage)
			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())

			// Reconcile the Schedulation
			By("Reconciling the created resource")
			controllerReconciler := getSchedulationReconciler()

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Check if the Schedulation has not been started
			schedulation = &crdv1alpha1.Schedulation{}
			err = k8sClient.Get(ctx, typeNamespacedName, schedulation)
			Expect(err).NotTo(HaveOccurred())
			Expect(schedulation.Status.GetStartedCondition().Status).To(Equal(metav1.ConditionFalse))
		})
	})

	Context("When reconciling a schedulation that is in execution time", func() {
		const resourceName = "test-resource"
		const resourceNamespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}

		BeforeEach(func() {
			By("Creating the custom resource for the Kind Schedulation")
			schedulation := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, schedulation)
			if err != nil && errors.IsNotFound(err) {
				schedulation = &crdv1alpha1.Schedulation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: crdv1alpha1.SchedulationSpec{
						Suspended: false,
						StartHour: 0,
						EndHour:   23,
					},
				}
				Expect(k8sClient.Create(ctx, schedulation)).To(Succeed())
			}

			By("Set set default conditions")
			schedulation.Status.SetDefaultConditionsIfNotSet()
			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())
		})

		AfterEach(func() {
			resource := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)

			if !errors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance Schedulation")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should requeue the schedulation, if it's already executed", func() {
			schedulation := &crdv1alpha1.Schedulation{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, schedulation)).To(Succeed())

			By("Set started and executed condition to true")
			schedulation.Status.SetStartedCondition(metav1.ConditionTrue, crdv1alpha1.StartedConditionStartedReason, crdv1alpha1.StartedConditionStartedMessage)
			schedulation.Status.SetExecutedCondition(metav1.ConditionTrue, crdv1alpha1.ExecutedConditionExecutedReason, crdv1alpha1.ExecutedConditionExecutedMessage)

			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())

			// Reconcile the Schedulation
			By("Reconciling the created resource")
			controllerReconciler := getSchedulationReconciler()

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(DefaultRequeueTime))
		})

		It("should start the schedulation, if it's not started", func() {
			// Reconcile the Schedulation
			By("Reconciling the created resource")
			controllerReconciler := getSchedulationReconciler()

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(DefaultRequeueTime))

			// Check if the Schedulation has been started
			By("Checking if the Schedulation has been started")
			schedulation := &crdv1alpha1.Schedulation{}
			err = k8sClient.Get(ctx, typeNamespacedName, schedulation)
			Expect(err).NotTo(HaveOccurred())
			Expect(schedulation.Status.GetStartedCondition().Status).To(Equal(metav1.ConditionTrue))

			By("Checking if the last execution time has been set")
			Expect(schedulation.Status.LastExecutionTime).NotTo(BeNil())
		})

		It("should continue the schedulation, if it's already started", func() {
			schedulation := &crdv1alpha1.Schedulation{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, schedulation)).To(Succeed())

			By("Set started condition to true")
			schedulation.Status.SetStartedCondition(metav1.ConditionTrue, crdv1alpha1.StartedConditionStartedReason, crdv1alpha1.StartedConditionStartedMessage)

			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())

			// Reconcile the Schedulation
			By("Reconciling the created resource")
			controllerReconciler := getSchedulationReconciler()

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Check if the Schedulation has been executed
			By("Checking if the Schedulation has been executed")
			err = k8sClient.Get(ctx, typeNamespacedName, schedulation)
			Expect(err).NotTo(HaveOccurred())
			Expect(schedulation.Status.GetExecutedCondition().Status).To(Equal(metav1.ConditionTrue))

		})

		//TODO test per deployment
		//TODO test per statefulset
		//TODO test con più risorse
	})

	When("Reconciling a one-shot schedulation in execution time", func() {
		const resourceName = "test-resource"
		const resourceNamespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}

		BeforeEach(func() {
			By("Creating the custom resource for the Kind Schedulation")
			schedulation := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, schedulation)
			if err != nil && errors.IsNotFound(err) {
				schedulation = &crdv1alpha1.Schedulation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: crdv1alpha1.SchedulationSpec{
						Suspended: false,
						OneShot:   true,
						StartHour: 0,
						EndHour:   23,
					},
				}
				Expect(k8sClient.Create(ctx, schedulation)).To(Succeed())
			}

			By("Set set default conditions")
			schedulation.Status.SetDefaultConditionsIfNotSet()
			Expect(k8sClient.Status().Update(ctx, schedulation)).To(Succeed())
		})

		AfterEach(func() {
			resource := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)

			if !errors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())

				By("Cleanup the specific resource instance Schedulation")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should requeue the schedulation to be deleted, after executing it", func() {
			// Reconcile the Schedulation
			By("Reconciling the created resource")
			controllerReconciler := getSchedulationReconciler()

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(OneShotExecutedSchedulationDeleteTime))
		})
	})
})
