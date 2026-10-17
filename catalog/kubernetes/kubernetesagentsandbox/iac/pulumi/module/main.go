package module

import (
	"github.com/pkg/errors"
	kubernetesagentsandboxv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsandbox/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one agent-sandbox Sandbox.
//
// The object is the manifest's projection (manifestcr), the same one the
// generated Terraform module applies: the pod template, the volume claim
// templates, `service`, the lifecycle and the operating mode reach the custom
// resource verbatim under their upstream keys, `namespace` becomes
// metadata.namespace, and the spec's `labels` and `annotations` become the
// object's own metadata under Planton's identity labels. Every field the spec
// gains reaches the object with no change here.
//
// No await: the agent-sandbox controller creates the pod, its claims and its
// Service after the object lands, and a sandbox may be declared Suspended,
// with no pod to wait for. The E2E verifier asserts the sandbox becomes Ready
// and that suspending it keeps its volume. Terraform equivalent:
// kubectl_manifest without a wait.
func Resources(ctx *pulumi.Context, iacInput *kubernetesagentsandboxv1alpha1.KubernetesAgentSandboxIacInput) error {
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	obj, err := manifestcr.Apply(ctx, iacInput.Target, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create sandbox")
	}

	ctx.Export(OpSandboxName, pulumi.String(obj.Name))
	ctx.Export(OpNamespace, pulumi.String(obj.Namespace))
	return nil
}
