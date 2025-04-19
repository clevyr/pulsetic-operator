/*
Copyright 2025.

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
	"errors"

	"github.com/clevyr/pulsetic-operator/internal/pulsetic"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	pulseticv1 "github.com/clevyr/pulsetic-operator/api/v1"
)

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pulsetic.clevyr.com,resources=monitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pulsetic.clevyr.com,resources=monitors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pulsetic.clevyr.com,resources=monitors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *MonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	monitor := &pulseticv1.Monitor{}
	if err := r.Client.Get(ctx, req.NamespacedName, monitor); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	account := &pulseticv1.Account{}
	if err := GetAccount(ctx, r.Client, account, monitor.Spec.Account.Name); err != nil {
		return ctrl.Result{}, err
	}

	apiKey, err := GetAPIKey(ctx, r.Client, account)
	if err != nil {
		return ctrl.Result{}, err
	}
	psclient := pulsetic.NewClient(apiKey)

	const myFinalizerName = "pulsetic.clevyr.com/finalizer"
	if !monitor.DeletionTimestamp.IsZero() {
		// Object is being deleted
		if controllerutil.ContainsFinalizer(monitor, myFinalizerName) {
			if monitor.Spec.Prune && monitor.Status.Ready {
				if err := psclient.Monitors().Delete(ctx, monitor.Status.ID); err != nil {
					return ctrl.Result{}, err
				}
			}

			controllerutil.RemoveFinalizer(monitor, myFinalizerName)
			if err := r.Update(ctx, monitor); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	var create bool
	psmonitor, err := psclient.Monitors().Get(ctx, pulsetic.FindByID(monitor.Status.ID))
	if err != nil {
		if !errors.Is(err, pulsetic.ErrMonitorNotFound) {
			return ctrl.Result{}, err
		}

		psmonitor, err = psclient.Monitors().Get(ctx, pulsetic.FindByURL(monitor.Spec.Monitor.URL))
		if err != nil {
			if !errors.Is(err, pulsetic.ErrMonitorNotFound) {
				return ctrl.Result{}, err
			}
			create = true
		}
	}

	if create {
		psmonitor, err = psclient.Monitors().Create(ctx, monitor.Spec.Monitor.ToMonitor())
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		psmonitor, err = psclient.Monitors().Update(ctx, psmonitor.ID, monitor.Spec.Monitor.ToMonitor())
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	monitor.Status.Ready = true
	monitor.Status.ID = psmonitor.ID
	monitor.Status.Running = psmonitor.IsRunning
	if err := r.Status().Update(ctx, monitor); err != nil {
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(monitor, myFinalizerName) {
		controllerutil.AddFinalizer(monitor, myFinalizerName)
		if err := r.Update(ctx, monitor); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: monitor.Spec.Interval.Duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &pulseticv1.Monitor{}, "spec.sourceRef", func(rawObj client.Object) []string {
		monitor := rawObj.(*pulseticv1.Monitor)
		if monitor.Spec.SourceRef == nil {
			return nil
		}
		return []string{monitor.Spec.SourceRef.Kind + "/" + monitor.Spec.SourceRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&pulseticv1.Monitor{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("monitor").
		Complete(r)
}
