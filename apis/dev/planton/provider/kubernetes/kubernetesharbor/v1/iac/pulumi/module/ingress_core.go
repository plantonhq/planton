package module

import (
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createCoreIngress(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider *kubernetes.Provider, namespaceDeps []pulumi.ResourceOption) error {

	// Skip if ingress is not enabled
	if locals.KubernetesHarbor.Spec.Ingress == nil ||
		locals.KubernetesHarbor.Spec.Ingress.Core == nil ||
		!locals.KubernetesHarbor.Spec.Ingress.Core.Enabled ||
		locals.IngressExternalHostname == "" {
		return nil
	}

	certOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	// Create TLS certificate for Harbor Core/Portal
	addedCertificate, err := certmanagerv1.NewCertificate(ctx,
		locals.IngressCertificateName,
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.IngressCertificateName),
				Namespace: pulumi.String(variables.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames:   pulumi.ToStringArray(locals.IngressHostnames),
				SecretName: pulumi.String(locals.IngressCertSecretName),
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					Name: pulumi.String(locals.IngressCertClusterIssuerName),
				},
			},
		}, certOpts...)
	if err != nil {
		return errors.Wrap(err, "error creating certificate")
	}

	// Create Gateway for external Harbor Core/Portal access
	createdGateway, err := gatewayv1.NewGateway(ctx,
		locals.IngressGatewayName,
		&gatewayv1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.IngressGatewayName),
				Namespace: pulumi.String(variables.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: gatewayv1.GatewaySpecArgs{
				GatewayClassName: pulumi.String(variables.GatewayIngressClassName),
				Addresses: gatewayv1.GatewaySpecAddressesArray{
					gatewayv1.GatewaySpecAddressesArgs{
						Type:  pulumi.String("Hostname"),
						Value: pulumi.String(variables.GatewayExternalLoadBalancerServiceHostname),
					},
				},
				Listeners: gatewayv1.GatewaySpecListenersArray{
					&gatewayv1.GatewaySpecListenersArgs{
						Name:     pulumi.String("https-external"),
						Hostname: pulumi.String(locals.IngressExternalHostname),
						Port:     pulumi.Int(443),
						Protocol: pulumi.String("HTTPS"),
						Tls: &gatewayv1.GatewaySpecListenersTlsArgs{
							Mode: pulumi.String("Terminate"),
							CertificateRefs: gatewayv1.GatewaySpecListenersTlsCertificateRefsArray{
								gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs{
									Name: pulumi.String(locals.IngressCertSecretName),
								},
							},
						},
						AllowedRoutes: gatewayv1.GatewaySpecListenersAllowedRoutesArgs{
							Namespaces: gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesArgs{
								From: pulumi.String("All"),
							},
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{
			pulumi.Provider(kubernetesProvider),
			pulumi.DependsOn([]pulumi.Resource{addedCertificate}),
		}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating gateway")
	}

	// Create HTTPRoute for HTTPS traffic to Harbor Core service
	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.IngressHttpRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.IngressHttpRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.IngressExternalHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:        pulumi.String(locals.IngressGatewayName),
						Namespace:   createdGateway.Metadata.Namespace(),
						SectionName: pulumi.String("https-external"),
					},
				},
				Rules: gatewayv1.HTTPRouteSpecRulesArray{
					gatewayv1.HTTPRouteSpecRulesArgs{
						Matches: gatewayv1.HTTPRouteSpecRulesMatchesArray{
							gatewayv1.HTTPRouteSpecRulesMatchesArgs{
								Path: gatewayv1.HTTPRouteSpecRulesMatchesPathArgs{
									Type:  pulumi.String("PathPrefix"),
									Value: pulumi.String("/"),
								},
							},
						},
						BackendRefs: gatewayv1.HTTPRouteSpecRulesBackendRefsArray{
							gatewayv1.HTTPRouteSpecRulesBackendRefsArgs{
								Name:      pulumi.String(locals.CoreServiceName),
								Namespace: pulumi.String(locals.Namespace),
								Port:      pulumi.Int(variables.HarborCorePort),
							},
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)

	return err
}
