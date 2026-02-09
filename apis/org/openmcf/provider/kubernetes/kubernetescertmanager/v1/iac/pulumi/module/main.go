package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetescertmanagerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetescertmanager/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	apiextensionsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources create all Pulumi resources for the Cert‑Manager Kubernetes add‑on.
func Resources(ctx *pulumi.Context, stackInput *kubernetescertmanagerv1.KubernetesCertManagerStackInput) error {
	// Initialize locals with computed values
	locals := initializeLocals(ctx, stackInput)

	// set up a kubernetes provider from the supplied cluster credential
	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	spec := stackInput.Target.Spec

	// validate spec has required ACME config and at least one DNS provider
	if spec.Acme == nil {
		return errors.New("spec.acme is required")
	}
	if len(spec.DnsProviders) == 0 {
		return errors.New("spec.dns_providers must contain at least one provider")
	}

	// get chart version from spec (proto default: "v1.19.1")
	chartVersion := spec.GetHelmChartVersion()

	// conditionally create namespace based on spec.CreateNamespace
	createdNamespace, err := namespace(ctx, stackInput, locals, kubeProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// build identity annotation map for ServiceAccount
	// multiple providers may need different identities configured
	annotations := pulumi.StringMap{}

	// process each DNS provider to build service account annotations
	for _, dnsProvider := range spec.DnsProviders {
		if gcp := dnsProvider.GetGcpCloudDns(); gcp != nil {
			annotations["iam.gke.io/gcp-service-account"] = pulumi.String(gcp.ServiceAccountEmail)
		} else if aws := dnsProvider.GetAwsRoute53(); aws != nil {
			annotations["eks.amazonaws.com/role-arn"] = pulumi.String(aws.RoleArn)
		} else if azure := dnsProvider.GetAzureDns(); azure != nil {
			annotations["azure.workload.identity/client-id"] = pulumi.String(azure.ClientId)
		}
		// Cloudflare providers don't need ServiceAccount annotations
	}

	// create a ServiceAccount with the chosen annotations
	// ServiceAccount name uses metadata.name for uniqueness when multiple instances share a namespace
	saOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubeProvider)}, namespaceDeps...)
	_, err = corev1.NewServiceAccount(ctx, locals.ServiceAccountName,
		&corev1.ServiceAccountArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:        pulumi.String(locals.ServiceAccountName),
				Namespace:   pulumi.String(locals.Namespace),
				Annotations: annotations,
			},
		},
		saOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create service account")
	}

	// deploy cert‑manager helm chart with DNS resolver configuration
	helmOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubeProvider)}, namespaceDeps...)
	helmRelease, err := helm.NewRelease(ctx, "cert-manager",
		&helm.ReleaseArgs{
			Name:            pulumi.String(vars.HelmChartName),
			Namespace:       pulumi.String(locals.Namespace),
			Chart:           pulumi.String(vars.HelmChartName),
			Version:         pulumi.String(chartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			Values: pulumi.Map{
				"installCRDs": pulumi.Bool(true),
				"serviceAccount": pulumi.Map{
					"create": pulumi.Bool(false),
					"name":   pulumi.String(locals.ServiceAccountName),
				},
				// Configure DNS resolvers for reliable DNS-01 propagation checks
				"extraArgs": pulumi.Array{
					pulumi.String("--dns01-recursive-nameservers-only"),
					pulumi.String("--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"),
				},
			},
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmChartRepo),
			},
		},
		helmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to install kubernetes-cert-manager helm release")
	}

	// create secrets for Cloudflare providers
	// Secret names use locals.CloudflareSecretName() for uniqueness when multiple instances share a namespace
	cloudflareSecrets := make(map[string]pulumi.StringOutput)
	for _, dnsProvider := range spec.DnsProviders {
		if cf := dnsProvider.GetCloudflare(); cf != nil {
			secretName := locals.CloudflareSecretName(dnsProvider.Name)
			secretOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubeProvider)}, namespaceDeps...)
			secret, err := corev1.NewSecret(ctx, secretName,
				&corev1.SecretArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Name:      pulumi.String(secretName),
						Namespace: pulumi.String(locals.Namespace),
					},
					StringData: pulumi.StringMap{
						"api-token": pulumi.String(cf.ApiToken),
					},
				},
				secretOpts...)
			if err != nil {
				return errors.Wrapf(err, "failed to create cloudflare secret for provider %s", dnsProvider.Name)
			}
			cloudflareSecrets[dnsProvider.Name] = secret.Metadata.Name().Elem()
		}
	}

	// create one ClusterIssuer per domain for better visibility
	clusterIssuerNames := make([]string, 0)
	for _, dnsProvider := range spec.DnsProviders {
		for _, dnsZone := range dnsProvider.DnsZones {
			// Create ClusterIssuer named after the domain
			err = createClusterIssuerForDomain(ctx, kubeProvider, helmRelease, spec, cloudflareSecrets, dnsProvider, dnsZone)
			if err != nil {
				return errors.Wrapf(err, "failed to create cluster issuer for domain %s", dnsZone)
			}
			clusterIssuerNames = append(clusterIssuerNames, dnsZone)
		}
	}

	// export stack outputs
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(vars.HelmChartName))
	ctx.Export(OpClusterIssuerNames, pulumi.ToStringArray(clusterIssuerNames))

	return nil
}

