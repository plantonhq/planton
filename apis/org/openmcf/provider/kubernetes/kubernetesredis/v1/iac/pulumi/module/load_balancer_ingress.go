package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// loadBalancerIngress creates an external LoadBalancer Service when spec.ingress.enabled is true.
func loadBalancerIngress(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := kubernetescorev1.NewService(ctx,
		locals.ExternalLbServiceName,
		&kubernetescorev1.ServiceArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.ExternalLbServiceName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
				Annotations: pulumi.StringMap{
					"external-dns.alpha.kubernetes.io/hostname": pulumi.String(locals.IngressExternalHostname),
				},
			},
			Spec: &kubernetescorev1.ServiceSpecArgs{
				Type: pulumi.String("LoadBalancer"), // Service type is LoadBalancer
				Ports: kubernetescorev1.ServicePortArray{
					&kubernetescorev1.ServicePortArgs{
						Name:     pulumi.String("tcp-redis"),
						Port:     pulumi.Int(vars.RedisPort),
						Protocol: pulumi.String("TCP"),
						// This assumes your Redis pod has a port named 'redis'
						TargetPort: pulumi.String("redis"),
					},
				},
				Selector: pulumi.ToStringMap(locals.RedisPodSelectorLabels),
			},
		}, opts...)
	if err != nil {
		return errors.Wrapf(err, "failed to create external load balancer service")
	}
	return nil
}
