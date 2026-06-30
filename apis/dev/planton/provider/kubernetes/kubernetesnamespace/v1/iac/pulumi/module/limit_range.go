package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createLimitRange creates a LimitRange in the namespace if enabled
func createLimitRange(ctx *pulumi.Context, locals *Locals, namespace *kubernetescorev1.Namespace, provider pulumi.ProviderResource) error {
	if !locals.LimitRange.Enabled {
		return nil
	}

	limitRangeName := fmt.Sprintf("%s-limits", locals.NamespaceName)

	defaultRequest := make(map[string]string)
	defaultLimit := make(map[string]string)

	if locals.LimitRange.DefaultCpuRequest != "" {
		defaultRequest["cpu"] = locals.LimitRange.DefaultCpuRequest
	}
	if locals.LimitRange.DefaultMemoryRequest != "" {
		defaultRequest["memory"] = locals.LimitRange.DefaultMemoryRequest
	}
	if locals.LimitRange.DefaultCpuLimit != "" {
		defaultLimit["cpu"] = locals.LimitRange.DefaultCpuLimit
	}
	if locals.LimitRange.DefaultMemoryLimit != "" {
		defaultLimit["memory"] = locals.LimitRange.DefaultMemoryLimit
	}

	_, err := kubernetescorev1.NewLimitRange(
		ctx,
		limitRangeName,
		&kubernetescorev1.LimitRangeArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(limitRangeName),
				Namespace: namespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: &kubernetescorev1.LimitRangeSpecArgs{
				Limits: kubernetescorev1.LimitRangeItemArray{
					&kubernetescorev1.LimitRangeItemArgs{
						Type:           pulumi.String("Container"),
						DefaultRequest: pulumi.ToStringMap(defaultRequest),
						Default:        pulumi.ToStringMap(defaultLimit),
					},
				},
			},
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{namespace}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create limit range in namespace %s", locals.NamespaceName)
	}

	return nil
}
