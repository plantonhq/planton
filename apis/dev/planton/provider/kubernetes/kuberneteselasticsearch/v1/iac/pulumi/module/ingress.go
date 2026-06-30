package module

import (
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ingress(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider *kubernetes.Provider,
	namespaceDeps []pulumi.ResourceOption) error {
	// Create certificate
	certOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
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

	// Elasticsearch ingress resources
	if locals.ElasticsearchIngressExternalHostname != "" {
		// Create external gateway
		esGwOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{createdCertificate})}, namespaceDeps...)
		elasticsearchCreatedGateway, err := gatewayv1.NewGateway(ctx,
			locals.ElasticsearchExternalGatewayName,
			&gatewayv1.GatewayArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.ElasticsearchExternalGatewayName),
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
							Hostname: pulumi.String(locals.ElasticsearchIngressExternalHostname),
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
							Hostname: pulumi.String(locals.ElasticsearchIngressExternalHostname),
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
			}, esGwOpts...)
		if err != nil {
			return errors.Wrap(err, "error creating gateway")
		}

		//create http-route for setting up https-redirect for external-hostname
		esRedirectOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{elasticsearchCreatedGateway})}, namespaceDeps...)
		_, err = gatewayv1.NewHTTPRoute(ctx,
			locals.ElasticsearchHttpRedirectRouteName,
			&gatewayv1.HTTPRouteArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.ElasticsearchHttpRedirectRouteName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Spec: gatewayv1.HTTPRouteSpecArgs{
					Hostnames: pulumi.StringArray{pulumi.String(locals.ElasticsearchIngressExternalHostname)},
					ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
						gatewayv1.HTTPRouteSpecParentRefsArgs{
							Name:        pulumi.String(locals.ElasticsearchExternalGatewayName),
							Namespace:   elasticsearchCreatedGateway.Metadata.Namespace(),
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
			}, esRedirectOpts...)

		// Create HTTP route for external hostname for https listener
		esHttpsOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		_, err = gatewayv1.NewHTTPRoute(ctx,
			locals.ElasticsearchHttpsRouteName,
			&gatewayv1.HTTPRouteArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.ElasticsearchHttpsRouteName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Spec: gatewayv1.HTTPRouteSpecArgs{
					Hostnames: pulumi.StringArray{pulumi.String(locals.ElasticsearchIngressExternalHostname)},
					ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
						gatewayv1.HTTPRouteSpecParentRefsArgs{
							Name:        pulumi.String(locals.ElasticsearchExternalGatewayName),
							Namespace:   elasticsearchCreatedGateway.Metadata.Namespace(),
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
									Name:      pulumi.String(locals.ElasticsearchKubeServiceName),
									Namespace: pulumi.String(locals.Namespace),
									Port:      pulumi.Int(vars.ElasticsearchPort),
								},
							},
						},
					},
				},
			}, esHttpsOpts...)

		if err != nil {
			return errors.Wrap(err, "error creating HTTP route")
		}
	}

	// Kibana ingress resources
	if locals.KibanaIngressExternalHostname != "" {
		// Create external gateway
		kbGwOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{createdCertificate})}, namespaceDeps...)
		kibanaCreatedGateway, err := gatewayv1.NewGateway(ctx,
			locals.KibanaExternalGatewayName,
			&gatewayv1.GatewayArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.KibanaExternalGatewayName),
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
							Hostname: pulumi.String(locals.KibanaIngressExternalHostname),
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
							Hostname: pulumi.String(locals.KibanaIngressExternalHostname),
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
			}, kbGwOpts...)
		if err != nil {
			return errors.Wrap(err, "error creating gateway")
		}

		//create http-route for setting up https-redirect for external-hostname
		kbRedirectOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		_, err = gatewayv1.NewHTTPRoute(ctx,
			locals.KibanaHttpRedirectRouteName,
			&gatewayv1.HTTPRouteArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.KibanaHttpRedirectRouteName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Spec: gatewayv1.HTTPRouteSpecArgs{
					Hostnames: pulumi.StringArray{pulumi.String(locals.KibanaIngressExternalHostname)},
					ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
						gatewayv1.HTTPRouteSpecParentRefsArgs{
							Name:        pulumi.String(locals.KibanaExternalGatewayName),
							Namespace:   kibanaCreatedGateway.Metadata.Namespace(),
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
			}, kbRedirectOpts...)

		// Create HTTP route for external hostname for https listener
		kbHttpsOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		_, err = gatewayv1.NewHTTPRoute(ctx,
			locals.KibanaHttpsRouteName,
			&gatewayv1.HTTPRouteArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.KibanaHttpsRouteName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Spec: gatewayv1.HTTPRouteSpecArgs{
					Hostnames: pulumi.StringArray{pulumi.String(locals.KibanaIngressExternalHostname)},
					ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
						gatewayv1.HTTPRouteSpecParentRefsArgs{
							Name:        pulumi.String(locals.KibanaExternalGatewayName),
							Namespace:   kibanaCreatedGateway.Metadata.Namespace(),
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
									Name:      pulumi.String(locals.KibanaKubeServiceName),
									Namespace: pulumi.String(locals.Namespace),
									Port:      pulumi.Int(5601),
								},
							},
						},
					},
				},
			}, kbHttpsOpts...)

		if err != nil {
			return errors.Wrap(err, "error creating HTTP route")
		}
	}

	return nil
}
