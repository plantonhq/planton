package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstackvolumev1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstackvolume/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackVolume         *openstackvolumev1.OpenStackVolume
	// ImageId is the resolved image ID from the optional StringValueOrRef.
	// Empty string when image_id is not set in the spec.
	ImageId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackvolumev1.OpenStackVolumeStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackVolume = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract image_id from StringValueOrRef (optional field).
	// At runtime, the value is resolved by the FK resolver middleware.
	if stackInput.Target.Spec.ImageId != nil {
		locals.ImageId = stackInput.Target.Spec.ImageId.GetValue()
	}

	return locals
}
