package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstacknetworkportv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstacknetworkport/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackNetworkPort    *openstacknetworkportv1.OpenStackNetworkPort
	// NetworkId is the resolved network ID from the StringValueOrRef.
	NetworkId string
	// SecurityGroupIds is the resolved list of security group IDs from
	// the repeated StringValueOrRef field. Each element was independently
	// resolved by the FK resolver middleware.
	SecurityGroupIds []string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstacknetworkportv1.OpenStackNetworkPortStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackNetworkPort = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract network_id from StringValueOrRef (required field).
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.NetworkId = stackInput.Target.Spec.NetworkId.GetValue()

	// Extract security_group_ids from repeated StringValueOrRef.
	// Each element is independently resolved by the FK resolver middleware.
	for _, sgRef := range stackInput.Target.Spec.SecurityGroupIds {
		locals.SecurityGroupIds = append(locals.SecurityGroupIds, sgRef.GetValue())
	}

	return locals
}
