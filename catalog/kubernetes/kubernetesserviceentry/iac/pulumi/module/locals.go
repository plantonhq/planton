package module

import (
	kubernetesserviceentryv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesserviceentry/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesServiceEntry *kubernetesserviceentryv1alpha1.KubernetesServiceEntry
	// ServiceEntryName is the Kubernetes resource name of the ServiceEntry. It equals the
	// Planton resource metadata.name.
	ServiceEntryName string
	// Namespace is the resolved namespace the ServiceEntry is created in.
	Namespace string
	Labels    map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesserviceentryv1alpha1.KubernetesServiceEntryIacInput) *Locals {
	target := iacInput.Target
	metadata := target.Metadata
	spec := target.Spec

	serviceEntryName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware resolves
	// valueFrom references to literal strings before the IaC module runs, so GetValue()
	// returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "service-entry",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "service-entry",
	}

	return &Locals{
		KubernetesServiceEntry: target,
		ServiceEntryName:       serviceEntryName,
		Namespace:              namespace,
		Labels:                 labels,
	}
}
