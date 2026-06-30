package module

import (
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createOtelCollectorIngress creates Kubernetes Gateway API resources for OpenTelemetry Collector external access
// This includes Certificate, Gateway, and HTTPRoute resources for HTTP endpoint
func createOtelCollectorIngress(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	// Skip if ingress is not enabled
	if locals.KubernetesSignoz.Spec.Ingress == nil ||
		locals.KubernetesSignoz.Spec.Ingress.OtelCollector == nil ||
		!locals.KubernetesSignoz.Spec.Ingress.OtelCollector.Enabled ||
		locals.KubernetesSignoz.Spec.Ingress.OtelCollector.Hostname == "" {
		return nil
	}

	// Certificate for OTEL Collector HTTP endpoint
	// Uses computed name to avoid conflicts when multiple instances share a namespace
	otelHostnames := []string{
		locals.OtelCollectorExternalHttpHostname,
	}

	addedOtelCertificate, err := certmanagerv1.NewCertificate(ctx,
		locals.OtelCertificateName,
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.OtelCertificateName),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames:   pulumi.ToStringArray(otelHostnames),
				SecretName: pulumi.String(locals.OtelCertificateName),
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					Name: pulumi.String(locals.IngressCertClusterIssuerName),
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating OTEL Collector certificate")
	}

	// Gateway for OTEL Collector HTTP endpoint
	// Uses computed name to avoid conflicts when multiple instances share a namespace
	createdOtelGateway, err := gatewayv1.NewGateway(ctx,
		locals.OtelGatewayName,
		&gatewayv1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.OtelGatewayName),
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
					// HTTPS listener for HTTP endpoint (port 443)
					&gatewayv1.GatewaySpecListenersArgs{
						Name:     pulumi.String("https-otel-http"),
						Hostname: pulumi.String(locals.OtelCollectorExternalHttpHostname),
						Port:     pulumi.Int(443),
						Protocol: pulumi.String("HTTPS"),
						Tls: &gatewayv1.GatewaySpecListenersTlsArgs{
							Mode: pulumi.String("Terminate"),
							CertificateRefs: gatewayv1.GatewaySpecListenersTlsCertificateRefsArray{
								gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs{
									Name: pulumi.String(locals.OtelCertificateName),
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
						Name:     pulumi.String("http-otel-http"),
						Hostname: pulumi.String(locals.OtelCollectorExternalHttpHostname),
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
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{addedOtelCertificate})}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating OTEL Collector HTTP gateway")
	}

	// HTTPRoute for HTTP-to-HTTPS redirect
	// Uses computed name to avoid conflicts when multiple instances share a namespace
	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.OtelHttpRedirectRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.OtelHttpRedirectRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.OtelCollectorExternalHttpHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:        pulumi.String(locals.OtelGatewayName),
						Namespace:   createdOtelGateway.Metadata.Namespace(),
						SectionName: pulumi.String("http-otel-http"),
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
		return errors.Wrap(err, "error creating HTTP redirect route for OTEL Collector")
	}

	// HTTPRoute for HTTP endpoint (routes to OTEL Collector HTTP port 4318)
	// Uses computed name to avoid conflicts when multiple instances share a namespace
	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.OtelHTTPRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.OtelHTTPRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.KubernetesLabels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.OtelCollectorExternalHttpHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:        pulumi.String(locals.OtelGatewayName),
						Namespace:   createdOtelGateway.Metadata.Namespace(),
						SectionName: pulumi.String("https-otel-http"),
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
								// Route to OTEL Collector HTTP port (4318)
								Name:      pulumi.Sprintf("%s-otel-collector", locals.KubernetesSignoz.Metadata.Name),
								Namespace: pulumi.String(locals.Namespace),
								Port:      pulumi.Int(vars.OtelHttpPort),
							},
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating HTTPRoute for OTEL Collector HTTP endpoint")
	}

	return nil
}
