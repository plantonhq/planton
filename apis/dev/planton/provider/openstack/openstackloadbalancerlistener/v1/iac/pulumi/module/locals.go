package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstackloadbalancerlistenerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstackloadbalancerlistener/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig       *openstackprovider.OpenStackProviderConfig
	OpenStackLoadBalancerListener *openstackloadbalancerlistenerv1.OpenStackLoadBalancerListener
	// LoadBalancerId is the resolved load balancer ID from the StringValueOrRef.
	LoadBalancerId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackloadbalancerlistenerv1.OpenStackLoadBalancerListenerStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackLoadBalancerListener = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract loadbalancer_id from StringValueOrRef.
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.LoadBalancerId = stackInput.Target.Spec.LoadbalancerId.GetValue()

	return locals
}
