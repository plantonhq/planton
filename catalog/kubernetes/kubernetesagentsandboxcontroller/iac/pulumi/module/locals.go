package module

import (
	kubernetesagentsandboxcontrollerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsandboxcontroller/v1alpha1"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared"
)

// Locals holds the values derived from the IaC input once.
type Locals struct {
	Spec *kubernetesagentsandboxcontrollerv1alpha1.KubernetesAgentSandboxControllerSpec

	// ResourceName keys the applied groups in the Pulumi state.
	ResourceName string

	// IdentityLabels go on the Namespace this module creates.
	IdentityLabels map[string]string
}

func initializeLocals(iacInput *kubernetesagentsandboxcontrollerv1alpha1.KubernetesAgentSandboxControllerIacInput) *Locals {
	return &Locals{
		Spec:           iacInput.GetTarget().GetSpec(),
		ResourceName:   iacInput.GetTarget().GetMetadata().GetName(),
		IdentityLabels: identityLabels(iacInput.GetTarget().GetMetadata()),
	}
}

// identityLabels are the resource-identity labels. The kind label is also the
// ownership mark the namespace check reads. Terraform twin: identity_labels
// in locals.tf.
func identityLabels(metadata *shared.CatalogObjectMetadata) map[string]string {
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     "true",
		kuberneteslabelkeys.ResourceName: metadata.GetName(),
		failure.WorkloadOwnerLabel:       kindName,
	}
	if metadata.GetId() != "" {
		labels[kuberneteslabelkeys.ResourceId] = metadata.GetId()
	}
	if metadata.GetOrg() != "" {
		labels[kuberneteslabelkeys.Organization] = metadata.GetOrg()
	}
	if metadata.GetEnv() != "" {
		labels[kuberneteslabelkeys.Environment] = metadata.GetEnv()
	}
	return labels
}
