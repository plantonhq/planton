package module

import (
	"strconv"

	kuberneteskeycloakoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskeycloakoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	KubernetesKeycloakOperator *kuberneteskeycloakoperatorv1alpha1.KubernetesKeycloakOperator
	Spec                       *kuberneteskeycloakoperatorv1alpha1.KubernetesKeycloakOperatorSpec

	// ResourceName keys the applied manifest bundle in the Pulumi state.
	ResourceName string

	// Namespace the operator installs into (resolved literal from the
	// spec's value-or-ref) — stamped onto every namespaced bundle
	// document; the bundle ships without namespace fields (upstream
	// expects kustomize to set them).
	Namespace string

	// Labels tie the install back to the Planton resource. The bundle's
	// own documents keep their upstream labels untouched (faithful
	// apply); these are stamped only on the module-created namespace
	// and the module's own state identity.
	Labels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskeycloakoperatorv1alpha1.KubernetesKeycloakOperatorIacInput) *Locals {
	target := iacInput.Target

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKeycloakOperator.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	return &Locals{
		KubernetesKeycloakOperator: target,
		Spec:                       target.Spec,
		ResourceName:               target.Metadata.Name,
		Namespace:                  target.Spec.Namespace.GetValue(),
		Labels:                     labels,
	}
}
