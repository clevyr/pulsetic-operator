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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// HTTPRouteReconciler reconciles a HTTPRoute object.
type HTTPRouteReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop.
func (r *HTTPRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	route := &gatewayv1.HTTPRoute{}
	if err := r.Get(ctx, req.NamespacedName, route); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	sr := &SourceReconciler{
		Client:   r.Client,
		Recorder: r.Recorder,
	}

	if err := sr.ReconcileSource(ctx, route, "HTTPRoute", r.getHTTPRouteValues); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HTTPRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1.HTTPRoute{}, builder.WithPredicates(
			predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{}),
		)).
		Named("httproute").
		Complete(r)
}

func (r *HTTPRouteReconciler) getHTTPRouteValues(obj client.Object, annotations map[string]string) (string, error) {
	route := obj.(*gatewayv1.HTTPRoute) //nolint:errcheck
	if _, ok := annotations["monitor.url"]; !ok {
		u := url.URL{
			Scheme: annotations["monitor.scheme"],
			Host:   annotations["monitor.host"],
			Path:   annotations["monitor.path"],
		}
		if u.Scheme == "" {
			u.Scheme = "https" // Default to https for routes unless specified
		}
		if u.Host == "" && len(route.Spec.Hostnames) != 0 {
			u.Host = string(route.Spec.Hostnames[0])
		}
		if u.Path == "" {
			u.Path = findFirstPath(route)
		}
		return u.String(), nil
	}
	return "", nil
}

func findFirstPath(route *gatewayv1.HTTPRoute) string {
	for _, rule := range route.Spec.Rules {
		for _, match := range rule.Matches {
			if match.Path == nil ||
				match.Path.Type == nil || *match.Path.Type == gatewayv1.PathMatchRegularExpression ||
				match.Path.Value == nil {
				continue
			}

			return *match.Path.Value
		}
	}
	return ""
}
