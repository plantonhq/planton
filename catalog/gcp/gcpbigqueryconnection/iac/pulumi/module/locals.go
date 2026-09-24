package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpbigqueryconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigqueryconnection/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig     *gcpprovider.GcpProviderConfig
	GcpBigQueryConnection *gcpbigqueryconnectionv1alpha1.GcpBigQueryConnection

	// ConnectionId is spec.connection_id when set, otherwise metadata.name
	// -- identical to the Terraform module's locals.connection_id.
	ConnectionId string
}

// initializeLocals derives the defaulted connection id. Connections carry
// no labels, so there is no attribution label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpbigqueryconnectionv1alpha1.GcpBigQueryConnectionStackInput) *Locals {
	locals := &Locals{}
	locals.GcpBigQueryConnection = stackInput.Target

	locals.ConnectionId = locals.GcpBigQueryConnection.Spec.ConnectionId
	if locals.ConnectionId == "" {
		locals.ConnectionId = locals.GcpBigQueryConnection.Metadata.Name
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
