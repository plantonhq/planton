package module

import (
	"strconv"

	kubernetesissuerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesissuer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesissuerv1alpha1.KubernetesIssuerSpec

	// Resource-identity labels stamped on the Issuer CR and every credential
	// Secret this module materializes.
	Labels map[string]string

	// The Issuer's metadata.name — the name same-namespace Certificates
	// reference.
	IssuerName string

	// The Issuer's namespace — also where its credential Secrets must live
	// (namespace-scoped issuers only read Secrets from their own namespace).
	Namespace string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesissuerv1alpha1.KubernetesIssuerIacInput) *Locals {
	target := iacInput.Target

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesIssuer.String(),
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
		Spec:       target.Spec,
		Labels:     labels,
		IssuerName: target.Metadata.Name,
		Namespace:  target.Spec.Namespace.GetValue(),
	}
}
