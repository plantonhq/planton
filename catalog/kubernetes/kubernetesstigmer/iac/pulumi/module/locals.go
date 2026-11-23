package module

import (
	"strconv"

	kubernetesstigmerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstigmer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// kindName is the catalog kind's name: the value of the ownership label
// (failure.WorkloadOwnerLabel, planton.ai/kind) on everything this module
// creates, which the preflight reads to tell this kind's install from anyone
// else's.
var kindName = catalogkind.CatalogKind_KubernetesStigmer.String()

// Locals holds what every file of the module reads, resolved once from the
// input. The Terraform module's locals.tf resolves the same values.
type Locals struct {
	Spec *kubernetesstigmerv1alpha1.KubernetesStigmerSpec

	// Namespace is the Stigmer's namespace; OperatorNamespace the operator's.
	Namespace         string
	OperatorNamespace string

	// Labels go on every object the module writes itself: the namespace it
	// creates, its Secrets and ConfigMaps, and the Stigmer. Never on what the
	// chart or the operator renders. Their planton.ai/kind is also the
	// ownership mark (failure.WorkloadOwnerLabel) the preflight reads to tell
	// this kind's namespace and Stigmer from anyone else's.
	Labels map[string]string
}

func initializeLocals(iacInput *kubernetesstigmerv1alpha1.KubernetesStigmerIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: kindName,
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
		Spec:              spec,
		Namespace:         spec.GetNamespace().GetValue(),
		OperatorNamespace: spec.GetOperator().GetNamespace().GetValue(),
		Labels:            labels,
	}
}
