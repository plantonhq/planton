package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// configMaps creates ConfigMap resources from the spec.config_maps map.
// Returns a map of ConfigMap name to the created ConfigMap resource.
func configMaps(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) (map[string]*kubernetescorev1.ConfigMap, error) {
	result := make(map[string]*kubernetescorev1.ConfigMap)

	if locals.KubernetesCronJob.Spec.ConfigMaps == nil {
		return result, nil
	}

	for name, content := range locals.KubernetesCronJob.Spec.ConfigMaps {
		opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		cm, err := kubernetescorev1.NewConfigMap(ctx,
			name,
			&kubernetescorev1.ConfigMapArgs{
				Metadata: &kubernetesmetav1.ObjectMetaArgs{
					Name:      pulumi.String(name),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Data: pulumi.StringMap{
					name: pulumi.String(content),
				},
			},
			opts...,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create configmap %s", name)
		}
		result[name] = cm
	}

	return result, nil
}
