package module

import (
	digitaloceandatabaseuserv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabaseuser/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. The database
// user resource has no tag surface, so no Planton label set applies here.
type Locals struct {
	DigitalOceanDatabaseUser *digitaloceandatabaseuserv1alpha1.DigitalOceanDatabaseUser
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceandatabaseuserv1alpha1.DigitalOceanDatabaseUserIacInput) *Locals {
	return &Locals{
		DigitalOceanDatabaseUser: iacInput.Target,
	}
}
