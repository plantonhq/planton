package module

import (
	digitaloceandatabasefirewallv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabasefirewall/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. The firewall
// resource has no tag-application surface (its "tag" rules TARGET tags,
// they don't apply them), so no Planton label set applies here.
type Locals struct {
	DigitalOceanDatabaseFirewall *digitaloceandatabasefirewallv1alpha1.DigitalOceanDatabaseFirewall
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceandatabasefirewallv1alpha1.DigitalOceanDatabaseFirewallIacInput) *Locals {
	return &Locals{
		DigitalOceanDatabaseFirewall: iacInput.Target,
	}
}
