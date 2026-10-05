package module

import (
	digitaloceandatabaseconnectionpoolv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabaseconnectionpool/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. The pool
// resource has no tag surface, so no Planton label set applies here.
type Locals struct {
	DigitalOceanDatabaseConnectionPool *digitaloceandatabaseconnectionpoolv1alpha1.DigitalOceanDatabaseConnectionPool
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceandatabaseconnectionpoolv1alpha1.DigitalOceanDatabaseConnectionPoolIacInput) *Locals {
	return &Locals{
		DigitalOceanDatabaseConnectionPool: iacInput.Target,
	}
}
