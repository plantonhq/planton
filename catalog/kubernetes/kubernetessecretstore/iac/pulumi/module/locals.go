package module

import (
	"strconv"

	kubernetessecretstorev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessecretstore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetessecretstorev1alpha1.KubernetesSecretStoreSpec

	// Resource-identity labels stamped on the CR and the credential
	// Secrets.
	Labels map[string]string

	// SecretStore metadata.name — the name ExternalSecrets in the same
	// namespace reference (kind SecretStore).
	StoreName string

	// Namespace the SecretStore and its credential Secrets live in
	// (resolved literal from the spec's value-or-ref).
	Namespace string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetessecretstorev1alpha1.KubernetesSecretStoreIacInput) *Locals {
	target := iacInput.Target

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesSecretStore.String(),
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
		Spec:      target.Spec,
		Labels:    labels,
		StoreName: target.Metadata.Name,
		Namespace: target.Spec.Namespace.GetValue(),
	}
}
