package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createResourceQuota creates a ResourceQuota in the namespace if enabled
func createResourceQuota(ctx *pulumi.Context, locals *Locals, namespace *kubernetescorev1.Namespace, provider pulumi.ProviderResource) error {
	if !locals.ResourceQuota.Enabled {
		return nil
	}

	quotaName := fmt.Sprintf("%s-quota", locals.NamespaceName)
	hard := make(map[string]string)

	// CPU and Memory quotas
	if locals.ResourceQuota.CpuRequests != "" {
		hard["requests.cpu"] = locals.ResourceQuota.CpuRequests
	}
	if locals.ResourceQuota.CpuLimits != "" {
		hard["limits.cpu"] = locals.ResourceQuota.CpuLimits
	}
	if locals.ResourceQuota.MemoryRequests != "" {
		hard["requests.memory"] = locals.ResourceQuota.MemoryRequests
	}
	if locals.ResourceQuota.MemoryLimits != "" {
		hard["limits.memory"] = locals.ResourceQuota.MemoryLimits
	}

	// Object count quotas
	if locals.ResourceQuota.Pods > 0 {
		hard["count/pods"] = fmt.Sprintf("%d", locals.ResourceQuota.Pods)
	}
	if locals.ResourceQuota.Services > 0 {
		hard["count/services"] = fmt.Sprintf("%d", locals.ResourceQuota.Services)
	}
	if locals.ResourceQuota.ConfigMaps > 0 {
		hard["count/configmaps"] = fmt.Sprintf("%d", locals.ResourceQuota.ConfigMaps)
	}
	if locals.ResourceQuota.Secrets > 0 {
		hard["count/secrets"] = fmt.Sprintf("%d", locals.ResourceQuota.Secrets)
	}
	if locals.ResourceQuota.PVCs > 0 {
		hard["count/persistentvolumeclaims"] = fmt.Sprintf("%d", locals.ResourceQuota.PVCs)
	}
	if locals.ResourceQuota.LoadBalancers > 0 {
		hard["count/services.loadbalancers"] = fmt.Sprintf("%d", locals.ResourceQuota.LoadBalancers)
	}

	_, err := kubernetescorev1.NewResourceQuota(
		ctx,
		quotaName,
		&kubernetescorev1.ResourceQuotaArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(quotaName),
				Namespace: namespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: &kubernetescorev1.ResourceQuotaSpecArgs{
				Hard: pulumi.ToStringMap(hard),
			},
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{namespace}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create resource quota in namespace %s", locals.NamespaceName)
	}

	return nil
}
