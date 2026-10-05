package module

import (
	kubernetesbackendtlspolicyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesbackendtlspolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strconv"
)

// Locals holds the resolved inputs the module operates on: the full target
// resource plus the scalar identifiers used for the resource name, namespace,
// labels, and outputs.
type Locals struct {
	KubernetesBackendTlsPolicy *kubernetesbackendtlspolicyv1alpha1.KubernetesBackendTlsPolicy
	PolicyName                 string
	Namespace                  string
	Labels                     map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesbackendtlspolicyv1alpha1.KubernetesBackendTlsPolicyIacInput) *Locals {
	target := iacInput.Target
	metadata := target.Metadata
	spec := target.Spec

	// namespace is a StringValueOrRef foreign key. The platform middleware
	// resolves valueFrom references to literal strings before the IaC module
	// runs, so GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesBackendTlsPolicy.String(),
	}
	if metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = metadata.Id
	}
	if metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = metadata.Env
	}

	return &Locals{
		KubernetesBackendTlsPolicy: target,
		PolicyName:                 metadata.Name,
		Namespace:                  namespace,
		Labels:                     labels,
	}
}
