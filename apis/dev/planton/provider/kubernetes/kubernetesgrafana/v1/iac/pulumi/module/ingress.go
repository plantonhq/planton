package module

import (
	"github.com/pkg/errors"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	kubernetesnetworkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ingress(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	namespaceDeps []pulumi.ResourceOption) error {

	if locals.KubernetesGrafana.Spec.Ingress == nil ||
		!locals.KubernetesGrafana.Spec.Ingress.Enabled {
		return nil
	}

	// Extract external hostname without https:// prefix
	externalHost := ""
	if locals.IngressExternalHostname != "" {
		// Remove https:// prefix if present
		externalHost = locals.KubernetesGrafana.Spec.Ingress.Hostname
	}

	// Extract internal hostname without https:// prefix
	internalHost := ""
	if locals.IngressInternalHostname != "" {
		internalHost = "internal-" + locals.KubernetesGrafana.Spec.Ingress.Hostname
	}

	pathType := "Prefix"

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	// Create external ingress if hostname is provided
	if externalHost != "" {
		_, err := kubernetesnetworkingv1.NewIngress(ctx,
			locals.ExternalIngressName,
			&kubernetesnetworkingv1.IngressArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.ExternalIngressName),
					Namespace: pulumi.String(locals.Namespace),
					Annotations: pulumi.StringMap{
						"kubernetes.io/ingress.class": pulumi.String("nginx"),
					},
				},
				Spec: &kubernetesnetworkingv1.IngressSpecArgs{
					Rules: kubernetesnetworkingv1.IngressRuleArray{
						&kubernetesnetworkingv1.IngressRuleArgs{
							Host: pulumi.String(externalHost),
							Http: &kubernetesnetworkingv1.HTTPIngressRuleValueArgs{
								Paths: kubernetesnetworkingv1.HTTPIngressPathArray{
									&kubernetesnetworkingv1.HTTPIngressPathArgs{
										Path:     pulumi.String("/"),
										PathType: pulumi.String(pathType),
										Backend: &kubernetesnetworkingv1.IngressBackendArgs{
											Service: &kubernetesnetworkingv1.IngressServiceBackendArgs{
												Name: pulumi.String(locals.KubeServiceName),
												Port: &kubernetesnetworkingv1.ServiceBackendPortArgs{
													Number: pulumi.Int(80),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}, opts...)
		if err != nil {
			return errors.Wrap(err, "failed to create external ingress")
		}
	}

	// Create internal ingress if hostname is provided
	if internalHost != "" {
		_, err := kubernetesnetworkingv1.NewIngress(ctx,
			locals.InternalIngressName,
			&kubernetesnetworkingv1.IngressArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.InternalIngressName),
					Namespace: pulumi.String(locals.Namespace),
					Annotations: pulumi.StringMap{
						"kubernetes.io/ingress.class": pulumi.String("nginx-internal"),
					},
				},
				Spec: &kubernetesnetworkingv1.IngressSpecArgs{
					Rules: kubernetesnetworkingv1.IngressRuleArray{
						&kubernetesnetworkingv1.IngressRuleArgs{
							Host: pulumi.String(internalHost),
							Http: &kubernetesnetworkingv1.HTTPIngressRuleValueArgs{
								Paths: kubernetesnetworkingv1.HTTPIngressPathArray{
									&kubernetesnetworkingv1.HTTPIngressPathArgs{
										Path:     pulumi.String("/"),
										PathType: pulumi.String(pathType),
										Backend: &kubernetesnetworkingv1.IngressBackendArgs{
											Service: &kubernetesnetworkingv1.IngressServiceBackendArgs{
												Name: pulumi.String(locals.KubeServiceName),
												Port: &kubernetesnetworkingv1.ServiceBackendPortArgs{
													Number: pulumi.Int(80),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}, opts...)
		if err != nil {
			return errors.Wrap(err, "failed to create internal ingress")
		}
	}

	return nil
}