// createClusterIssuerForDomain creates a ClusterIssuer for a single domain
func createClusterIssuerForDomain(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	helmRelease *helm.Release,
	spec *kubernetescertmanagerv1.KubernetesCertManagerSpec,
	cloudflareSecrets map[string]pulumi.StringOutput,
	dnsProvider *kubernetescertmanagerv1.DnsProviderConfig,
	domain string,
) error {
	// ClusterIssuer name is the domain itself for better visibility
	issuerName := domain

	// build the single solver for this domain
	var solverConfig map[string]interface{}

	if gcp := dnsProvider.GetGcpCloudDns(); gcp != nil {
		solverConfig = map[string]interface{}{
			"dns01": map[string]interface{}{
				"cloudDNS": map[string]interface{}{
					"project": gcp.ProjectId,
				},
			},
		}
	} else if aws := dnsProvider.GetAwsRoute53(); aws != nil {
		solverConfig = map[string]interface{}{
			"dns01": map[string]interface{}{
				"route53": map[string]interface{}{
					"region": aws.Region,
				},
			},
		}
	} else if azure := dnsProvider.GetAzureDns(); azure != nil {
		solverConfig = map[string]interface{}{
			"dns01": map[string]interface{}{
				"azureDNS": map[string]interface{}{
					"subscriptionID":    azure.SubscriptionId,
					"resourceGroupName": azure.ResourceGroup,
				},
			},
		}
	} else if cf := dnsProvider.GetCloudflare(); cf != nil {
		// get the secret name for this provider
		secretName, ok := cloudflareSecrets[dnsProvider.Name]
		if !ok {
			return errors.Errorf("cloudflare secret not found for provider %s", dnsProvider.Name)
		}

		solverConfig = map[string]interface{}{
			"dns01": map[string]interface{}{
				"cloudflare": map[string]interface{}{
					"apiTokenSecretRef": map[string]interface{}{
						"name": secretName,
						"key":  "api-token",
					},
				},
			},
		}
	}

	// create the ClusterIssuer for this domain
	_, err := apiextensionsv1.NewCustomResource(ctx, issuerName,
		&apiextensionsv1.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("ClusterIssuer"),
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(issuerName),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"acme": map[string]interface{}{
						"email":  spec.Acme.Email,
						"server": spec.Acme.GetServer(),
						"privateKeySecretRef": map[string]interface{}{
							"name": fmt.Sprintf("letsencrypt-%s-account-key", domain),
						},
						"solvers": []interface{}{solverConfig},
					},
				},
			},
		},
		pulumi.Provider(kubeProvider),
		pulumi.DependsOn([]pulumi.Resource{helmRelease}))

	return err
}
