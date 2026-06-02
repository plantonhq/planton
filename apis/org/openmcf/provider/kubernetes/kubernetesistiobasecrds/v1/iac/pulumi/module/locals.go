package module

import (
	kubernetesistiobasecrdsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesistiobasecrds/v1"
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

// initializeLocals computes values from the stack input.
func initializeLocals(_ *pulumi.Context, stackInput *kubernetesistiobasecrdsv1.KubernetesIstioBaseCrdsStackInput) *Locals {
	metadata := stackInput.Target.Metadata

	resourceName := metadata.Name + "-istio-base-crds"

	labels := map[string]string{
		"app.kubernetes.io/name":       "istio-base-crds",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "openmcf",
		"app.kubernetes.io/component":  "crds",
		"istio/release":                IstioRelease,
	}

	return &Locals{
		Release:      IstioRelease,
		ManifestURL:  GetCrdManifestURL(),
		ResourceName: resourceName,
		Labels:       labels,
	}
}
