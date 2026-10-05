package module

import (
	"github.com/pkg/errors"
	kubernetesservicemonitorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesservicemonitor/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one prometheus-operator ServiceMonitor.
//
// The object is the manifest's projection (manifestcr), the same one the
// generated Terraform module applies: every spec field reaches the
// custom resource verbatim under its upstream key, `namespace` becomes
// metadata.namespace, and the spec's `labels` and `annotations` become the
// object's own metadata under Planton's identity labels. Every field the spec
// gains reaches the object with no change here.
//
// No await: Prometheus picks the monitor up on its next configuration
// reload, which is not part of applying the object. The E2E verifier asserts
// the targets are discovered and scraped. Terraform equivalent: kubectl_manifest
// without a wait.
func Resources(ctx *pulumi.Context, iacInput *kubernetesservicemonitorv1alpha1.KubernetesServiceMonitorIacInput) error {
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	obj, err := manifestcr.Apply(ctx, iacInput.Target, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create service monitor")
	}

	ctx.Export(OpServiceMonitorName, pulumi.String(obj.Name))
	ctx.Export(OpNamespace, pulumi.String(obj.Namespace))
	return nil
}
