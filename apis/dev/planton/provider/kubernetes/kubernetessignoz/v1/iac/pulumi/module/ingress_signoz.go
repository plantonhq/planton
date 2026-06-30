package module

import (
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createSignozUIIngress creates Kubernetes Gateway API resources for SigNoz UI external access
// This includes Certificate, Gateway, and HTTPRoute resources following the Gateway API standard
func createSignozUIIngress(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	// Skip if ingress is not enabled
	if locals.KubernetesSignoz.Spec.Ingress == nil ||
		locals.KubernetesSignoz.Spec.Ingress.Ui == nil ||
		!locals.KubernetesSignoz.Spec.Ingress.Ui.Enabled ||
		locals.KubernetesSignoz.Spec.Ingress.Ui.Hostname == "" {
		return nil
	}

	// Create TLS certificate for the SigNoz UI
	// Uses computed name to avoid conflicts when multiple instances share a namespace
	addedCertificate, err := certmanagerv1.NewCertificate(ctx,
		locals.SignozCertificateName,
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.SignozCertificateName),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
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
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating certificate")
	}

	// Create Gateway for external SigNoz UI access
	// Uses computed name to avoid conflicts when multiple instances share a namespace
	createdGateway, err := gatewayv1.NewGateway(ctx,
		locals.SignozGatewayName,
		&gatewayv1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.SignozGatewayName),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: gatewayv1.GatewaySpecArgs{
				GatewayClassName: pulumi.String(vars.GatewayIngressClassName),
				Addresses: gatewayv1.GatewaySpecAddressesArray{
					gatewayv1.GatewaySpecAddressesArgs{
						Type:  pulumi.String("Hostname"),
						Value: pulumi.String(vars.GatewayExternalLoadBalancerServiceHostname),
					},
				},
				Listeners: gatewayv1.GatewaySpecListenersArray{
					// HTTPS listener for secure SigNoz UI access
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
					// HTTP listener for HTTP-to-HTTPS redirect
					&gatewayv1.GatewaySpecListenersArgs{
						Name:     pulumi.String("http-external"),
						Hostname: pulumi.String(locals.IngressExternalHostname),
						Port:     pulumi.Int(80),
						Protocol: pulumi.String("HTTP"),
						AllowedRoutes: gatewayv1.GatewaySpecListenersAllowedRoutesArgs{
							Namespaces: gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesArgs{
								From: pulumi.String("All"),
							},
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{addedCertificate})}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating gateway")
	}

	// Create HTTPRoute for HTTP-to-HTTPS redirect
	// Uses computed name to avoid conflicts when multiple instances share a namespace
	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.SignozHttpRedirectRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.SignozHttpRedirectRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.IngressExternalHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:        pulumi.String(locals.SignozGatewayName),
						Namespace:   createdGateway.Metadata.Namespace(),
						SectionName: pulumi.String("http-external"),
					},
				},
				Rules: gatewayv1.HTTPRouteSpecRulesArray{
					gatewayv1.HTTPRouteSpecRulesArgs{
						Filters: gatewayv1.HTTPRouteSpecRulesFiltersArray{
							gatewayv1.HTTPRouteSpecRulesFiltersArgs{
								RequestRedirect: gatewayv1.HTTPRouteSpecRulesFiltersRequestRedirectArgs{
									Scheme:     pulumi.String("https"),
									StatusCode: pulumi.Int(301),
								},
								Type: pulumi.String("RequestRedirect"),
							},
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP redirect route for SigNoz UI")
	}

	// Create HTTPRoute for HTTPS traffic to SigNoz frontend service
	// Uses computed name to avoid conflicts when multiple instances share a namespace
	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.SignozHTTPRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.SignozHTTPRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.IngressExternalHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:        pulumi.String(locals.SignozGatewayName),
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
								// Route to the actual frontend service created by SigNoz Helm chart
								Name:      pulumi.Sprintf("%s-frontend", locals.KubernetesSignoz.Metadata.Name),
								Namespace: pulumi.String(locals.Namespace),
								Port:      pulumi.Int(vars.SignozFrontendPort),
							},
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating HTTPS route for SigNoz UI")
	}

	return nil
}
