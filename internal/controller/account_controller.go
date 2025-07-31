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
	"fmt"

	pulseticv1 "github.com/clevyr/pulsetic-operator/api/v1"
	"github.com/clevyr/pulsetic-operator/internal/pulsetic"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//nolint:gochecknoglobals
var ClusterResourceNamespace = "pulsetic-system"

// AccountReconciler reconciles a Account object.
type AccountReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

var ErrKeyNotFound = errors.New("secret key not found")

//+kubebuilder:rbac:groups=pulsetic.clevyr.com,resources=accounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pulsetic.clevyr.com,resources=accounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pulsetic.clevyr.com,resources=accounts/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *AccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	account := &pulseticv1.Account{}
	if err := r.Get(ctx, req.NamespacedName, account); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	apiKey, err := GetAPIKey(ctx, r.Client, account)
	if err != nil {
		r.Recorder.Event(account, "Warning", "GetAPIKeyFailed", err.Error())
		return ctrl.Result{}, err
	}

	psclient := pulsetic.NewClient(apiKey)
	for _, err := range psclient.Monitors().List(ctx) {
		if err != nil {
			account.Status.Ready = false
			if err := r.Status().Update(ctx, account); err != nil {
				r.Recorder.Event(account, "Warning", "UpdateStatusFailed", err.Error())
				return ctrl.Result{}, err
			}

			r.Recorder.Event(account, "Warning", "AuthenticationFailed", err.Error())
			return ctrl.Result{}, err
		}
		break
	}

	account.Status.Ready = true
	if err := r.Status().Update(ctx, account); err != nil {
		r.Recorder.Event(account, "Warning", "UpdateStatusFailed", err.Error())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &pulseticv1.Account{}, "spec.isDefault", func(rawObj client.Object) []string {
		account := rawObj.(*pulseticv1.Account) //nolint:errcheck
		if !account.Spec.IsDefault {
			return nil
		}
		return []string{"true"}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&pulseticv1.Account{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("account").
		Complete(r)
}

var (
	ErrNoDefaultAccount       = errors.New("no default account")
	ErrMultipleDefaultAccount = errors.New("more than 1 default account found")
)

func GetAccount(ctx context.Context, c client.Client, account *pulseticv1.Account, name string) error {
	if name != "" {
		return c.Get(ctx, client.ObjectKey{Name: name}, account)
	}

	list := &pulseticv1.AccountList{}
	err := c.List(ctx, list, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.isDefault", "true"),
	})
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		return ErrNoDefaultAccount
	}
	if len(list.Items) > 1 {
		return ErrMultipleDefaultAccount
	}

	*account = list.Items[0]
	return nil
}

func GetAPIKey(ctx context.Context, c client.Client, account *pulseticv1.Account) (string, error) {
	secret := &corev1.Secret{}
	err := c.Get(ctx, client.ObjectKey{
		Namespace: ClusterResourceNamespace,
		Name:      account.Spec.APIKeySecretRef.Name,
	}, secret)
	if err != nil {
		return "", err
	}

	apiKey, ok := secret.Data[account.Spec.APIKeySecretRef.Key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, account.Spec.APIKeySecretRef.Key)
	}

	return string(apiKey), nil
}
