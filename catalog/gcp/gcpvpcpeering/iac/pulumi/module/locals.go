package module

import (
	gcpvpcpeeringv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvpcpeering/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the two derivations both engines share -- the peering name
// and which of the two provider resources this instance is.
type Locals struct {
	GcpVpcPeering *gcpvpcpeeringv1alpha1.GcpVpcPeering

	// PeeringName is the spec's peering_name, or metadata.name when the
	// spec leaves it empty -- the same naming basis every kind uses. In the
	// routes-config form it names the EXISTING peering.
	PeeringName string

	// IsCreateForm is true when peer_network is set: this side creates the
	// peering (google_compute_network_peering). False selects the
	// routes-config form (google_compute_network_peering_routes_config) on
	// a peering someone else created.
	IsCreateForm bool
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpvpcpeeringv1alpha1.GcpVpcPeeringStackInput) *Locals {
	target := stackInput.Target

	peeringName := target.Spec.PeeringName
	if peeringName == "" {
		peeringName = target.Metadata.Name
	}

	return &Locals{
		GcpVpcPeering: target,
		PeeringName:   peeringName,
		IsCreateForm:  target.Spec.PeerNetwork.GetValue() != "",
	}
}
