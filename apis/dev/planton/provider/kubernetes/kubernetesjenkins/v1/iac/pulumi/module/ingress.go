package module

import (
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ingress(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	namespaceDeps []pulumi.ResourceOption) error {
	// Create certificate
	certOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
	}, namespaceDeps...)

	createdCertificate, err := certmanagerv1.NewCertificate(ctx,
		locals.IngressCertificateName,
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.IngressCertificateName),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
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

	// Create external gateway
	createdGateway, err := gatewayv1.NewGateway(ctx,
		locals.ExternalGatewayName,
		&gatewayv1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.ExternalGatewayName),
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
		}, append([]pulumi.ResourceOption{
			pulumi.Provider(kubernetesProvider),
			pulumi.DependsOn([]pulumi.Resource{createdCertificate}),
		}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "error creating gateway")
	}

	//create http-route for setting up https-redirect for external-hostname
	redirectOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
	}, namespaceDeps...)

	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.HttpRedirectRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.HttpRedirectRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.IngressExternalHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:        pulumi.String(locals.ExternalGatewayName),
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
		}, redirectOpts...)

	// Create HTTP route for external hostname for https listener
	httpsOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
	}, namespaceDeps...)

	_, err = gatewayv1.NewHTTPRoute(ctx,
		locals.HttpsRouteName,
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.HttpsRouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.IngressExternalHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:        pulumi.String(locals.ExternalGatewayName),
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
								Name:      pulumi.String(locals.KubeServiceName),
								Namespace: pulumi.String(locals.Namespace),
								Port:      pulumi.Int(80),
							},
						},
					},
				},
			},
		}, httpsOpts...)

	if err != nil {
		return errors.Wrap(err, "error creating HTTP route")
	}

	return nil
}
