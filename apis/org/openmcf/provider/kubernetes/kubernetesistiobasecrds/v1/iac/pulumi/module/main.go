package module

import (
	"github.com/pkg/errors"
	kubernetesistiobasecrdsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesistiobasecrds/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Istio base CRDs on the target Kubernetes cluster.
func Resources(ctx *pulumi.Context, stackInput *kubernetesistiobasecrdsv1.KubernetesIstioBaseCrdsStackInput) error {
	// Initialize locals with computed values
	locals := initializeLocals(ctx, stackInput)

	// Set up kubernetes provider from the supplied cluster credential
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// --------------------------------------------------------------------
	// Apply the istio/base CRDs-only bundle.
	//
	// This installs ONLY the Istio CustomResourceDefinitions (networking,
	// security, telemetry, etc.) -- no istiod and no controller -- so the typed
	// Istio API components can be applied and server-side validated. The bundle
	// version is pinned (vars.IstioRelease) to the crd2pulumi SDK ref so the CRD
	// schema matches the typed custom resources.
	//
	// NOTE: Istio CRDs are large; if a real apply hits the client-side
	// last-applied-configuration annotation size limit, switch the kubernetes
	// provider to server-side apply.
	// --------------------------------------------------------------------
	crds, err := pulumiyaml.NewConfigFile(ctx, locals.ResourceName,
		&pulumiyaml.ConfigFileArgs{
			File: locals.ManifestURL,
		},
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to apply istio base CRDs")
	}

	// Export outputs
	return exportOutputs(ctx, locals, crds)
}
