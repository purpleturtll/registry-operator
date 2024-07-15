package factories

import (
	"fmt"

	registryoperatordevv1alpha1 "github.com/registry-operator/registry-operator/api/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type PodFactory struct{}

func NewPodFactory() *PodFactory {
	return &PodFactory{}
}

// CreatePod creates a Kubernetes Pod based on the registry specification.
func (f *PodFactory) NewPod(registry *registryoperatordevv1alpha1.Registry) (*apiv1.Pod, error) {
	switch registry.Spec.Storage.Type {
	case registryoperatordevv1alpha1.StorageTypeInMemory:
		return f.createInMemoryPod(registry), nil
	default:
		return nil, fmt.Errorf("storage type %s not supported", registry.Spec.Storage.Type)
	}
}

// createInMemoryPod generates a pod configuration for in-memory storage.
func (f *PodFactory) createInMemoryPod(registry *registryoperatordevv1alpha1.Registry) *apiv1.Pod {
	return &apiv1.Pod{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      registry.Name,
			Namespace: registry.Namespace,
			Labels: map[string]string{
				"app":      "registry",
				"registry": registry.Name,
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  registry.Name,
					Image: "registry:2",
					VolumeMounts: []apiv1.VolumeMount{
						{
							Name:      "config",
							MountPath: "/etc/distribution",
						},
					},
				},
			},
			Volumes: []apiv1.Volume{
				{
					Name: "config",
					VolumeSource: apiv1.VolumeSource{
						ConfigMap: &apiv1.ConfigMapVolumeSource{
							LocalObjectReference: apiv1.LocalObjectReference{
								Name: registry.Name,
							},
						},
					},
				},
			},
		},
	}
}
