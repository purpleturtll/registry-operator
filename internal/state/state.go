package state

import (
	"context"

	"github.com/registry-operator/registry-operator/api/v1alpha1"
	"github.com/registry-operator/registry-operator/internal/components"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Handler interface {
	Handle(ctx context.Context, registry *v1alpha1.Registry) (reconcile.Result, error)
}

// Pending ---Pod creation---> Running.
type Pending struct {
	RegistryOperations *components.RegistryOperations
}

func (s *Pending) Handle(ctx context.Context, registry *v1alpha1.Registry) (reconcile.Result, error) {
	l := log.FromContext(ctx)

	if registry.Spec.Storage.Type == "" {
		// We got registry before defaults were applied.
		// We should requeue the request to get the registry with defaults.
		return reconcile.Result{Requeue: true}, nil
	}

	// Create the ConfigMap for the registry if it doesn't exist.
	exists, err := s.RegistryOperations.CheckRegistryConfigMapExists(ctx, registry)
	if err != nil {
		l.Error(err, "Failed to check if the ConfigMap exists", "name", registry.Name)
		return reconcile.Result{}, err
	}

	if !exists {
		err = s.RegistryOperations.CreateRegistryConfigMap(ctx, registry)
		if err != nil {
			l.Error(err, "Failed to create the ConfigMap", "name", registry.Name)
			return reconcile.Result{}, err
		}
	}

	// Create the pod for the registry if it doesn't exist.
	exists, err = s.RegistryOperations.CheckRegistryPodExists(ctx, registry)
	if err != nil {
		l.Error(err, "Failed to check if the pod exists", "name", registry.Name)
		return reconcile.Result{}, err
	}

	if exists {
		// Add finalizer to the registry.
		err = s.RegistryOperations.AddFinalizer(ctx, registry)
		if err != nil {
			l.Error(err, "Failed to add finalizer to the registry", "name", registry.Name)
			return reconcile.Result{}, err
		}

		// If the pod already exists, move to the Running state.
		registry.Status.Phase = v1alpha1.RegistryPhaseRunning
		err = s.RegistryOperations.UpdateRegistryStatus(ctx, registry)
		if err != nil {
			l.Error(err, "Failed to update the registry status", "name", registry.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// If the pod doesn't exist, create it.
	err = s.RegistryOperations.CreateRegistryPod(ctx, registry)
	if err != nil {
		l.Error(err, "Failed to create or update the pod", "name", registry.Name)
		return reconcile.Result{}, err
	}

	// Add finalizer to the registry.
	err = s.RegistryOperations.AddFinalizer(ctx, registry)
	if err != nil {
		l.Error(err, "Failed to add finalizer to the registry", "name", registry.Name)
		return reconcile.Result{}, err
	}

	// If the pod is created, move to the Running state.
	registry.Status.Phase = v1alpha1.RegistryPhaseRunning
	err = s.RegistryOperations.UpdateRegistryStatus(ctx, registry)
	if err != nil {
		l.Error(err, "Failed to update the registry status", "name", registry.Name)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// Running ---Registry deletion---> Deleting.
type Running struct {
	RegistryOperations *components.RegistryOperations
}

func (s *Running) Handle(ctx context.Context, registry *v1alpha1.Registry) (reconcile.Result, error) {
	l := log.FromContext(ctx)

	if registry.DeletionTimestamp.IsZero() {
		// Here we could check if there were configuration changes and add states for updates to child resources.
		// We could compare checksum of current registry spec with checksum of registry spec
		// that was generated and saved on resources annotation when they were being first created.
		return reconcile.Result{}, nil
	}

	// If the registry is being deleted, move to the Deleting state.
	registry.Status.Phase = v1alpha1.RegistryPhaseDeleting
	err := s.RegistryOperations.UpdateRegistryStatus(ctx, registry)
	if err != nil {
		l.Error(err, "Failed to update the registry status", "name", registry.Name)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// Deleting - remove all resources tied to the registry.
type Deleting struct {
	RegistryOperations *components.RegistryOperations
}

func (s *Deleting) Handle(ctx context.Context, registry *v1alpha1.Registry) (reconcile.Result, error) {
	l := log.FromContext(ctx)

	// Check if finalizer is present on the registry.
	exists, err := s.RegistryOperations.CheckFinalizerExists(ctx, registry)
	if err != nil {
		l.Error(err, "Failed to check if the finalizer exists", "name", registry.Name)
		return reconcile.Result{}, err
	}

	if !exists {
		// If the finalizer is not present, the registry is already deleted.
		return reconcile.Result{}, nil
	}

	// This block will probably be something reoccuring for every resoure that we have to delete.
	// It may be a good idea to extract this to a separate function if it happens.
	{
		// Delete the pod for the registry.
		exists, err = s.RegistryOperations.CheckRegistryPodExists(ctx, registry)
		if err != nil {
			l.Error(err, "Failed to check if the pod exists", "name", registry.Name)
			return reconcile.Result{}, err
		}

		if !exists {
			// If the pod doesn't exist, remove the finalizer and finish the deletion.
			err = s.RegistryOperations.RemoveFinalizer(ctx, registry)
			if err != nil {
				l.Error(err, "Failed to remove finalizer from the registry", "name", registry.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}

		err = s.RegistryOperations.DeleteRegistryPod(ctx, registry)
		if err != nil {
			l.Error(err, "Failed to delete the pod", "name", registry.Name)
			return reconcile.Result{}, err
		}
	}

	// Delete the ConfigMap for the registry.
	exists, err = s.RegistryOperations.CheckRegistryConfigMapExists(ctx, registry)
	if err != nil {
		l.Error(err, "Failed to check if the ConfigMap exists", "name", registry.Name)
		return reconcile.Result{}, err
	}

	if exists {
		err = s.RegistryOperations.DeleteRegistryConfigMap(ctx, registry)
		if err != nil {
			l.Error(err, "Failed to delete the ConfigMap", "name", registry.Name)
			return reconcile.Result{}, err
		}
	}

	// Remove the finalizer from the registry.
	err = s.RegistryOperations.RemoveFinalizer(ctx, registry)
	if err != nil {
		l.Error(err, "Failed to remove finalizer from the registry", "name", registry.Name)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
