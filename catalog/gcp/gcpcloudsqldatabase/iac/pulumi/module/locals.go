package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudsqldatabasev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudsqldatabase/v1alpha1"
)

// Locals holds handy references used across this module.
//
// No label map here: google_sql_database has no labels surface in the API,
// so there is nothing to stamp — attribution lives on the instance node.
type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpCloudSqlDatabase *gcpcloudsqldatabasev1alpha1.GcpCloudSqlDatabase
}

// initializeLocals fills the Locals struct from the incoming IaC input.
func initializeLocals(iacInput *gcpcloudsqldatabasev1alpha1.GcpCloudSqlDatabaseIacInput) *Locals {
	return &Locals{
		GcpCloudSqlDatabase: iacInput.Target,
		GcpProviderConfig:   iacInput.ProviderConfig,
	}
}
