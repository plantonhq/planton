package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackloadbalancerpoolv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackloadbalancerpool/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig  *openstackprovider.OpenStackProviderConfig
	OpenStackLoadBalancerPool *openstackloadbalancerpoolv1.OpenStackLoadBalancerPool
	// ListenerId is the resolved listener ID from the StringValueOrRef.
	ListenerId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackloadbalancerpoolv1.OpenStackLoadBalancerPoolStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackLoadBalancerPool = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract listener_id from StringValueOrRef.
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.ListenerId = stackInput.Target.Spec.ListenerId.GetValue()

	return locals
}
