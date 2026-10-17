package module

import (
	"github.com/pkg/errors"
	kubernetesagentsandboxwarmpoolv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsandboxwarmpool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one agent-sandbox SandboxWarmPool.
//
// The object is the manifest's projection (manifestcr), the same one the
// generated Terraform module applies: the size, the template reference and
// the update strategy reach the custom resource verbatim under their upstream
// keys, `namespace` becomes metadata.namespace, and the spec's `labels` and
// `annotations` become the object's own metadata under Planton's identity
// labels. An unset `replicas` is left out of the object, so an autoscaler that
// owns it is never overwritten.
//
// No await: the controller fills the pool after the object lands, at its own
// pace. The E2E verifier asserts the pool fills to its size. Terraform
// equivalent: kubectl_manifest without a wait.
func Resources(ctx *pulumi.Context, iacInput *kubernetesagentsandboxwarmpoolv1alpha1.KubernetesAgentSandboxWarmPoolIacInput) error {
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	obj, err := manifestcr.Apply(ctx, iacInput.Target, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create sandbox warm pool")
	}

	ctx.Export(OpSandboxWarmPoolName, pulumi.String(obj.Name))
	ctx.Export(OpNamespace, pulumi.String(obj.Namespace))
	return nil
}
