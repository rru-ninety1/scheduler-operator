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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	crdv1alpha1 "github.com/rru-ninety1/scheduler-operator/api/v1alpha1"
)

// SchedulationReconciler reconciles a Schedulation object
type SchedulationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	OneShotExecutedSchedulationDeleteTime = time.Minute * 2
)

//+kubebuilder:rbac:groups=crd.rru.io,resources=schedulations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.rru.io,resources=schedulations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.rru.io,resources=schedulations/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *SchedulationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Get the logger
	log := log.FromContext(ctx)

	// Get the schedulation object
	schedulation := &crdv1alpha1.Schedulation{}
	if err := r.Get(ctx, req.NamespacedName, schedulation); err != nil {
		log.Error(err, "unable to fetch Schedulation")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If the schedulation status is not set, set it to Waiting
	if schedulation.Status.SchedulationExecutionStatus == "" {
		schedulation.Status.SchedulationExecutionStatus = crdv1alpha1.SchedulationExecutionStatusWaiting
	}

	if !schedulation.Spec.Suspended {
		// The schedulation is not suspended

		// Get current hour
		currentHour := int32(time.Now().Hour())

		isExecutionTime := false
		if schedulation.Spec.StartHour <= currentHour && schedulation.Spec.EndHour >= currentHour {
			// Now is beetwen the start and end time of the schedulation
			isExecutionTime = true
		}

		switch schedulation.Status.SchedulationExecutionStatus {
		case crdv1alpha1.SchedulationExecutionStatusRunning:
			// The schedulation is running
			executed, err := CheckSchedulationDesiredState(schedulation)
			if err != nil {
				log.Error(err, "Error checking Schedulation desired state")

				//TODO impostare la condizione di errore
			} else if executed {
				// The desired state is reached
				// Change the status to Executed and set the last execution time
				schedulation.Status.SchedulationExecutionStatus = crdv1alpha1.SchedulationExecutionStatusExecuted
				now := metav1.Now()
				schedulation.Status.LastExecutionTime = &now

				// Update the schedulation status
				r.UpdateSchedulationStatus(ctx, log, schedulation)

				// Record event SchedulationExecuted
				r.Recorder.Event(schedulation, "Normal", "SchedulationExecuted", "Schedulation executed")

				if schedulation.Spec.OneShot {
					// The schedulation is one shot
					//TODO: spostare
					// Requeue the schedulation to be deleted
					return ctrl.Result{RequeueAfter: OneShotExecutedSchedulationDeleteTime}, nil
				}
			}

		case crdv1alpha1.SchedulationExecutionStatusExecuted:
			// The schedulation is executed
			if schedulation.Spec.OneShot {
				// The schedulation is one shot
				lastExecution := schedulation.Status.LastExecutionTime.Time
				if time.Now().After(lastExecution.Add(OneShotExecutedSchedulationDeleteTime)) {
					// Last execution time is 2 minutes ago
					// Delete the schedulation
					if err := r.Delete(ctx, schedulation); err != nil {
						log.Error(err, "unable to delete Schedulation")

						return ctrl.Result{}, err
					}
				}
			} else if !isExecutionTime {
				// Change the status to Waiting
				schedulation.Status.SchedulationExecutionStatus = crdv1alpha1.SchedulationExecutionStatusWaiting

				// Update the schedulation status
				r.UpdateSchedulationStatus(ctx, log, schedulation)
			}

		case crdv1alpha1.SchedulationExecutionStatusError:
			// The schedulation is in error
			if !isExecutionTime {
				// Change the status to Waiting
				schedulation.Status.SchedulationExecutionStatus = crdv1alpha1.SchedulationExecutionStatusWaiting

				// Update the schedulation status
				r.UpdateSchedulationStatus(ctx, log, schedulation)
			}

		case crdv1alpha1.SchedulationExecutionStatusWaiting:
			// The schedulation is waiting
			if isExecutionTime {
				// Record event SchedulationStarted
				r.Recorder.Event(schedulation, "Normal", "SchedulationStarted", "Schedulation started")

				// Change the status to Running
				schedulation.Status.SchedulationExecutionStatus = crdv1alpha1.SchedulationExecutionStatusRunning

				// Update the schedulation status
				r.UpdateSchedulationStatus(ctx, log, schedulation)

				//TODO: eseguire la schedulazione
			}
		}

		//TODO impostare lo stato di errore alla schedulazione quando necessario
	}

	// Requeue after 10 minutes
	return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
}

// CheckSchedulationDesiredState checks if the desired state of the schedulation is reached
func CheckSchedulationDesiredState(schedulation *crdv1alpha1.Schedulation) (bool, error) {
	//TODO: controllare se lo stato desiderato è stato raggiunto

	return true, nil
}

// UpdateSchedulationStatus updates the status of the schedulation
func (r *SchedulationReconciler) UpdateSchedulationStatus(ctx context.Context, log logr.Logger, schedulation *crdv1alpha1.Schedulation) {
	if err := r.Status().Update(ctx, schedulation); err != nil {
		log.Error(err, "unable to update Schedulation status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SchedulationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1alpha1.Schedulation{}).
		Complete(r)
}
