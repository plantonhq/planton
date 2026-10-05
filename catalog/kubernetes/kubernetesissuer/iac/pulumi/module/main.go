package module

import (
	"github.com/pkg/errors"
	kubernetesissuerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesissuer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/certmanagerissuer"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one cert-manager Issuer (namespace-scoped) plus the
// credential Secrets its configuration needs — in the Issuer's OWN
// namespace, the only namespace a namespace-scoped issuer reads Secrets
// from.
//
// The CR spec is rendered by the shared certmanagerissuer builder — the SAME
// builder the KubernetesClusterIssuer module uses (upstream gives the two
// kinds an identical spec; one builder keeps the two Planton kinds
// structurally incapable of drifting). See that module's comment for why the
// shared builder is preferred over the disjoint typed crd2pulumi type trees.
//
// Neither engine waits for the issuer to reach Ready: readiness depends on
// external reachability (the ACME server, Vault, DNS) that is not part of
// applying the resource. Terraform equivalent: kubectl_manifest without a
// wait_for block.
func Resources(ctx *pulumi.Context, iacInput *kubernetesissuerv1alpha1.KubernetesIssuerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	result, err := certmanagerissuer.BuildSpec(locals.IssuerName, iacInput.Target.Spec.Config)
	if err != nil {
		return errors.Wrap(err, "failed to build issuer spec")
	}

	// Credential Secrets first; the CR depends on them so cert-manager
	// never observes an issuer whose secretRefs dangle.
	var secretResources []pulumi.Resource
	for _, credential := range result.Secrets {
		createdSecret, err := corev1.NewSecret(ctx, credential.Name,
			&corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(credential.Name),
					Namespace: pulumi.String(locals.Namespace),
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

	_, err = apiextensions.NewCustomResource(ctx, locals.IssuerName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("Issuer"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.IssuerName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": result.Spec,
			},
		},
		opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create issuer")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpIssuerName, pulumi.String(locals.IssuerName))
	ctx.Export(OpAcmeAccountKeySecretName, pulumi.String(result.AcmeAccountKeySecretName))

	return nil
}
