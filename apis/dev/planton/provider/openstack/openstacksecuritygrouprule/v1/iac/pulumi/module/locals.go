package module

import (
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
	openstacksecuritygrouprulev1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstacksecuritygrouprule/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig    *openstackprovider.OpenStackProviderConfig
	OpenStackSecurityGroupRule *openstacksecuritygrouprulev1.OpenStackSecurityGroupRule
	// SecurityGroupId is the resolved security group ID from the required StringValueOrRef.
	SecurityGroupId string
	// RemoteGroupId is the resolved remote security group ID from the optional StringValueOrRef.
	// Empty string if the optional FK was not set.
	RemoteGroupId string
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstacksecuritygrouprulev1.OpenStackSecurityGroupRuleStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackSecurityGroupRule = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	// Extract security_group_id from StringValueOrRef (required field).
	// At runtime, the value is resolved by the FK resolver middleware.
	locals.SecurityGroupId = stackInput.Target.Spec.SecurityGroupId.GetValue()

	// Extract remote_group_id from StringValueOrRef (optional field).
	// Returns empty string if the FK was not set.
	if stackInput.Target.Spec.RemoteGroupId != nil {
		locals.RemoteGroupId = stackInput.Target.Spec.RemoteGroupId.GetValue()
	}

	return locals
}
