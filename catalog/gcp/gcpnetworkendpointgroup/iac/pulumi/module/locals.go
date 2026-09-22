package module

import (
	gcpnetworkendpointgroupv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpnetworkendpointgroup/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derivations both engines share -- the group name and
// which of the two provider collections this instance is.
type Locals struct {
	GcpNetworkEndpointGroup *gcpnetworkendpointgroupv1alpha1.GcpNetworkEndpointGroup

	// ProjectId is empty when the manifest omits it — the provider's default
	// project then applies (the same ambient contract the Terraform module
	// honors by passing null).
	ProjectId string

	// NegName is the spec's neg_name, or metadata.name when the spec leaves
	// it empty — the same naming basis every kind uses.
	NegName string

	// IsZonal is the scope selector: an empty spec.zone builds the global
	// internet network endpoint group, a zone name builds the zonal one.
	IsZonal bool
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpnetworkendpointgroupv1alpha1.GcpNetworkEndpointGroupStackInput) *Locals {
	target := stackInput.Target

	negName := target.Spec.NegName
	if negName == "" {
		negName = target.Metadata.Name
	}

	return &Locals{
		GcpNetworkEndpointGroup: target,
		ProjectId:               target.Spec.ProjectId.GetValue(),
		NegName:                 negName,
		IsZonal:                 target.Spec.Zone != "",
	}
}
