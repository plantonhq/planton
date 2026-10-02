package module

import (
	"github.com/pkg/errors"
	kubernetesprometheusrulev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesprometheusrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one prometheus-operator PrometheusRule.
//
// The object is the manifest's projection (manifestcr), the same one the
// generated Terraform module applies: the spec's `groups` reach the custom
// resource verbatim under their upstream keys, `namespace` becomes
// metadata.namespace, and the spec's `labels` and `annotations` become the
// object's own metadata under Planton's identity labels. Every field the spec
// gains reaches the object with no change here.
//
// No await: Prometheus loads the rules on its next configuration reload,
// which is not part of applying the object. The E2E verifier asserts the
// rules are loaded and evaluating. Terraform equivalent: kubectl_manifest
// without a wait.
func Resources(ctx *pulumi.Context, stackInput *kubernetesprometheusrulev1alpha1.KubernetesPrometheusRuleStackInput) error {
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	obj, err := manifestcr.Apply(ctx, stackInput.Target, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create prometheus rule")
	}

	ctx.Export(OpPrometheusRuleName, pulumi.String(obj.Name))
	ctx.Export(OpNamespace, pulumi.String(obj.Namespace))
	return nil
}
