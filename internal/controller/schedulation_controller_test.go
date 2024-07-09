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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crdv1alpha1 "github.com/rru-ninety1/scheduler-operator/api/v1alpha1"
)

var _ = Describe("Schedulation Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		schedulation := &crdv1alpha1.Schedulation{}

		BeforeEach(func() {
			By("Creating the custom resource for the Kind Schedulation")
			err := k8sClient.Get(ctx, typeNamespacedName, schedulation)
			if err != nil && errors.IsNotFound(err) {
				resource := &crdv1alpha1.Schedulation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &crdv1alpha1.Schedulation{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Schedulation")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &SchedulationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})

	Context("When reconciling a resource that does not exist", func() {
		const resourceName = "non-existent-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		It("should not return an error", func() {
			By("Reconciling the non-existent resource")
			controllerReconciler := &SchedulationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

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
			controllerReconciler := &SchedulationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

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
			controllerReconciler := &SchedulationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

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
})
