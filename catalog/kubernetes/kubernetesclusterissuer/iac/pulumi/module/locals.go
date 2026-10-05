package module

import (
	"strconv"

	kubernetesclusterissuerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesclusterissuer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesclusterissuerv1alpha1.KubernetesClusterIssuerSpec

	// Resource-identity labels stamped on the ClusterIssuer CR and every
	// credential Secret this module materializes.
	Labels map[string]string

	// The ClusterIssuer's metadata.name — the name Certificates and
	// ingress-shim annotations reference.
	ClusterIssuerName string

	// Namespace credential Secrets are created in: cert-manager's
	// cluster-resource namespace (resolved from the spec's value-or-ref).
	SecretsNamespace string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesclusterissuerv1alpha1.KubernetesClusterIssuerIacInput) *Locals {
	target := iacInput.Target

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesClusterIssuer.String(),
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
		Spec:              target.Spec,
		Labels:            labels,
		ClusterIssuerName: target.Metadata.Name,
		SecretsNamespace:  target.Spec.CertManagerNamespace.GetValue(),
	}
}
