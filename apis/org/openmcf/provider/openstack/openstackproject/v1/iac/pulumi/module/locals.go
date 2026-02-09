package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackprojectv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackproject/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackProject        *openstackprojectv1.OpenStackProject
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackprojectv1.OpenStackProjectStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackProject = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	return locals
}
