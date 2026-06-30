package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstackfloatingipv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstackfloatingip/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackFloatingIp     *openstackfloatingipv1.OpenStackFloatingIp
	// FloatingNetworkId is the resolved external network ID from the StringValueOrRef.
	FloatingNetworkId string
	// PortId is the resolved port ID from the StringValueOrRef.
	// Empty if no port association is configured (allocation-only mode).
	PortId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackfloatingipv1.OpenStackFloatingIpStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackFloatingIp = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract floating_network_id from StringValueOrRef (required).
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.FloatingNetworkId = stackInput.Target.Spec.FloatingNetworkId.GetValue()

	// Extract port_id from StringValueOrRef if present (optional).
	// Empty when no port association is configured.
	if stackInput.Target.Spec.PortId != nil {
		locals.PortId = stackInput.Target.Spec.PortId.GetValue()
	}

	return locals
}
