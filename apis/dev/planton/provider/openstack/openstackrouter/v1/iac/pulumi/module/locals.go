package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstackrouterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstackrouter/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackRouter         *openstackrouterv1.OpenStackRouter
	// ExternalNetworkId is the resolved external network ID from the StringValueOrRef.
	// Empty if no external gateway is configured.
	ExternalNetworkId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackrouterv1.OpenStackRouterStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackRouter = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract external_network_id from StringValueOrRef if present.
	// This field is optional -- routers can exist without an external gateway.
	// At runtime, the value is resolved by the FK resolver middleware.
	if stackInput.Target.Spec.ExternalNetworkId != nil {
		locals.ExternalNetworkId = stackInput.Target.Spec.ExternalNetworkId.GetValue()
	}

	return locals
}
