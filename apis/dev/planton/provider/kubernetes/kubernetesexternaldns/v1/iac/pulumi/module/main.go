package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesexternaldnsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesexternaldns/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates all Pulumi resources for the ExternalDNS Kubernetes add‑on.
func Resources(ctx *pulumi.Context, stackInput *kubernetesexternaldnsv1.KubernetesExternalDnsStackInput) error {
	// Initialize locals with computed values
	locals := initializeLocals(stackInput)

	// Set up the Kubernetes provider from the supplied cluster credential.
	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubeProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Build ServiceAccount annotations and Helm values according to provider.
	annotations := pulumi.StringMap{}
	values := pulumi.Map{
		"serviceAccount": pulumi.Map{
			"create": pulumi.Bool(false),
			"name":   pulumi.String(locals.KsaName),
		},
	}

	switch {
	case locals.IsGke:
		values["provider"] = pulumi.String("google")
		values["google"] = pulumi.Map{
			"project": pulumi.String(locals.GkeProjectId),
		}
		// Zone‑filter keeps ExternalDNS scoped to the desired zone.
		values["zoneIdFilters"] = pulumi.StringArray{
			pulumi.String(locals.GkeDnsZoneId),
		}
		// Best‑effort GSA e‑mail derivation; users can override by patching the SA.
		annotations["iam.gke.io/gcp-service-account"] = pulumi.String(locals.GkeGsaEmail)

	case locals.IsEks:
		values["provider"] = pulumi.String("aws")
		values["zoneIdFilters"] = pulumi.StringArray{
			pulumi.String(locals.EksRoute53ZoneId),
		}
		if locals.EksIrsaRoleArn != "" {
			annotations["eks.amazonaws.com/role-arn"] = pulumi.String(locals.EksIrsaRoleArn)
		}

	case locals.IsAks:
		values["provider"] = pulumi.String("azure")
		values["zoneIdFilters"] = pulumi.StringArray{
			pulumi.String(locals.AksDnsZoneId),
		}
		if locals.AksManagedIdentityClientId != "" {
			annotations["azure.workload.identity/client-id"] =
				pulumi.String(locals.AksManagedIdentityClientId)
		}

	case locals.IsCloudflare:
		// Create secret for Cloudflare API token with unique name
		// Uses computed name: {metadata.name}-cloudflare-api-token
		cfSecretOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubeProvider)}, namespaceDeps...)
		secret, err := corev1.NewSecret(ctx, locals.CloudflareApiTokenSecretName,
			&corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.CloudflareApiTokenSecretName),
					Namespace: pulumi.String(locals.Namespace),
				},
				StringData: pulumi.StringMap{
					"apiKey": pulumi.String(locals.CfApiToken),
				},
			},
			cfSecretOpts...)
		if err != nil {
			return errors.Wrap(err, "failed to create cloudflare api token secret")
		}

		// Configure provider
		values["provider"] = pulumi.String("cloudflare")

		// Top-level sources parameter (REQUIRED for RBAC generation)
		// Sources determine which Kubernetes resources ExternalDNS watches for DNS records:
		// - service: Services with external-dns.alpha.kubernetes.io/hostname annotation
		// - ingress: Ingress resources
		// - gateway-httproute: Gateway API HTTPRoute resources
		// - istio-gateway: Istio Gateway resources (networking.istio.io/v1 Gateway)
		values["sources"] = pulumi.StringArray{
			pulumi.String("service"),
			pulumi.String("ingress"),
			pulumi.String("gateway-httproute"),
			pulumi.String("istio-gateway"),
		}

		// Mount the API token secret as environment variable
		values["env"] = pulumi.Array{
			pulumi.Map{
				"name": pulumi.String("CF_API_TOKEN"),
				"valueFrom": pulumi.Map{
					"secretKeyRef": pulumi.Map{
						"name": secret.Metadata.Name(),
						"key":  pulumi.String("apiKey"),
					},
				},
			},
		}

		// Configure extra args for Cloudflare-specific features only
		extraArgs := pulumi.StringArray{
			pulumi.String("--cloudflare-dns-records-per-page=5000"),
			pulumi.String(fmt.Sprintf("--zone-id-filter=%s", locals.CfDnsZoneId)),
		}

		// Add proxy flag if enabled
		if locals.CfIsProxied {
			extraArgs = append(extraArgs, pulumi.String("--cloudflare-proxied"))
		}

		values["extraArgs"] = extraArgs

	default:
		return errors.New("spec.provider_config must be set (gke, eks, aks, or cloudflare)")
	}

	// Honor an optional custom ExternalDNS version.
	if locals.ExternalDnsVersion != "" {
		values["image"] = pulumi.Map{
			"tag": pulumi.String(locals.ExternalDnsVersion),
		}
	}

	// Create the ServiceAccount.
	saOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubeProvider)}, namespaceDeps...)
	_, err = corev1.NewServiceAccount(ctx, locals.KsaName,
		&corev1.ServiceAccountArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:        pulumi.String(locals.KsaName),
				Namespace:   pulumi.String(locals.Namespace),
				Annotations: annotations,
			},
		},
		saOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create service account")
	}

	// Deploy the Helm release.
	helmOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubeProvider)}, namespaceDeps...)
	_, err = helm.NewRelease(ctx, locals.ReleaseName,
		&helm.ReleaseArgs{
			Name:            pulumi.String(locals.ReleaseName),
			Namespace:       pulumi.String(locals.Namespace),
			Chart:           pulumi.String(vars.HelmChartName),
			Version:         pulumi.String(locals.HelmChartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			Values:          values,
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmChartRepo),
			},
		},
		helmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to install external-dns helm release")
	}

	// Export stack outputs.
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpSolverSa, pulumi.String(locals.KsaName))

	return nil
}
