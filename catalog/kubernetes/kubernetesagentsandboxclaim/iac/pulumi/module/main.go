package module

import (
	"github.com/pkg/errors"
	kubernetesagentsandboxclaimv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsandboxclaim/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one agent-sandbox SandboxClaim.
//
// The object is the manifest's projection (manifestcr), the same one the
// generated Terraform module applies: the pool reference, the lifecycle, the
// pod metadata, the environment variables and the volume claim templates reach
// the custom resource verbatim under their upstream keys, `namespace` becomes
// metadata.namespace, and the spec's `labels` and `annotations` become the
// object's own metadata under Planton's identity labels.
//
// No await: the controller binds the claim to a sandbox after the object
// lands. The E2E verifier asserts the claim binds, warm when it should.
// Terraform equivalent: kubectl_manifest without a wait.
func Resources(ctx *pulumi.Context, iacInput *kubernetesagentsandboxclaimv1alpha1.KubernetesAgentSandboxClaimIacInput) error {
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	obj, err := manifestcr.Apply(ctx, iacInput.Target, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create sandbox claim")
	}

	ctx.Export(OpSandboxClaimName, pulumi.String(obj.Name))
	ctx.Export(OpNamespace, pulumi.String(obj.Namespace))
	return nil
}
