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

	// Set default status conditions if not set
	schedulation.Status.SetDefaultConditionsIfNotSet()

	if schedulation.Spec.OneShot {
		// The schedulation is one shot
		// Get executed condition
		executedCondition := schedulation.Status.GetExecutedCondition()

		if executedCondition.Status == metav1.ConditionTrue {
			// The schedulation is executed
			lastExecution := schedulation.Status.LastExecutionTime.Time
			if time.Now().After(lastExecution.Add(OneShotExecutedSchedulationDeleteTime)) {
				// It's time to delete the schedulation
				// Delete the schedulation
				if err := r.Delete(ctx, schedulation); err != nil {
					log.Error(err, "unable to delete Schedulation")

					return ctrl.Result{}, err
				}
			}

			// Requeue the schedulation to be deleted
			return ctrl.Result{RequeueAfter: OneShotExecutedSchedulationDeleteTime}, nil
		}
	}

	if !schedulation.Spec.Suspended {
		// The schedulation is not suspended

		// Get current hour
		currentHour := int32(time.Now().Hour())

		if schedulation.Spec.StartHour <= currentHour && schedulation.Spec.EndHour >= currentHour {
			// Now is beetwen the start and end time of the schedulation
			return r.reconcileExecutionTime(ctx, log, schedulation)
		} else {
			return r.reconcileNotExecutionTime(ctx, log, schedulation)
		}

	}

	// Requeue after 10 minutes
	return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
}

func (r *SchedulationReconciler) reconcileExecutionTime(ctx context.Context, log logr.Logger, schedulation *crdv1alpha1.Schedulation) (ctrl.Result, error) {

	if schedulation.Status.GetStartedCondition().Status != metav1.ConditionTrue {
		// The schedulation is not started
		// Update the conditions
		schedulation.Status.SetStartedCondition(metav1.ConditionTrue, "Started", "The schedulation is started")
		schedulation.Status.SetExecutedCondition(metav1.ConditionFalse, "NotExecuted", "The schedulation is not executed")
		schedulation.Status.SetErrorCondition(metav1.ConditionFalse, "NoError", "The schedulation has no error")

		if err := r.Status().Update(ctx, schedulation); err != nil {
			log.Error(err, "unable to update Schedulation status")

			return ctrl.Result{}, err
		}

		// Record event SchedulationStarted
		r.Recorder.Event(schedulation, "Normal", "SchedulationStarted", "Schedulation started")

		// Run the schedulation
		return r.runSchedulation(ctx, log, schedulation)

	} else if schedulation.Status.GetExecutedCondition().Status != metav1.ConditionTrue {
		// The schedulation is started but not executed

		//TODO vedere a che punto è la schedulazione
		//proseguire con l'esecuzione
		return r.runSchedulation(ctx, log, schedulation)
	}

	// Requeue after 10 minutes
	return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
}

func (r *SchedulationReconciler) runSchedulation(ctx context.Context, log logr.Logger, schedulation *crdv1alpha1.Schedulation) (ctrl.Result, error) {

	// TODO eseguire, tornare errore in caso di errore, altrimenti, alla fine

	// Set the last execution time
	now := metav1.Now()
	schedulation.Status.LastExecutionTime = &now

	schedulation.Status.SetExecutedCondition(metav1.ConditionTrue, "Executed", "The schedulation is executed")

	// Update the schedulation status
	if err := r.Status().Update(ctx, schedulation); err != nil {
		log.Error(err, "unable to update Schedulation status")

		return ctrl.Result{}, err
	}

	// Record event SchedulationExecuted
	r.Recorder.Event(schedulation, "Normal", "SchedulationExecuted", "Schedulation executed")

	if schedulation.Spec.OneShot {
		// The schedulation is one shot

		// Requeue the schedulation to be deleted
		return ctrl.Result{RequeueAfter: OneShotExecutedSchedulationDeleteTime}, nil
	} else {
		// Requeue after 10 minutes
		return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
	}
}

func (r *SchedulationReconciler) reconcileNotExecutionTime(ctx context.Context, log logr.Logger, schedulation *crdv1alpha1.Schedulation) (ctrl.Result, error) {
	// Get started condition
	startedCondition := schedulation.Status.GetStartedCondition()

	if startedCondition.Status != metav1.ConditionFalse {
		// Update the started condition
		schedulation.Status.SetStartedCondition(metav1.ConditionFalse, "NotStarted", "The schedulation is not started")

		if err := r.Status().Update(ctx, schedulation); err != nil {
			log.Error(err, "unable to update Schedulation status")

			return ctrl.Result{}, err
		}
	}

	if schedulation.Spec.OneShot {
		// The schedulation is one shot

		// Requeue the schedulation to be deleted
		return ctrl.Result{RequeueAfter: OneShotExecutedSchedulationDeleteTime}, nil
	}

	// Requeue after 10 minutes
	return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
}

// checkSchedulationDesiredState checks if the desired state of the schedulation is reached
func checkSchedulationDesiredState(schedulation *crdv1alpha1.Schedulation) (bool, error) {
	//TODO: controllare se lo stato desiderato è stato raggiunto

	return true, nil
}

// updateSchedulationStatus updates the status of the schedulation
func (r *SchedulationReconciler) updateSchedulationStatus(ctx context.Context, log logr.Logger, schedulation *crdv1alpha1.Schedulation) {
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
