package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackdnszonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackdnszone/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackDnsZone        *openstackdnszonev1.OpenStackDnsZone
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackdnszonev1.OpenStackDnsZoneStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackDnsZone = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	return locals
}
