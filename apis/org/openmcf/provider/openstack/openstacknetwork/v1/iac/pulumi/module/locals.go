package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstacknetworkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstacknetwork/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackNetwork        *openstacknetworkv1.OpenStackNetwork
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstacknetworkv1.OpenStackNetworkStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackNetwork = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	return locals
}
