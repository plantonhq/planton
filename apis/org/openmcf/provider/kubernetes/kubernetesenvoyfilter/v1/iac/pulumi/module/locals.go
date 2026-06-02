package module

import (
	kubernetesenvoyfilterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesenvoyfilter/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesEnvoyFilter *kubernetesenvoyfilterv1.KubernetesEnvoyFilter
	// EnvoyFilterName is the Kubernetes resource name of the EnvoyFilter. It equals the
	// OpenMCF resource metadata.name.
	EnvoyFilterName string
	// Namespace is the resolved namespace the EnvoyFilter is created in.
	Namespace string
	Labels    map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetesenvoyfilterv1.KubernetesEnvoyFilterStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	envoyFilterName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware resolves
	// valueFrom references to literal strings before the IaC module runs, so GetValue()
	// returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "envoy-filter",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "openmcf",
		"app.kubernetes.io/component":  "envoy-filter",
	}

	return &Locals{
		KubernetesEnvoyFilter: target,
		EnvoyFilterName:       envoyFilterName,
		Namespace:             namespace,
		Labels:                labels,
	}
}
