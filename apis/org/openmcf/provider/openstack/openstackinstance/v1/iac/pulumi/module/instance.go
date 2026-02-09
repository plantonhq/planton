package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// instance provisions the OpenStack Compute instance and exports outputs.
func instance(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackInstance.Spec
	resourceName := locals.OpenStackInstance.Metadata.Name

	instanceArgs := &compute.InstanceArgs{
		Name: pulumi.String(resourceName),
	}

	// Flavor (exactly one of flavor_name or flavor_id is set, enforced by proto validation).
	if spec.FlavorName != "" {
		instanceArgs.FlavorName = pulumi.StringPtr(spec.FlavorName)
	}
	if spec.FlavorId != "" {
		instanceArgs.FlavorId = pulumi.StringPtr(spec.FlavorId)
	}

	// Image (optional -- not needed when booting from block_device).
	if spec.ImageName != "" {
		instanceArgs.ImageName = pulumi.StringPtr(spec.ImageName)
	}
	if spec.ImageId != "" {
		instanceArgs.ImageId = pulumi.StringPtr(spec.ImageId)
	}

	// Keypair (optional).
	if locals.KeyPair != "" {
		instanceArgs.KeyPair = pulumi.StringPtr(locals.KeyPair)
	}

	// Network attachments (at least one, enforced by proto validation).
	var networks compute.InstanceNetworkArray
	for _, net := range locals.NetworkAttachments {
		networkArgs := compute.InstanceNetworkArgs{
			AccessNetwork: pulumi.Bool(net.AccessNetwork),
		}
		if net.Uuid != "" {
			networkArgs.Uuid = pulumi.String(net.Uuid)
		}
		if net.Port != "" {
			networkArgs.Port = pulumi.String(net.Port)
		}
		if net.FixedIpV4 != "" {
			networkArgs.FixedIpV4 = pulumi.StringPtr(net.FixedIpV4)
		}
		networks = append(networks, networkArgs)
	}
	instanceArgs.Networks = networks

	// Security groups (optional, resolved names).
	if len(locals.SecurityGroups) > 0 {
		var sgArray pulumi.StringArray
		for _, sg := range locals.SecurityGroups {
			sgArray = append(sgArray, pulumi.String(sg))
		}
		instanceArgs.SecurityGroups = sgArray
	}

	// Block device mappings (optional).
	if len(spec.BlockDevice) > 0 {
		var blockDevices compute.InstanceBlockDeviceArray
		for _, bd := range spec.BlockDevice {
			bdArgs := compute.InstanceBlockDeviceArgs{
				SourceType:          pulumi.String(bd.SourceType),
				BootIndex:           pulumi.Int(int(bd.BootIndex)),
				DeleteOnTermination: pulumi.Bool(bd.DeleteOnTermination),
			}
			if bd.Uuid != "" {
				bdArgs.Uuid = pulumi.StringPtr(bd.Uuid)
			}
			if bd.DestinationType != "" {
				bdArgs.DestinationType = pulumi.StringPtr(bd.DestinationType)
			}
			if bd.VolumeSize > 0 {
				bdArgs.VolumeSize = pulumi.IntPtr(int(bd.VolumeSize))
			}
			if bd.VolumeType != "" {
				bdArgs.VolumeType = pulumi.StringPtr(bd.VolumeType)
			}
			blockDevices = append(blockDevices, bdArgs)
		}
		instanceArgs.BlockDevices = blockDevices
	}

	// User data (optional, ForceNew).
	if spec.UserData != "" {
		instanceArgs.UserData = pulumi.StringPtr(spec.UserData)
	}

	// Metadata (optional).
	if len(spec.Metadata) > 0 {
		metadataMap := pulumi.StringMap{}
		for k, v := range spec.Metadata {
			metadataMap[k] = pulumi.String(v)
		}
		instanceArgs.Metadata = metadataMap
	}

	// Config drive (optional).
	if spec.ConfigDrive != nil {
		instanceArgs.ConfigDrive = pulumi.BoolPtr(*spec.ConfigDrive)
	}

	// Server group (via scheduler_hints).
	if locals.ServerGroupId != "" {
		instanceArgs.SchedulerHints = compute.InstanceSchedulerHintArray{
			compute.InstanceSchedulerHintArgs{
				Group: pulumi.String(locals.ServerGroupId),
			},
		}
	}

	// Availability zone (optional, ForceNew).
	if spec.AvailabilityZone != "" {
		instanceArgs.AvailabilityZone = pulumi.StringPtr(spec.AvailabilityZone)
	}

	// Tags (optional).
	if len(spec.Tags) > 0 {
		var tagArray pulumi.StringArray
		for _, tag := range spec.Tags {
			tagArray = append(tagArray, pulumi.String(tag))
		}
		instanceArgs.Tags = tagArray
	}

	// Region override (optional).
	if spec.Region != "" {
		instanceArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdInstance, err := compute.NewInstance(
		ctx,
		strings.ToLower(resourceName),
		instanceArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack instance")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpName, createdInstance.Name)
	ctx.Export(OpAccessIpV4, createdInstance.AccessIpV4)
	ctx.Export(OpAccessIpV6, createdInstance.AccessIpV6)
	ctx.Export(OpRegion, createdInstance.Region)

	return nil
}
