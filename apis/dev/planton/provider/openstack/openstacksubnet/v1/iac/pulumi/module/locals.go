package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstacksubnetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstacksubnet/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackSubnet         *openstacksubnetv1.OpenStackSubnet
	// NetworkId is the resolved network ID from the StringValueOrRef.
	NetworkId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstacksubnetv1.OpenStackSubnetStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackSubnet = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract network_id from StringValueOrRef.
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.NetworkId = stackInput.Target.Spec.NetworkId.GetValue()

	return locals
}
