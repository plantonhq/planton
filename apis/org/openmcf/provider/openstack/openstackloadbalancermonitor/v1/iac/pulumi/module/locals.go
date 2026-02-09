package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackloadbalancermonitorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackloadbalancermonitor/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig       *openstackprovider.OpenStackProviderConfig
	OpenStackLoadBalancerMonitor  *openstackloadbalancermonitorv1.OpenStackLoadBalancerMonitor
	// PoolId is the resolved pool ID from the StringValueOrRef.
	PoolId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackloadbalancermonitorv1.OpenStackLoadBalancerMonitorStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackLoadBalancerMonitor = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract pool_id from StringValueOrRef.
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.PoolId = stackInput.Target.Spec.PoolId.GetValue()

	return locals
}
