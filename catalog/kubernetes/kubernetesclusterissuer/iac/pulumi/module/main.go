package module

import (
	"github.com/pkg/errors"
	kubernetesclusterissuerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesclusterissuer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/certmanagerissuer"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one cert-manager ClusterIssuer plus the credential
// Secrets its configuration needs.
//
// The CR spec is rendered by the shared certmanagerissuer builder — the SAME
// builder the KubernetesIssuer module uses, because upstream ClusterIssuer
// and Issuer share an identical spec and the two Planton kinds share the
// CertManagerIssuerConfig proto. One builder means the two kinds can never
// drift. (The typed crd2pulumi ClusterIssuer/Issuer args are two disjoint
// generated type trees for the same schema — using them here would force two
// divergent copies of this whole mapping. cert-manager's validating webhook
// checks the applied spec strictly, and the kind-cluster E2E lanes exercise
// every arm live, so shape errors still fail loudly.)
//
// Credential Secrets land in cert-manager's cluster-resource namespace —
// the ONLY namespace cert-manager reads Secrets from for cluster-scoped
// resources.
//
// Neither engine waits for the issuer to reach Ready: readiness depends on
// external reachability (the ACME server, Vault, DNS) that is not part of
// applying the resource — the same never-block-on-a-controller posture as
// Ingress. Terraform equivalent: kubectl_manifest without a wait_for block.
func Resources(ctx *pulumi.Context, iacInput *kubernetesclusterissuerv1alpha1.KubernetesClusterIssuerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	result, err := certmanagerissuer.BuildSpec(locals.ClusterIssuerName, iacInput.Target.Spec.Config)
	if err != nil {
		return errors.Wrap(err, "failed to build cluster issuer spec")
	}

	// Credential Secrets first; the CR depends on them so cert-manager
	// never observes an issuer whose secretRefs dangle.
	var secretResources []pulumi.Resource
	for _, credential := range result.Secrets {
		createdSecret, err := corev1.NewSecret(ctx, credential.Name,
			&corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(credential.Name),
					Namespace: pulumi.String(locals.SecretsNamespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				StringData: pulumi.ToSecret(pulumi.ToStringMap(credential.Data)).(pulumi.StringMapOutput),
			},
			pulumi.Provider(kubernetesProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create credential secret %s", credential.Name)
		}
		secretResources = append(secretResources, createdSecret)
	}

	opts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	if len(secretResources) > 0 {
		opts = append(opts, pulumi.DependsOn(secretResources))
	}

	_, err = apiextensions.NewCustomResource(ctx, locals.ClusterIssuerName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("ClusterIssuer"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.ClusterIssuerName),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": result.Spec,
			},
		},
		opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create cluster issuer")
	}

	ctx.Export(OpClusterIssuerName, pulumi.String(locals.ClusterIssuerName))
	ctx.Export(OpSecretsNamespace, pulumi.String(locals.SecretsNamespace))
	ctx.Export(OpAcmeAccountKeySecretName, pulumi.String(result.AcmeAccountKeySecretName))

	return nil
}
