package module

import (
	kubernetesenvoyfilterv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesenvoyfilter/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesEnvoyFilter *kubernetesenvoyfilterv1alpha1.KubernetesEnvoyFilter
	// EnvoyFilterName is the Kubernetes resource name of the EnvoyFilter. It equals the
	// Planton resource metadata.name.
	EnvoyFilterName string
	// Namespace is the resolved namespace the EnvoyFilter is created in.
	Namespace string
	Labels    map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesenvoyfilterv1alpha1.KubernetesEnvoyFilterIacInput) *Locals {
	target := iacInput.Target
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
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "envoy-filter",
	}

	return &Locals{
		KubernetesEnvoyFilter: target,
		EnvoyFilterName:       envoyFilterName,
		Namespace:             namespace,
		Labels:                labels,
	}
}
