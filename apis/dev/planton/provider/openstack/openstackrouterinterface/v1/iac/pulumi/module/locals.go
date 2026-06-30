package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstackrouterinterfacev1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstackrouterinterface/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig  *openstackprovider.OpenStackProviderConfig
	OpenStackRouterInterface *openstackrouterinterfacev1.OpenStackRouterInterface
	// RouterId is the resolved router ID from the StringValueOrRef.
	RouterId string
	// SubnetId is the resolved subnet ID from the StringValueOrRef.
	SubnetId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackrouterinterfacev1.OpenStackRouterInterfaceStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackRouterInterface = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract router_id from StringValueOrRef (required field).
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.RouterId = stackInput.Target.Spec.RouterId.GetValue()

	// Extract subnet_id from StringValueOrRef (required field).
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.SubnetId = stackInput.Target.Spec.SubnetId.GetValue()

	return locals
}
