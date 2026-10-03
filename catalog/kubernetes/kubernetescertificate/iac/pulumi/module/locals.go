package module

import (
	"strconv"

	kubernetescertificatev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetescertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetescertificatev1alpha1.KubernetesCertificateSpec

	// Resource-identity labels stamped on the Certificate CR (the output
	// Secret's labels are user-controlled via secret_template).
	Labels map[string]string

	// The Certificate's namespace (resolved literal from the value-or-ref).
	Namespace string

	// The Certificate's metadata.name.
	CertificateName string

	// The TLS Secret name consumers reference — exported as an output.
	SecretName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetescertificatev1alpha1.KubernetesCertificateIacInput) *Locals {
	target := iacInput.Target

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesCertificate.String(),
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
		Spec:            target.Spec,
		Labels:          labels,
		Namespace:       target.Spec.Namespace.GetValue(),
		CertificateName: target.Metadata.Name,
		SecretName:      target.Spec.SecretName,
	}
}
