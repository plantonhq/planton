package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// frontendIngress installs a single external LoadBalancer Service that
// exposes the Temporal gRPC frontend.  HTTP traffic is handled separately in
// frontendHttpIngress.go via Gateway-API / Istio.
func frontendIngress(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	if locals.KubernetesTemporal.Spec.Ingress == nil ||
		locals.KubernetesTemporal.Spec.Ingress.Frontend == nil ||
		!locals.KubernetesTemporal.Spec.Ingress.Frontend.Enabled ||
		locals.KubernetesTemporal.Spec.Ingress.Frontend.GrpcHostname == "" {
		// frontend gRPC ingress disabled – nothing to provision
		return nil
	}

	selector := map[string]string{
		"app.kubernetes.io/instance": locals.KubernetesTemporal.Metadata.Name,
		// Temporal Helm labels its pods with workload-name = “temporal-frontend”
		"app.kubernetes.io/component": "frontend",
	}

	serviceOpts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	serviceOpts = append(serviceOpts, namespaceDeps...)
	_, err := kubernetescorev1.NewService(ctx,
		locals.FrontendGrpcLbServiceName,
		&kubernetescorev1.ServiceArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.FrontendGrpcLbServiceName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
				Annotations: pulumi.StringMap{
					"external-dns.alpha.kubernetes.io/hostname": pulumi.String(locals.IngressFrontendGrpcHostname),
				},
			},
			Spec: &kubernetescorev1.ServiceSpecArgs{
				Type: pulumi.String("LoadBalancer"),
				Ports: kubernetescorev1.ServicePortArray{
					&kubernetescorev1.ServicePortArgs{
						Name:       pulumi.String("grpc-frontend"),
						Port:       pulumi.Int(vars.FrontendGrpcPort),
						Protocol:   pulumi.String("TCP"),
						TargetPort: pulumi.Int(vars.FrontendGrpcPort),
					},
				},
				Selector: pulumi.ToStringMap(selector),
			},
		}, serviceOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create frontend gRPC load balancer service")
	}

	return nil
}
