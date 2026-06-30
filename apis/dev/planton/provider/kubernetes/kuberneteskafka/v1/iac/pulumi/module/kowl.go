package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/fileutil"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/strimzioperator/kubernetes/kafka/v1beta2"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kowl(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	createdKafkaCluster *v1beta2.Kafka, namespaceDeps []pulumi.ResourceOption) error {

	type kowlConfigTemplateInput struct {
		BootstrapKubeServiceFqdn       string
		BootstrapServerKubeServicePort int
		SaslUsername                   string
		SchemaRegistryHostname         string
		RefreshIntervalMinutes         int
		IsSchemaRegistryEnabled        bool
	}

	kowlConfig, err := fileutil.RenderTemplate(&kowlConfigTemplateInput{
		BootstrapKubeServiceFqdn:       locals.BootstrapKubeServiceFqdn,
		BootstrapServerKubeServicePort: vars.InternalListenerPortNumber,
		SaslUsername:                   locals.AdminUsername,
		SchemaRegistryHostname:         locals.SchemaRegistryKubeServiceFqdn,
		RefreshIntervalMinutes:         vars.KowlRefreshIntervalMinutes,
		IsSchemaRegistryEnabled:        locals.KubernetesKafka.Spec.SchemaRegistryContainer.IsEnabled,
	}, vars.KowlConfigFileTemplate)
	if err != nil {
		return errors.Wrap(err, "failed to render kowl config file")
	}

	cmOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdConfigMap, err := kubernetescorev1.NewConfigMap(ctx,
		locals.KowlConfigMapName,
		&kubernetescorev1.ConfigMapArgs{
			Data: pulumi.ToStringMap(map[string]string{vars.KowlConfigKeyName: string(kowlConfig)}),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.KowlConfigMapName),
				Namespace: pulumi.String(locals.Namespace),
			},
		}, cmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to add config-map")
	}

	labels := locals.Labels
	labels["app"] = locals.KowlDeploymentName

	_, err = appsv1.NewDeployment(ctx, locals.KowlDeploymentName, &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.KowlDeploymentName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(labels),
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String(locals.KowlDeploymentName),
				},
			},
			Template: &kubernetescorev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String(locals.KowlDeploymentName),
					},
				},
				Spec: &kubernetescorev1.PodSpecArgs{
					Volumes: kubernetescorev1.VolumeArray{
						kubernetescorev1.VolumeArgs{
							ConfigMap: kubernetescorev1.ConfigMapVolumeSourceArgs{
								Name: createdConfigMap.Metadata.Name(),
							},
							Name: pulumi.String(vars.KowlConfigVolumeName),
						},
					},
					Containers: kubernetescorev1.ContainerArray{
						&kubernetescorev1.ContainerArgs{
							Name:  pulumi.String(locals.KowlDeploymentName),
							Image: pulumi.String(vars.KowlDockerImage),
							Args: pulumi.ToStringArray([]string{
								//https://github.com/cloudhut/charts/blob/master/kowl/templates/deployment.yaml
								fmt.Sprintf("--config.filepath=%s", vars.KowlConfigVolumeMountPath),
								fmt.Sprintf("--kafka.sasl.password=$%s", vars.KowlEnvVarNameKafkaSaslPassword),
							}),
							Ports: kubernetescorev1.ContainerPortArray{
								&kubernetescorev1.ContainerPortArgs{
									Name:          pulumi.String("http"),
									ContainerPort: pulumi.Int(vars.KowlContainerPort),
								},
							},
							Env: kubernetescorev1.EnvVarArray{
								kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
									Name: pulumi.String(vars.KowlEnvVarNameKafkaSaslPassword),
									ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
										SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
											Name: pulumi.String(locals.AdminPasswordSecretName),
											Key:  pulumi.String(vars.SaslPasswordKeyInSecret),
										},
									},
								}),
							},
							VolumeMounts: kubernetescorev1.VolumeMountArray{
								kubernetescorev1.VolumeMountArgs{
									MountPath: pulumi.String(vars.KowlConfigVolumeMountPath),
									Name:      pulumi.String(vars.KowlConfigVolumeName),
									SubPath:   pulumi.String(vars.KowlConfigKeyName),
								},
							},
							Resources: kubernetescorev1.ResourceRequirementsArgs{
								Limits: pulumi.ToStringMap(map[string]string{
									"cpu":    vars.KowlCpuLimits,
									"memory": vars.KowlMemoryLimits,
								}),
								Requests: pulumi.ToStringMap(map[string]string{
									"cpu":    vars.KowlCpuRequests,
									"memory": vars.KowlMemoryRequests,
								}),
							},
						},
					},
				},
			},
		},
	},
		pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{createdKafkaCluster, createdConfigMap}),
		pulumi.IgnoreChanges([]string{"metadata", "status"}))
	if err != nil {
		return errors.Wrap(err, "failed to add kowl deployment")
	}

	//create service
	createdService, err := kubernetescorev1.NewService(ctx,
		locals.KowlDeploymentName,
		&kubernetescorev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.KowlKubeServiceName),
				Namespace: pulumi.String(locals.Namespace),
			},
			Spec: &kubernetescorev1.ServiceSpecArgs{
				Type: pulumi.String("ClusterIP"),
				Selector: pulumi.StringMap{
					"app": pulumi.String(locals.KowlDeploymentName),
				},
				Ports: kubernetescorev1.ServicePortArray{
					&kubernetescorev1.ServicePortArgs{
						Name:       pulumi.String("http"),
						Protocol:   pulumi.String("TCP"),
						Port:       pulumi.Int(80),
						TargetPort: pulumi.Int(vars.KowlContainerPort),
					},
				},
			},
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to add kowl service")
	}

	if locals.KubernetesKafka.Spec.Ingress == nil || !locals.KubernetesKafka.Spec.Ingress.Enabled {
		//skip creating ingress for kowl if the ingress is not enabled for kafka itself.
		return nil
	}

	//crate new certificate
	addedCertificate, err := certmanagerv1.NewCertificate(ctx,
		"kowl-ingress-certificate",
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(fmt.Sprintf("%s-kowl", locals.KubernetesKafka.Metadata.Name)),
				Namespace: pulumi.String(vars.IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames: pulumi.StringArray{
					pulumi.String(locals.IngressExternalKowlHostname),
				},
				SecretName: pulumi.String(locals.IngressKowlCertSecretName),
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					Name: pulumi.String(locals.IngressCertClusterIssuerName),
				},
			},
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "error creating kowl certificate")
	}

	// create external gateway
	createdGateway, err := gatewayv1.NewGateway(ctx,
		"kowl",
		&gatewayv1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.Sprintf("%s-kowl-external", locals.KubernetesKafka.Metadata.Name),
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
						Hostname: pulumi.String(locals.IngressExternalKowlHostname),
						Port:     pulumi.Int(443),
						Protocol: pulumi.String("HTTPS"),
						Tls: &gatewayv1.GatewaySpecListenersTlsArgs{
							Mode: pulumi.String("Terminate"),
							CertificateRefs: gatewayv1.GatewaySpecListenersTlsCertificateRefsArray{
								gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs{
									Name: pulumi.String(locals.IngressKowlCertSecretName),
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
						Hostname: pulumi.String(locals.IngressExternalKowlHostname),
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
		return errors.Wrap(err, "error creating gateway for kowl")
	}

	// Create HTTP route for external hostname
	_, err = gatewayv1.NewHTTPRoute(ctx,
		"kowl",
		&gatewayv1.HTTPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(fmt.Sprintf("%s-kowl", locals.KubernetesKafka.Metadata.Name)),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: gatewayv1.HTTPRouteSpecArgs{
				Hostnames: pulumi.StringArray{pulumi.String(locals.IngressExternalKowlHostname)},
				ParentRefs: gatewayv1.HTTPRouteSpecParentRefsArray{
					gatewayv1.HTTPRouteSpecParentRefsArgs{
						Name:      pulumi.Sprintf("%s-kowl-external", locals.KubernetesKafka.Metadata.Name),
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
								Name:      pulumi.String(locals.KowlKubeServiceName),
								Namespace: pulumi.String(locals.Namespace),
								Port:      pulumi.Int(80),
							},
						},
					},
				},
			},
		}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{createdService}))

	if err != nil {
		return errors.Wrap(err, "error creating HTTP route for kowl")
	}

	return nil
}
