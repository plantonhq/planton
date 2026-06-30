package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesclusterissuerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesclusterissuer/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesclusterissuerv1.KubernetesClusterIssuerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	spec := stackInput.Target.Spec

	var clusterIssuerDeps []pulumi.Resource

	if cf := spec.GetCloudflare(); cf != nil {
		secret, err := createCloudflareSecret(ctx, kubeProvider, locals, cf)
		if err != nil {
			return errors.Wrap(err, "failed to create cloudflare secret")
		}
		clusterIssuerDeps = append(clusterIssuerDeps, secret)
	}

	if err := createClusterIssuer(ctx, kubeProvider, locals, spec, clusterIssuerDeps); err != nil {
		return errors.Wrap(err, "failed to create cluster issuer")
	}

	ctx.Export(OpClusterIssuerName, pulumi.String(locals.DnsDomain))
	ctx.Export(OpAcmeAccountKeySecretName, pulumi.String(locals.AcmeAccountKeySecretName))

	return nil
}

func createCloudflareSecret(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
	cf *kubernetesclusterissuerv1.CloudflareDnsSolver,
) (*corev1.Secret, error) {
	return corev1.NewSecret(ctx, locals.CloudflareSecretName,
		&corev1.SecretArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.CloudflareSecretName),
				Namespace: pulumi.String(locals.CertManagerNamespace),
			},
			StringData: pulumi.StringMap{
				"api-token": pulumi.String(cf.ApiToken),
			},
		},
		pulumi.Provider(kubeProvider))
}

// createClusterIssuer creates a cert-manager ClusterIssuer using the typed
// crd2pulumi SDK (certmanagerv1.NewClusterIssuer). This provides compile-time
// type safety for the full ACME/DNS01 solver configuration hierarchy,
// consistent with how all other cert-manager resources in Planton (15+
// components using NewCertificate) use the typed SDK.
//
// Previously this function used apiextensionsv1.NewCustomResource with
// untyped OtherFields maps. The typed approach catches field name and
// structure errors at compile time rather than at deployment time.
func createClusterIssuer(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
	spec *kubernetesclusterissuerv1.KubernetesClusterIssuerSpec,
	deps []pulumi.Resource,
) error {
	dns01Config, err := buildDns01SolverConfig(spec, locals)
	if err != nil {
		return err
	}

	opts := []pulumi.ResourceOption{pulumi.Provider(kubeProvider)}
	if len(deps) > 0 {
		opts = append(opts, pulumi.DependsOn(deps))
	}

	// Metadata uses value ObjectMetaArgs (not pointer), matching the pattern
	// established by all 17 ingress components that use certmanagerv1.NewCertificate.
	_, err = certmanagerv1.NewClusterIssuer(ctx, locals.DnsDomain,
		&certmanagerv1.ClusterIssuerArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String(locals.DnsDomain),
			},
			Spec: certmanagerv1.ClusterIssuerSpecArgs{
				Acme: certmanagerv1.ClusterIssuerSpecAcmeArgs{
					Email:  pulumi.String(spec.Acme.Email),
					Server: pulumi.String(spec.Acme.GetServer()),
					PrivateKeySecretRef: certmanagerv1.ClusterIssuerSpecAcmePrivateKeySecretRefArgs{
						Name: pulumi.String(locals.AcmeAccountKeySecretName),
					},
					Solvers: certmanagerv1.ClusterIssuerSpecAcmeSolversArray{
						certmanagerv1.ClusterIssuerSpecAcmeSolversArgs{
							Dns01: dns01Config,
						},
					},
				},
			},
		},
		opts...)

	return err
}

// buildDns01SolverConfig maps the proto provider oneof to the typed
// cert-manager DNS01 solver configuration. Each provider branch populates
// only the fields that cert-manager needs for that DNS provider.
//
// All four providers use DNS-01 challenges (not HTTP-01), so the return type
// is the DNS01-specific args struct which the caller wraps in a
// ClusterIssuerSpecAcmeSolversArgs.
//
// crd2pulumi field naming follows the cert-manager CRD's JSON field names,
// which differ from both the proto field names and Go conventions:
//   - CloudDNS, not CloudDns (CRD JSON: "cloudDNS")
//   - AzureDNS, not AzureDns (CRD JSON: "azureDNS")
//   - SubscriptionID, not SubscriptionId (CRD JSON: "subscriptionID")
//   - ResourceGroupName, not ResourceGroup (CRD JSON: "resourceGroupName")
func buildDns01SolverConfig(
	spec *kubernetesclusterissuerv1.KubernetesClusterIssuerSpec,
	locals *Locals,
) (certmanagerv1.ClusterIssuerSpecAcmeSolversDns01Args, error) {
	if gcp := spec.GetGcpCloudDns(); gcp != nil {
		return certmanagerv1.ClusterIssuerSpecAcmeSolversDns01Args{
			CloudDNS: certmanagerv1.ClusterIssuerSpecAcmeSolversDns01CloudDNSArgs{
				Project: pulumi.String(gcp.ProjectId),
			},
		}, nil
	}

	if aws := spec.GetAwsRoute53(); aws != nil {
		return certmanagerv1.ClusterIssuerSpecAcmeSolversDns01Args{
			Route53: certmanagerv1.ClusterIssuerSpecAcmeSolversDns01Route53Args{
				Region: pulumi.String(aws.Region),
			},
		}, nil
	}

	if azure := spec.GetAzureDns(); azure != nil {
		return certmanagerv1.ClusterIssuerSpecAcmeSolversDns01Args{
			AzureDNS: certmanagerv1.ClusterIssuerSpecAcmeSolversDns01AzureDNSArgs{
				SubscriptionID:    pulumi.String(azure.SubscriptionId),
				ResourceGroupName: pulumi.String(azure.ResourceGroup),
			},
		}, nil
	}

	if spec.GetCloudflare() != nil {
		return certmanagerv1.ClusterIssuerSpecAcmeSolversDns01Args{
			Cloudflare: certmanagerv1.ClusterIssuerSpecAcmeSolversDns01CloudflareArgs{
				ApiTokenSecretRef: certmanagerv1.ClusterIssuerSpecAcmeSolversDns01CloudflareApiTokenSecretRefArgs{
					Name: pulumi.String(locals.CloudflareSecretName),
					Key:  pulumi.String("api-token"),
				},
			},
		}, nil
	}

	return certmanagerv1.ClusterIssuerSpecAcmeSolversDns01Args{},
		fmt.Errorf("no DNS provider configured -- spec.provider oneof must have exactly one branch set (cloudflare, gcp_cloud_dns, aws_route53, or azure_dns)")
}
