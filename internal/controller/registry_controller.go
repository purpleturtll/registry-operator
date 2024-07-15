package controller

import (
	"context"

	"github.com/registry-operator/registry-operator/api/v1alpha1"
	"github.com/registry-operator/registry-operator/internal/components"
	"github.com/registry-operator/registry-operator/internal/components/factories"
	"github.com/registry-operator/registry-operator/internal/state"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RegistryReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	PodFactory         *factories.PodFactory
	ConfigMapFactory   *factories.ConfigMapFactory
	RegistryOperations *components.RegistryOperations
}

// newReconciler initializes a new RegistryReconciler with dependencies.
func NewReconciler(client client.Client, scheme *runtime.Scheme) *RegistryReconciler {
	return &RegistryReconciler{
		Client:             client,
		Scheme:             scheme,
		RegistryOperations: components.NewRegistryOperations(client),
	}
}

// Reconcile is part of the main Kubernetes reconciliation loop.
func (r *RegistryReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	l := log.FromContext(ctx)
	registry := &v1alpha1.Registry{}
	if err := r.Get(ctx, request.NamespacedName, registry); err != nil {
		l.Info("Failed to get registry", "error", err)
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	var handler state.Handler
	switch registry.Status.Phase {
	case v1alpha1.RegistryPhasePending:
		handler = &state.Pending{RegistryOperations: r.RegistryOperations}
	case v1alpha1.RegistryPhaseRunning:
		handler = &state.Running{RegistryOperations: r.RegistryOperations}
	case v1alpha1.RegistryPhaseDeleting:
		handler = &state.Deleting{RegistryOperations: r.RegistryOperations}
	default:
		l.Error(nil, "Unknown registry phase", "phase", registry.Status.Phase)
		return reconcile.Result{}, nil
	}

	return handler.Handle(ctx, registry)
}

func (r *RegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Registry{}).
		Complete(r)
}
