package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createIngressLoadBalancer creates a LoadBalancer Service for external access to ClickHouse
// This is optional and only created when ingress is enabled in the spec
//
// The LoadBalancer exposes both HTTP (8123) and native protocol (9000) ports
// and includes external-dns annotation for automatic DNS configuration
func createIngressLoadBalancer(
	ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	namespaceDeps []pulumi.ResourceOption,
) error {
	// Skip if ingress is not enabled
	if locals.KubernetesClickHouse.Spec.Ingress == nil || !locals.KubernetesClickHouse.Spec.Ingress.Enabled {
		return nil
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	// Create LoadBalancer service with external DNS annotation
	// Use computed name to avoid conflicts when multiple instances share a namespace
	_, err := kubernetescorev1.NewService(ctx,
		locals.ExternalLbServiceName,
		&kubernetescorev1.ServiceArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.ExternalLbServiceName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
				Annotations: pulumi.StringMap{
					// External DNS annotation for automatic DNS record creation
					"external-dns.alpha.kubernetes.io/hostname": pulumi.String(locals.IngressExternalHostname),
				},
			},
			Spec: &kubernetescorev1.ServiceSpecArgs{
				Type: pulumi.String("LoadBalancer"),
				Ports: kubernetescorev1.ServicePortArray{
					// HTTP port for ClickHouse HTTP interface
					&kubernetescorev1.ServicePortArgs{
						Name:       pulumi.String("http"),
						Port:       pulumi.Int(vars.ClickhouseHttpPort),
						Protocol:   pulumi.String("TCP"),
						TargetPort: pulumi.String("http"),
					},
					// Native protocol port for ClickHouse client connections
					&kubernetescorev1.ServicePortArgs{
						Name:       pulumi.String("tcp"),
						Port:       pulumi.Int(vars.ClickhouseNativePort),
						Protocol:   pulumi.String("TCP"),
						TargetPort: pulumi.String("tcp"),
					},
				},
				// Selector targets ClickHouse pods managed by Altinity operator
				Selector: pulumi.ToStringMap(locals.ClickhousePodSelectorLabels),
			},
		}, opts...)

	if err != nil {
		return errors.Wrapf(err, "failed to create external load balancer service")
	}

	return nil
}
