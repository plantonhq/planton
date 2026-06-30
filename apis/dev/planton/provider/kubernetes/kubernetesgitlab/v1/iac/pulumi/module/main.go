package module

import (
	"github.com/pkg/errors"
	kubernetesgitlabv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgitlab/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	networkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the single entry-point consumed by the Planton
// runtime. It wires together noun-style helpers in a Terraform-like
// top-down order so the flow is easy for DevOps engineers to follow.
func Resources(ctx *pulumi.Context, stackInput *kubernetesgitlabv1.KubernetesGitlabStackInput) error {
	// ----------------------------- locals ---------------------------------
	locals := initializeLocals(stackInput)

	// ------------------------- kubernetes provider ------------------------
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to setup kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Export namespace for reference
	ctx.Export("namespace", pulumi.String(locals.Namespace))

	// ------------------------------ service -------------------------------
	// Create a placeholder service for GitLab
	// Note: In production, this would typically use the official GitLab Helm chart
	// https://docs.gitlab.com/charts/
	serviceOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	service, err := corev1.NewService(ctx,
		locals.ServiceName,
		&corev1.ServiceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.ServiceName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: &corev1.ServiceSpecArgs{
				Type: pulumi.String("ClusterIP"),
				Ports: corev1.ServicePortArray{
					&corev1.ServicePortArgs{
						Name:       pulumi.String("http"),
						Port:       pulumi.Int(80),
						TargetPort: pulumi.Int(8080),
						Protocol:   pulumi.String("TCP"),
					},
				},
				Selector: pulumi.ToStringMap(map[string]string{
					"app":         "gitlab",
					"resource_id": locals.Labels["resource_id"],
				}),
			},
		},
		serviceOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create GitLab service")
	}

	// Export service information
	ctx.Export("service_name", service.Metadata.Name())
	ctx.Export("service_fqdn", pulumi.String(locals.ServiceFQDN))
	ctx.Export("port_forward_command", pulumi.String(locals.PortForwardCmd))

	// ----------------------------- ingress --------------------------------
	if locals.IngressHostname != "" {
		serviceName := service.Metadata.Name().Elem()
		ingressOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

		_, err := networkingv1.NewIngress(ctx,
			locals.IngressName,
			&networkingv1.IngressArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.IngressName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
					Annotations: pulumi.StringMap{
						"cert-manager.io/cluster-issuer": pulumi.String("letsencrypt-prod"),
					},
				},
				Spec: &networkingv1.IngressSpecArgs{
					IngressClassName: pulumi.String("istio"),
					Tls: networkingv1.IngressTLSArray{
						&networkingv1.IngressTLSArgs{
							Hosts: pulumi.StringArray{
								pulumi.String(locals.IngressHostname),
							},
							SecretName: pulumi.String(locals.TlsSecretName),
						},
					},
					Rules: networkingv1.IngressRuleArray{
						&networkingv1.IngressRuleArgs{
							Host: pulumi.String(locals.IngressHostname),
							Http: &networkingv1.HTTPIngressRuleValueArgs{
								Paths: networkingv1.HTTPIngressPathArray{
									&networkingv1.HTTPIngressPathArgs{
										Path:     pulumi.String("/"),
										PathType: pulumi.String("Prefix"),
										Backend: &networkingv1.IngressBackendArgs{
											Service: &networkingv1.IngressServiceBackendArgs{
												Name: serviceName,
												Port: &networkingv1.ServiceBackendPortArgs{
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
			},
			ingressOpts...)
		if err != nil {
			return errors.Wrap(err, "failed to create ingress")
		}

		// Export ingress hostname
		ctx.Export("ingress_hostname", pulumi.String(locals.IngressHostname))
		ctx.Export("external_url", pulumi.Sprintf("https://%s", locals.IngressHostname))
	}

	return nil
}
