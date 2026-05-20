package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesclusterissuerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesclusterissuer/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	apiextensionsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
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

func createClusterIssuer(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
	spec *kubernetesclusterissuerv1.KubernetesClusterIssuerSpec,
	deps []pulumi.Resource,
) error {
	solverConfig, err := buildSolverConfig(spec, locals)
	if err != nil {
		return err
	}

	opts := []pulumi.ResourceOption{pulumi.Provider(kubeProvider)}
	if len(deps) > 0 {
		opts = append(opts, pulumi.DependsOn(deps))
	}

	_, err = apiextensionsv1.NewCustomResource(ctx, locals.DnsDomain,
		&apiextensionsv1.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("ClusterIssuer"),
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(locals.DnsDomain),
			},
			OtherFields: map[string]interface{}{
				"spec": map[string]interface{}{
					"acme": map[string]interface{}{
						"email":  spec.Acme.Email,
						"server": spec.Acme.GetServer(),
						"privateKeySecretRef": map[string]interface{}{
							"name": locals.AcmeAccountKeySecretName,
						},
						"solvers": []interface{}{solverConfig},
					},
				},
			},
		},
		opts...)

	return err
}

func buildSolverConfig(
	spec *kubernetesclusterissuerv1.KubernetesClusterIssuerSpec,
	locals *Locals,
) (map[string]interface{}, error) {
	if gcp := spec.GetGcpCloudDns(); gcp != nil {
		return map[string]interface{}{
			"dns01": map[string]interface{}{
				"cloudDNS": map[string]interface{}{
					"project": gcp.ProjectId,
				},
			},
		}, nil
	}

	if aws := spec.GetAwsRoute53(); aws != nil {
		return map[string]interface{}{
			"dns01": map[string]interface{}{
				"route53": map[string]interface{}{
					"region": aws.Region,
				},
			},
		}, nil
	}

	if azure := spec.GetAzureDns(); azure != nil {
		return map[string]interface{}{
			"dns01": map[string]interface{}{
				"azureDNS": map[string]interface{}{
					"subscriptionID":    azure.SubscriptionId,
					"resourceGroupName": azure.ResourceGroup,
				},
			},
		}, nil
	}

	if spec.GetCloudflare() != nil {
		return map[string]interface{}{
			"dns01": map[string]interface{}{
				"cloudflare": map[string]interface{}{
					"apiTokenSecretRef": map[string]interface{}{
						"name": locals.CloudflareSecretName,
						"key":  "api-token",
					},
				},
			},
		}, nil
	}

	return nil, fmt.Errorf("no DNS provider configured")
}
