package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackkeypairv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackkeypair/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackKeypair        *openstackkeypairv1.OpenStackKeypair
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackkeypairv1.OpenStackKeypairStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackKeypair = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	return locals
}
