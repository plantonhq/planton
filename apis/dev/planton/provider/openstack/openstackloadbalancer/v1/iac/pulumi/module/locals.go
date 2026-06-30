package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstackloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstackloadbalancer/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackLoadBalancer   *openstackloadbalancerv1.OpenStackLoadBalancer
	// VipSubnetId is the resolved subnet ID from the StringValueOrRef.
	VipSubnetId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackloadbalancerv1.OpenStackLoadBalancerStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackLoadBalancer = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract vip_subnet_id from StringValueOrRef.
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.VipSubnetId = stackInput.Target.Spec.VipSubnetId.GetValue()

	return locals
}
