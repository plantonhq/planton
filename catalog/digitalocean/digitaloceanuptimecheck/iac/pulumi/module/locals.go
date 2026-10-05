package module

import (
	digitaloceanuptimecheckv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanuptimecheck/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. An uptime
// check has no tag surface, so no Planton label set applies here.
type Locals struct {
	DigitalOceanUptimeCheck *digitaloceanuptimecheckv1alpha1.DigitalOceanUptimeCheck
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceanuptimecheckv1alpha1.DigitalOceanUptimeCheckIacInput) *Locals {
	return &Locals{
		DigitalOceanUptimeCheck: iacInput.Target,
	}
}
