package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudsqluserv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudsqluser/v1alpha1"
)

// Locals holds handy references used across this module.
//
// No label map here: google_sql_user has no labels surface in the API, so
// there is nothing to stamp — attribution lives on the instance node.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpCloudSqlUser   *gcpcloudsqluserv1alpha1.GcpCloudSqlUser
}

// initializeLocals fills the Locals struct from the incoming IaC input.
func initializeLocals(iacInput *gcpcloudsqluserv1alpha1.GcpCloudSqlUserIacInput) *Locals {
	return &Locals{
		GcpCloudSqlUser:   iacInput.Target,
		GcpProviderConfig: iacInput.ProviderConfig,
	}
}
