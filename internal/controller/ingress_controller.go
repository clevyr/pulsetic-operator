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
	"net/url"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// IngressReconciler reconciles a Ingress object.
type IngressReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	ingress := &networkingv1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, ingress); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	sr := &SourceReconciler{
		Client:   r.Client,
		Recorder: r.Recorder,
	}

	if err := sr.ReconcileSource(ctx, ingress, "Ingress", r.getIngressValues); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}, builder.WithPredicates(
			predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{}),
		)).
		Named("ingress").
		Complete(r)
}

func (r *IngressReconciler) getIngressValues(obj client.Object, annotations map[string]string) (string, error) {
	ingress := obj.(*networkingv1.Ingress) //nolint:errcheck
	if _, ok := annotations["monitor.url"]; !ok {
		u := url.URL{
			Scheme: annotations["monitor.scheme"],
			Host:   annotations["monitor.host"],
			Path:   annotations["monitor.path"],
		}

		if u.Scheme == "" {
			if len(ingress.Spec.TLS) == 0 {
				u.Scheme = "http"
			} else {
				u.Scheme = "https"
			}
		}

		if len(ingress.Spec.Rules) != 0 {
			rule := ingress.Spec.Rules[0]
			if u.Host == "" {
				u.Host = rule.Host
			}
			if u.Path == "" && len(rule.HTTP.Paths) != 0 {
				if path := rule.HTTP.Paths[0].Path; path != "/" {
					u.Path = path
				}
			}
		}
		return u.String(), nil
	}
	return "", nil
}
