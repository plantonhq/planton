package module

import (
	"github.com/pkg/errors"
	kubernetesexternalsecretv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesexternalsecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one External Secrets Operator ExternalSecret — the
// declaration that syncs entries from a store's backend into a materialized
// Kubernetes Secret in this namespace.
//
// The CR spec renders from the typed fields (spec_builder.go), applied as an
// untyped custom resource — the same posture as the store kinds and the
// cert-manager family: ESO's validating webhook checks the applied spec
// strictly, and the kind-cluster E2E lanes verify the full sync loop live
// (store Ready, ExternalSecret synced, Secret contents), so shape errors
// fail loudly.
//
// Neither engine waits for the SYNC to complete: the materialized Secret
// appears when the operator reaches the backend, which is not part of
// applying the resource. The E2E verifier (not the module) asserts the
// synced state. Terraform equivalent: kubectl_manifest without a wait_for
// block.
func Resources(ctx *pulumi.Context, iacInput *kubernetesexternalsecretv1alpha1.KubernetesExternalSecretIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	_, err = apiextensions.NewCustomResource(ctx, locals.ExternalSecretName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("external-secrets.io/v1"),
			Kind:       pulumi.String("ExternalSecret"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.ExternalSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": buildExternalSecretSpec(locals),
			},
		},
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create external secret")
	}

	ctx.Export(OpExternalSecretName, pulumi.String(locals.ExternalSecretName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpSecretName, pulumi.String(locals.SecretName))

	return nil
}
