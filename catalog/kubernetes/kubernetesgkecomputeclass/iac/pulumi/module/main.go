package module

import (
	"github.com/pkg/errors"
	kubernetesgkecomputeclassv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgkecomputeclass/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one GKE ComputeClass.
//
// The object is the manifest's projection (manifestcr), the same one the
// generated Terraform module applies: every spec field reaches the custom
// resource verbatim under its upstream key, the spec's `name` becomes the
// object's metadata.name (the value pods select it by, and `default` for the
// cluster's default class), and the spec's `labels` and `annotations` become
// the object's own metadata under Planton's identity labels. Every field the
// spec gains reaches the object with no change here.
//
// No await: GKE's cluster autoscaler reads the class the next time a pod that
// selects it is pending, which is not part of applying the object. A class
// GKE cannot honor reports it in its own status conditions; the E2E verifier
// reads them. Terraform equivalent: kubectl_manifest without a wait.
func Resources(ctx *pulumi.Context, iacInput *kubernetesgkecomputeclassv1alpha1.KubernetesGkeComputeClassIacInput) error {
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	obj, err := manifestcr.Apply(ctx, iacInput.Target, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create compute class")
	}

	ctx.Export(OpComputeClassName, pulumi.String(obj.Name))
	return nil
}
