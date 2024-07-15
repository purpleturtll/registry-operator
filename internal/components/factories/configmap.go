package factories

import (
	"fmt"

	registryoperatordevv1alpha1 "github.com/registry-operator/registry-operator/api/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ConfigMapFactory struct{}

func NewConfigMapFactory() *ConfigMapFactory {
	return &ConfigMapFactory{}
}

// CreatePod creates a Kubernetes Pod based on the registry specification.
func (f *ConfigMapFactory) NewConfigMap(registry *registryoperatordevv1alpha1.Registry) *apiv1.ConfigMap {
	return &apiv1.ConfigMap{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      registry.Name,
			Namespace: registry.Namespace,
			Labels: map[string]string{
				"app":      "registry",
				"registry": registry.Name,
			},
		},
		// We could use Configuration struct from registry repo in the future.
		Data: map[string]string{
			"config.yml": fmt.Sprintf("version: 0.1\nstorage:\n\t%s:\n", registry.Spec.Storage.Type),
		},
	}
}
