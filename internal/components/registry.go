package components

import (
	"context"
	"slices"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	registryoperatordevv1alpha1 "github.com/registry-operator/registry-operator/api/v1alpha1"
	"github.com/registry-operator/registry-operator/internal"
	"github.com/registry-operator/registry-operator/internal/components/factories"
)

type RegistryOperations struct {
	Client           client.Client
	PodFactory       *factories.PodFactory
	ConfigMapFactory *factories.ConfigMapFactory
}

func NewRegistryOperations(client client.Client) *RegistryOperations {
	return &RegistryOperations{Client: client}
}

func (ro *RegistryOperations) CheckRegistryPodExists(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) (bool, error) {
	l := log.FromContext(ctx)
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registry.Name,
			Namespace: registry.Namespace,
		},
	}
	l.Info("Checking if pod exists for", "registry", registry.Name)
	err := ro.Client.Get(ctx, client.ObjectKeyFromObject(pod), pod)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (ro *RegistryOperations) GetRegistryPod(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) (*apiv1.Pod, error) {
	l := log.FromContext(ctx)
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registry.Name,
			Namespace: registry.Namespace,
		},
	}
	l.Info("Getting pod for", "registry", registry.Name)
	err := ro.Client.Get(ctx, client.ObjectKeyFromObject(pod), pod)
	return pod, err
}

func (ro *RegistryOperations) CreateRegistryPod(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) error {
	l := log.FromContext(ctx)
	pod, err := ro.PodFactory.NewPod(registry)
	if err != nil {
		return err
	}
	l.Info("Creating pod for", "registry", registry.Name)
	return ro.Client.Create(ctx, pod)
}

func (ro *RegistryOperations) UpdateRegistryStatus(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) error {
	l := log.FromContext(ctx)
	l.Info("Updating status for", "registry", registry.Name)
	return ro.Client.Status().Update(ctx, registry)
}

func (ro *RegistryOperations) DeleteRegistryPod(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) error {
	l := log.FromContext(ctx)
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registry.Name,
			Namespace: registry.Namespace,
		},
	}
	l.Info("Deleting pod for", "registry", registry.Name)
	return ro.Client.Delete(ctx, pod)
}

func (ro *RegistryOperations) AddFinalizer(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) error {
	l := log.FromContext(ctx)
	l.Info("Adding finalizer to", "registry", registry.Name)
	registry.SetFinalizers(append(registry.GetFinalizers(), internal.RegistryFinalizer))
	return ro.Client.Update(ctx, registry)
}

func (ro *RegistryOperations) RemoveFinalizer(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) error {
	l := log.FromContext(ctx)
	l.Info("Removing finalizer from", "registry", registry.Name)
	finalizers := registry.GetFinalizers()
	idx := slices.Index(finalizers, internal.RegistryFinalizer)
	if idx == -1 {
		return nil
	}
	registry.SetFinalizers(append(finalizers[:idx], finalizers[idx+1:]...))
	return ro.Client.Update(ctx, registry)
}

func (ro *RegistryOperations) CheckFinalizerExists(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) (bool, error) {
	l := log.FromContext(ctx)
	l.Info("Checking finalizer for", "registry", registry.Name)
	finalizers := registry.GetFinalizers()
	idx := slices.Index(finalizers, internal.RegistryFinalizer)
	if idx == -1 {
		return false, nil
	}
	return true, nil
}

func (ro *RegistryOperations) CheckRegistryConfigMapExists(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) (bool, error) {
	l := log.FromContext(ctx)
	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registry.Name,
			Namespace: registry.Namespace,
		},
	}
	l.Info("Checking if ConfigMap exists for", "registry", registry.Name)
	err := ro.Client.Get(ctx, client.ObjectKeyFromObject(configMap), configMap)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (ro *RegistryOperations) CreateRegistryConfigMap(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) error {
	l := log.FromContext(ctx)
	l.Info("Creating ConfigMap for", "registry", registry.Name)
	configMap := ro.ConfigMapFactory.NewConfigMap(registry)
	return ro.Client.Create(ctx, configMap)
}

func (ro *RegistryOperations) DeleteRegistryConfigMap(ctx context.Context, registry *registryoperatordevv1alpha1.Registry) error {
	l := log.FromContext(ctx)
	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registry.Name,
			Namespace: registry.Namespace,
		},
	}
	l.Info("Deleting ConfigMap for", "registry", registry.Name)
	return ro.Client.Delete(ctx, configMap)
}
