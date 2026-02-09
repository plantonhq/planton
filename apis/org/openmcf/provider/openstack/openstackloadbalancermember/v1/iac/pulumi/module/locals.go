package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackloadbalancermemberv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackloadbalancermember/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig      *openstackprovider.OpenStackProviderConfig
	OpenStackLoadBalancerMember  *openstackloadbalancermemberv1.OpenStackLoadBalancerMember
	// PoolId is the resolved pool ID from the StringValueOrRef.
	PoolId string
	// SubnetId is the optional resolved subnet ID from the StringValueOrRef.
	SubnetId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackloadbalancermemberv1.OpenStackLoadBalancerMemberStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackLoadBalancerMember = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract pool_id from StringValueOrRef.
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.PoolId = stackInput.Target.Spec.PoolId.GetValue()

	// Extract optional subnet_id from StringValueOrRef.
	if stackInput.Target.Spec.SubnetId != nil {
		locals.SubnetId = stackInput.Target.Spec.SubnetId.GetValue()
	}

	return locals
}
