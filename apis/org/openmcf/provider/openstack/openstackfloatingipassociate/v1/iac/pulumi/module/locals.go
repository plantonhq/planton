package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackfloatingipassociatev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackfloatingipassociate/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig      *openstackprovider.OpenStackProviderConfig
	OpenStackFloatingIpAssociate *openstackfloatingipassociatev1.OpenStackFloatingIpAssociate
	// FloatingIp is the resolved floating IP address from the StringValueOrRef.
	// This targets OpenStackFloatingIp.status.outputs.address (the IP address,
	// not the UUID) because the TF resource takes an address or ID.
	FloatingIp string
	// PortId is the resolved port ID from the StringValueOrRef.
	PortId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackfloatingipassociatev1.OpenStackFloatingIpAssociateStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackFloatingIpAssociate = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract floating_ip from StringValueOrRef (required field).
	// At runtime, the value is resolved by the FK resolver middleware.
	// This is typically an IP address like "203.0.113.42" (not a UUID).
	locals.FloatingIp = stackInput.Target.Spec.FloatingIp.GetValue()

	// Extract port_id from StringValueOrRef (required field).
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.PortId = stackInput.Target.Spec.PortId.GetValue()

	return locals
}
