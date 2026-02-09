package module

import (
	openstackprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	openstackinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackinstance/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	OpenStackProviderConfig *openstackprovider.OpenStackProviderConfig
	OpenStackInstance       *openstackinstancev1.OpenStackInstance
	// KeyPair is the resolved keypair name from the StringValueOrRef.
	KeyPair string
	// ServerGroupId is the resolved server group UUID from the StringValueOrRef.
	ServerGroupId string
	// SecurityGroups are the resolved security group names from the repeated StringValueOrRef.
	SecurityGroups []string
	// NetworkAttachments holds resolved network/port UUIDs for each network entry.
	NetworkAttachments []NetworkAttachment
}

// NetworkAttachment holds the resolved values for a single network entry.
type NetworkAttachment struct {
	// Uuid is the resolved network UUID (empty if using port mode).
	Uuid string
	// Port is the resolved port UUID (empty if using network mode).
	Port string
	// FixedIpV4 is the requested fixed IPv4 address.
	FixedIpV4 string
	// AccessNetwork marks this as the instance's access network.
	AccessNetwork bool
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *openstackinstancev1.OpenStackInstanceStackInput) *Locals {
	locals := &Locals{}

	locals.OpenStackInstance = stackInput.Target
	locals.OpenStackProviderConfig = stackInput.ProviderConfig

	spec := stackInput.Target.Spec

	// Extract key_pair from StringValueOrRef (optional).
	if spec.KeyPair != nil {
		locals.KeyPair = spec.KeyPair.GetValue()
	}

	// Extract server_group_id from StringValueOrRef (optional).
	if spec.ServerGroupId != nil {
		locals.ServerGroupId = spec.ServerGroupId.GetValue()
	}

	// Extract security_groups from repeated StringValueOrRef.
	for _, sg := range spec.SecurityGroups {
		locals.SecurityGroups = append(locals.SecurityGroups, sg.GetValue())
	}

	// Extract network attachments, resolving FKs.
	for _, net := range spec.Networks {
		attachment := NetworkAttachment{
			FixedIpV4:     net.FixedIpV4,
			AccessNetwork: net.AccessNetwork,
		}
		if net.Uuid != nil {
			attachment.Uuid = net.Uuid.GetValue()
		}
		if net.Port != nil {
			attachment.Port = net.Port.GetValue()
		}
		locals.NetworkAttachments = append(locals.NetworkAttachments, attachment)
	}

	return locals
}
