package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpbigqueryreservationgroupv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigqueryreservationgroup/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig           *gcpprovider.GcpProviderConfig
	GcpBigQueryReservationGroup *gcpbigqueryreservationgroupv1alpha1.GcpBigQueryReservationGroup

	// ReservationGroupName is spec.reservation_group_name when set,
	// otherwise metadata.name -- identical to the Terraform module's
	// locals.reservation_group_name.
	ReservationGroupName string
}

// initializeLocals derives the defaulted group name. Groups carry no
// labels, so there is no attribution label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpbigqueryreservationgroupv1alpha1.GcpBigQueryReservationGroupStackInput) *Locals {
	locals := &Locals{}
	locals.GcpBigQueryReservationGroup = stackInput.Target

	locals.ReservationGroupName = locals.GcpBigQueryReservationGroup.Spec.ReservationGroupName
	if locals.ReservationGroupName == "" {
		locals.ReservationGroupName = locals.GcpBigQueryReservationGroup.Metadata.Name
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
