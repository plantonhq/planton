package module

import (
	"strconv"

	kubernetesexternalsecretv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesexternalsecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesexternalsecretv1alpha1.KubernetesExternalSecretSpec

	// Resource-identity labels stamped on the CR.
	Labels map[string]string

	// ExternalSecret metadata.name.
	ExternalSecretName string

	// Namespace the ExternalSecret (and its materialized Secret) lives in
	// (resolved literal from the spec's value-or-ref).
	Namespace string

	// Name of the Kubernetes Secret the operator materializes:
	// target.name when set, else metadata.name (upstream's own default).
	// Exported — workloads wire env/volume references to THIS name.
	SecretName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesexternalsecretv1alpha1.KubernetesExternalSecretIacInput) *Locals {
	target := iacInput.Target

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesExternalSecret.String(),
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

	secretName := target.Metadata.Name
	if target.Spec.GetTarget() != nil && target.Spec.GetTarget().GetName() != "" {
		secretName = target.Spec.GetTarget().GetName()
	}

	return &Locals{
		Spec:               target.Spec,
		Labels:             labels,
		ExternalSecretName: target.Metadata.Name,
		Namespace:          target.Spec.Namespace.GetValue(),
		SecretName:         secretName,
	}
}
