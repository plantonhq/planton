package module

import (
	"fmt"

	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/strimzioperator/kubernetes/kafka/v1beta2"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	k8scorev1 "k8s.io/api/core/v1"
)

func schemaRegistry(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	createdKafkaCluster *v1beta2.Kafka, namespaceDeps []pulumi.ResourceOption) error {

	labels := locals.Labels
	labels["app"] = locals.SchemaRegistryDeploymentName

	//create schema-registry deployment
	deployOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := appsv1.NewDeployment(ctx,
		locals.SchemaRegistryDeploymentName,
		&appsv1.DeploymentArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.SchemaRegistryDeploymentName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: &appsv1.DeploymentSpecArgs{
				Replicas: pulumi.Int(1),
				Selector: &metav1.LabelSelectorArgs{
					MatchLabels: pulumi.StringMap{
						"app": pulumi.String(locals.SchemaRegistryDeploymentName),
					},
				},
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: pulumi.StringMap{
							"app": pulumi.String(locals.SchemaRegistryDeploymentName),
						},
					},
					Spec: &corev1.PodSpecArgs{
						//InitContainers: corev1.ContainerArray{
						//	common.GetKafkaReadyCheckContainer(),
						//},
						Containers: corev1.ContainerArray{
							&corev1.ContainerArgs{
								Name:  pulumi.String(locals.SchemaRegistryDeploymentName),
								Image: pulumi.String(vars.SchemaRegistryDockerImage),
								Ports: corev1.ContainerPortArray{
									&corev1.ContainerPortArgs{
										Name:          pulumi.String("http"),
										ContainerPort: pulumi.Int(vars.SchemaRegistryContainerPort),
									},
								},
								Resources: corev1.ResourceRequirementsArgs{
									Limits: pulumi.ToStringMap(map[string]string{
										string(k8scorev1.ResourceCPU):    locals.KubernetesKafka.Spec.SchemaRegistryContainer.Resources.Limits.Cpu,
										string(k8scorev1.ResourceMemory): locals.KubernetesKafka.Spec.SchemaRegistryContainer.Resources.Limits.Memory,
									}),
									Requests: pulumi.ToStringMap(map[string]string{
										string(k8scorev1.ResourceCPU):    locals.KubernetesKafka.Spec.SchemaRegistryContainer.Resources.Requests.Cpu,
										string(k8scorev1.ResourceMemory): locals.KubernetesKafka.Spec.SchemaRegistryContainer.Resources.Requests.Memory,
									}),
								},
								Env: corev1.EnvVarArray{
									corev1.EnvVarInput(corev1.EnvVarArgs{
										Name: pulumi.String("SCHEMA_REGISTRY_HOST_NAME"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											FieldRef: &corev1.ObjectFieldSelectorArgs{
												FieldPath: pulumi.String("status.podIP"),
											},
										},
									}),
									corev1.EnvVarInput(corev1.EnvVarArgs{
										Name:  pulumi.String("SCHEMA_REGISTRY_LISTENERS"),
										Value: pulumi.String("http://0.0.0.0:8081"),
									}),
									corev1.EnvVarInput(corev1.EnvVarArgs{
										Name:  pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_SASL_MECHANISM"),
										Value: pulumi.String("SCRAM-SHA-512"),
									}),
									corev1.EnvVarInput(corev1.EnvVarArgs{
										Name:  pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_SECURITY_PROTOCOL"),
										Value: pulumi.String("SASL_PLAINTEXT"),
									}),
									corev1.EnvVarInput(corev1.EnvVarArgs{
										Name:  pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_TOPIC"),
										Value: pulumi.String(vars.SchemaRegistryKafkaStoreTopicName),
									}),
									corev1.EnvVarInput(corev1.EnvVarArgs{
										Name: pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS"),
										Value: pulumi.Sprintf("%s:%d", locals.BootstrapKubeServiceFqdn,
											vars.InternalListenerPortNumber),
									}),
									corev1.EnvVarInput(corev1.EnvVarArgs{
										Name: pulumi.String("SCHEMA_REGISTRY_KAFKASTORE_SASL_JAAS_CONFIG"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											SecretKeyRef: &corev1.SecretKeySelectorArgs{
												Name: pulumi.String(locals.AdminPasswordSecretName),
												Key:  pulumi.String(vars.SaslJaasConfigKeyInSecret),
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		deployOpts...)

	//create kubernetes service
	svcOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdService, err := kubernetescorev1.NewService(ctx,
		locals.SchemaRegistryDeploymentName,
		&kubernetescorev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.SchemaRegistryKubeServiceName),
				Namespace: pulumi.String(locals.Namespace),
			},
			Spec: &kubernetescorev1.ServiceSpecArgs{
				Type: pulumi.String("ClusterIP"),
				Selector: pulumi.StringMap{
					"app": pulumi.String(locals.SchemaRegistryDeploymentName),
				},
				Ports: kubernetescorev1.ServicePortArray{
					&kubernetescorev1.ServicePortArgs{
						Name:       pulumi.String("http"),
						Protocol:   pulumi.String("TCP"),
						Port:       pulumi.Int(80),
						TargetPort: pulumi.Int(vars.SchemaRegistryContainerPort),
					},
				},
			},
		}, svcOpts...)
	if err != nil {
		return errors.Wrapf(err, "failed to add schema registry service")
	}

	if locals.KubernetesKafka.Spec.Ingress == nil || !locals.KubernetesKafka.Spec.Ingress.Enabled {
		//skip creating ingress for schema-registry if the ingress is not enabled for kafka itself.
		return nil
	}

	//crate new certificate
	addedCertificate, err := certmanagerv1.NewCertificate(ctx,
		"schema-registry-ingress-certificate",
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String(fmt.Sprintf("%s-schema-registry",
					locals.KubernetesKafka.Metadata.Name)),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames:   pulumi.ToStringArray(locals.IngressSchemaRegistryHostnames),
				SecretName: pulumi.String(locals.IngressSchemaRegistryCertSecretName),
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					Name: pulumi.String(locals.IngressCertClusterIssuerName),
				},
			},
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "error creating schema registry certificate")
	}

	// create external gateway
	createdGateway, err := gatewayv1.NewGateway(ctx,
		"schema-registry",
		&gatewayv1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.Sprintf("%s-schema-registry-external", locals.KubernetesKafka.Metadata.Name),
				// All gateway resources should be created in the ingress deployment namespace
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(labels),
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
						Hostname: pulumi.String(locals.IngressExternalSchemaRegistryHostname),
						Port:     pulumi.Int(443),
						Protocol: pulumi.String("HTTPS"),
						Tls: &gatewayv1.GatewaySpecListenersTlsArgs{
							Mode: pulumi.String("Terminate"),
							CertificateRefs: gatewayv1.GatewaySpecListenersTlsCertificateRefsArray{
								gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs{
									Name: pulumi.String(locals.IngressSchemaRegistryCertSecretName),
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
						Hostname: pulumi.String(locals.IngressExternalSchemaRegistryHostname),
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
		}, pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{addedCertificate}))
	if err != nil {
		return errors.Wrap(err, "error creating gateway for schema-registry")
	}

	// Create HTTP route for external hostname
	_, err = gatewayv1.NewHTTPRoute(ctx,
		"schema-registry",
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(fmt.Sprintf("%s-schema-registry", locals.KubernetesKafka.Metadata.Name)),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.IngressExternalSchemaRegistryHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:      pulumi.Sprintf("%s-schema-registry-external", locals.KubernetesKafka.Metadata.Name),
						Namespace: createdGateway.Metadata.Namespace(),
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
								Name:      pulumi.String(locals.SchemaRegistryKubeServiceName),
								Namespace: pulumi.String(locals.Namespace),
								Port:      pulumi.Int(80),
							},
						},
					},
				},
			},
		}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{createdService}))

	if err != nil {
		return errors.Wrap(err, "error creating HTTP route for schema-registry")
	}
	return nil
}
