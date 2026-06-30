package module

import (
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dashboardIngress exposes the Tekton Dashboard via the Gateway API.
// Creates: Certificate → Gateway → HTTPRoute (redirect) → HTTPRoute (HTTPS)
//
// Pattern follows existing Planton components (Solr, Temporal).
func dashboardIngress(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider *kubernetes.Provider, dependsOn []pulumi.Resource) error {

	if !locals.IngressEnabled {
		return nil
	}

	// --------------------- Certificate -------------------------------------
	addedCertificate, err := certmanagerv1.NewCertificate(ctx,
		"dashboard-cert",
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.CertSecretName),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames:   pulumi.ToStringArray([]string{locals.IngressHostname}),
				SecretName: pulumi.String(locals.CertSecretName),
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					Name: pulumi.String(locals.ClusterIssuerName),
				},
			},
		}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn(dependsOn))
	if err != nil {
		return errors.Wrap(err, "error creating dashboard certificate")
	}

	// --------------------- Gateway -----------------------------------------
	gwName := pulumi.String(locals.GatewayName)
	createdGateway, err := gatewayv1.NewGateway(ctx,
		locals.GatewayName,
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
						Hostname: pulumi.String(locals.IngressHostname),
						Port:     pulumi.Int(443),
						Protocol: pulumi.String("HTTPS"),
						Tls: &gatewayv1.GatewaySpecListenersTlsArgs{
							Mode: pulumi.String("Terminate"),
							CertificateRefs: gatewayv1.GatewaySpecListenersTlsCertificateRefsArray{
								gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs{
									Name: pulumi.String(locals.CertSecretName),
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
						Hostname: pulumi.String(locals.IngressHostname),
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
		}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{addedCertificate}))
	if err != nil {
		return errors.Wrap(err, "error creating dashboard gateway")
	}

	// ----------------- HTTPRoute (HTTP → HTTPS redirect) --------------------
	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.HttpRedirectRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.HttpRedirectRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.IngressHostname)},
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
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "error creating http→https redirect route")
	}

	// ----------------- HTTPRoute (HTTPS → Dashboard backend) ----------------
	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.HttpsRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.HttpsRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.IngressHostname)},
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
								Name:      pulumi.String(vars.DashboardServiceName),
								Namespace: pulumi.String(locals.Namespace),
								Port:      pulumi.Int(vars.DashboardServicePort),
							},
						},
					},
				},
			},
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "error creating HTTPS route for dashboard")
	}

	return nil
}
