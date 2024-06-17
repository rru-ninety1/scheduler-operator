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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	crdv1alpha1 "github.com/rru-ninety1/scheduler-operator/api/v1alpha1"
)

// SchedulationReconciler reconciles a Schedulation object
type SchedulationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=crd.rru.io,resources=schedulations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.rru.io,resources=schedulations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.rru.io,resources=schedulations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Schedulation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *SchedulationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO(user): your logic here
	// Get the schedulation object
	schedulation := &crdv1alpha1.Schedulation{}
	if err := r.Get(ctx, req.NamespacedName, schedulation); err != nil {
		log.Error(err, "unable to fetch Schedulation")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !schedulation.Spec.Suspended {
		// The schedulation is not suspended

		// Get current hour
		currentHour := time.Now().Hour()

		isExecutionTime := false
		if schedulation.Spec.StartHour <= currentHour && schedulation.Spec.EndHour >= currentHour {
			// Now is beetwen the start and end time
			isExecutionTime = true
		}

		switch schedulation.Status.CurrentStatus {
		case "Running":
			// The schedulation is running
			//TODO: controllare se lo stato desiderato è stato raggiunto, in quel caso impostare lo stato a Executed
		case "Executed":
			// The schedulation is executed
			if !isExecutionTime {
				// Change the status to Waiting
				schedulation.Status.CurrentStatus = "Waiting"
			}
		case "Error":
			// The schedulation is in error
			if !isExecutionTime {
				// Change the status to Waiting
				schedulation.Status.CurrentStatus = "Waiting"
			}
		case "Waiting":
			// The schedulation is waiting
			if isExecutionTime {
				//TODO: eseguire la schedulazione
			}
		}

		//TODO gestire schedulazioni one shot
	}

	//TODO: se necessario, controllare la cancellazione e cosa fare in quel caso

	// Requeue after 10 minutes
	return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SchedulationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1alpha1.Schedulation{}).
		Complete(r)
}
