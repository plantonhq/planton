package module

import (
	kubernetesistiobasecrdsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesistiobasecrds/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values used throughout the module.
type Locals struct {
	// Release is the Istio release ref the CRDs are installed from.
	Release string

	// ManifestURL is the istio/base CRDs-only bundle URL.
	ManifestURL string

	// ResourceName is the Pulumi resource name for the CRD bundle.
	ResourceName string

	// Labels applied to managed resources.
	Labels map[string]string
}

// initializeLocals computes values from the IaC input.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesistiobasecrdsv1alpha1.KubernetesIstioBaseCrdsIacInput) *Locals {
	metadata := iacInput.Target.Metadata

	resourceName := metadata.Name + "-istio-base-crds"

	// Planton identity labels — the planton.ai/* convention, identical to the
	// Terraform module's label set (twin discipline). Note: neither engine
	// stamps these onto the CRD documents themselves (the bundle applies
	// verbatim); they identify only module-owned satellites, of which this
	// module currently has none.
	labels := map[string]string{
		"planton.ai/resource":      "true",
		"planton.ai/resource-name": metadata.Name,
		"planton.ai/resource-kind": "KubernetesIstioBaseCrds",
	}
	if metadata.Id != "" {
		labels["planton.ai/resource-id"] = metadata.Id
	}
	if metadata.Org != "" {
		labels["planton.ai/organization"] = metadata.Org
	}
	if metadata.Env != "" {
		labels["planton.ai/environment"] = metadata.Env
	}

	return &Locals{
		Release:      IstioRelease,
		ManifestURL:  GetCrdManifestURL(),
		ResourceName: resourceName,
		Labels:       labels,
	}
}
