package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstacksecuritygroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstacksecuritygroup/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackSecurityGroup  *openstacksecuritygroupv1.OpenStackSecurityGroup
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstacksecuritygroupv1.OpenStackSecurityGroupStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackSecurityGroup = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	return locals
}
