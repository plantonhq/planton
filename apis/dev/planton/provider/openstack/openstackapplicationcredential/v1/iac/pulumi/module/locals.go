package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstackappcredv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstackapplicationcredential/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig        *openstackprovider.OpenStackProviderConfig
	OpenStackApplicationCredential *openstackappcredv1.OpenStackApplicationCredential
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackappcredv1.OpenStackApplicationCredentialStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackApplicationCredential = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	return locals
}
