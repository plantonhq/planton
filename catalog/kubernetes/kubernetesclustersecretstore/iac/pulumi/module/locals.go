package module

import (
	"strconv"

	kubernetesclustersecretstorev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesclustersecretstore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesclustersecretstorev1alpha1.KubernetesClusterSecretStoreSpec

	// Resource-identity labels stamped on the CR and the credential
	// Secrets.
	Labels map[string]string

	// ClusterSecretStore metadata.name — the name ExternalSecrets
	// reference (kind ClusterSecretStore).
	StoreName string

	// Namespace where declared credential Secrets materialize (resolved
	// literal from the spec's value-or-ref; conventionally the operator's
	// install namespace).
	SecretsNamespace string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesclustersecretstorev1alpha1.KubernetesClusterSecretStoreIacInput) *Locals {
	target := iacInput.Target

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesClusterSecretStore.String(),
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
		Spec:             target.Spec,
		Labels:           labels,
		StoreName:        target.Metadata.Name,
		SecretsNamespace: target.Spec.SecretsNamespace.GetValue(),
	}
}
