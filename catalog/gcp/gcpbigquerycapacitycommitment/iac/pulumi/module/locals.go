package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpbigquerycapacitycommitmentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigquerycapacitycommitment/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig             *gcpprovider.GcpProviderConfig
	GcpBigQueryCapacityCommitment *gcpbigquerycapacitycommitmentv1alpha1.GcpBigQueryCapacityCommitment

	// CapacityCommitmentId is spec.capacity_commitment_id when set,
	// otherwise metadata.name -- identical to the Terraform module's
	// locals.capacity_commitment_id.
	CapacityCommitmentId string
}

// initializeLocals derives the defaulted commitment id. Commitments carry
// no labels, so there is no attribution label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpbigquerycapacitycommitmentv1alpha1.GcpBigQueryCapacityCommitmentStackInput) *Locals {
	locals := &Locals{}
	locals.GcpBigQueryCapacityCommitment = stackInput.Target

	locals.CapacityCommitmentId = locals.GcpBigQueryCapacityCommitment.Spec.CapacityCommitmentId
	if locals.CapacityCommitmentId == "" {
		locals.CapacityCommitmentId = locals.GcpBigQueryCapacityCommitment.Metadata.Name
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
