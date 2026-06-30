package module

import (
	"strings"

	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// frontendHttpIngress exposes the Temporal Frontend HTTP API via the shared Istio / Gateway-API
// ingress stack (no separate external LB Service).  Pattern copied from webUiIngress:
// Certificate ➜ Gateway ➜ HTTPS+redirect HTTPRoutes.
func frontendHttpIngress(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider *kubernetes.Provider, namespaceDeps []pulumi.ResourceOption) error {

	if locals.KubernetesTemporal.Spec.Ingress == nil ||
		locals.KubernetesTemporal.Spec.Ingress.Frontend == nil ||
		!locals.KubernetesTemporal.Spec.Ingress.Frontend.Enabled ||
		locals.KubernetesTemporal.Spec.Ingress.Frontend.HttpHostname == "" {
		// No frontend HTTP ingress required.
		return nil
	}

	// Hostname + cert secret name
	httpHostname := locals.IngressFrontendHttpHostname // User-specified hostname
	certSecret := locals.FrontendHttpCertSecretName    // Computed from metadata.name

	// Extract domain from hostname for ClusterIssuer name
	hostnameParts := strings.Split(httpHostname, ".")
	var clusterIssuerName string
	if len(hostnameParts) > 1 {
		clusterIssuerName = strings.Join(hostnameParts[1:], ".")
	}

	// --------------------- Certificate -------------------------------------
	addedCertificate, err := certmanagerv1.NewCertificate(ctx,
		"frontend-http-cert",
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(certSecret),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames:   pulumi.ToStringArray([]string{httpHostname}),
				SecretName: pulumi.String(certSecret),
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					Name: pulumi.String(clusterIssuerName),
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating frontend HTTP certificate")
	}

	// --------------------- Gateway -----------------------------------------
	gwName := pulumi.String(locals.FrontendHttpGatewayName)
	createdGateway, err := gatewayv1.NewGateway(ctx,
		locals.FrontendHttpGatewayName,
		&gatewayv1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      gwName,
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
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
					&gatewayv1.GatewaySpecListenersArgs{
						Name:     pulumi.String("https-external"),
						Hostname: pulumi.String(httpHostname),
						Port:     pulumi.Int(443),
						Protocol: pulumi.String("HTTPS"),
						Tls: &gatewayv1.GatewaySpecListenersTlsArgs{
							Mode: pulumi.String("Terminate"),
							CertificateRefs: gatewayv1.GatewaySpecListenersTlsCertificateRefsArray{
								gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs{
									Name: pulumi.String(certSecret),
								},
							},
						},
						AllowedRoutes: gatewayv1.GatewaySpecListenersAllowedRoutesArgs{
							Namespaces: gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesArgs{
								From: pulumi.String("All"),
							},
						},
					},
					&gatewayv1.GatewaySpecListenersArgs{
						Name:     pulumi.String("http-external"),
						Hostname: pulumi.String(httpHostname),
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
		return errors.Wrap(err, "error creating frontend HTTP gateway")
	}

	// ----------------- HTTPRoute (redirect) --------------------------------
	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.FrontendHttpRedirectRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.FrontendHttpRedirectRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(httpHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:        gwName,
						Namespace:   createdGateway.Metadata.Namespace(),
						SectionName: pulumi.String("http-external"),
					},
				},
				Rules: gatewayv1.HTTPRouteSpecRulesArray{
					gatewayv1.HTTPRouteSpecRulesArgs{
						Filters: gatewayv1.HTTPRouteSpecRulesFiltersArray{
							gatewayv1.HTTPRouteSpecRulesFiltersArgs{
								Type: pulumi.String("RequestRedirect"),
								RequestRedirect: gatewayv1.HTTPRouteSpecRulesFiltersRequestRedirectArgs{
									Scheme:     pulumi.String("https"),
									StatusCode: pulumi.Int(301),
								},
							},
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating http→https redirect route for frontend HTTP")
	}

	// ----------------- HTTPRoute (HTTPS) -----------------------------------
	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.FrontendHttpsRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.FrontendHttpsRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(httpHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:        gwName,
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
								Name:      pulumi.String(locals.FrontendServiceName),
								Namespace: pulumi.String(locals.Namespace),
								Port:      pulumi.Int(vars.FrontendHttpPort),
							},
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating HTTPS route for frontend HTTP")
	}

	return nil
}
