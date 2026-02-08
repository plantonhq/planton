package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackcomputekeypairv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackcomputekeypair/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackComputeKeypair *openstackcomputekeypairv1.OpenStackComputeKeypair
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackcomputekeypairv1.OpenStackComputeKeypairStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackComputeKeypair = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	return locals
}
