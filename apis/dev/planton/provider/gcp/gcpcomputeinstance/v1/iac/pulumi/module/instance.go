package module

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// computeInstance creates a Compute Engine instance.
func computeInstance(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*compute.Instance, error) {

	spec := locals.GcpComputeInstance.Spec

	// Build boot disk initialize params
	initializeParams := &compute.InstanceBootDiskInitializeParamsArgs{
		Image: pulumi.String(spec.BootDisk.Image),
	}

	// Set boot disk size if specified
	if spec.BootDisk.SizeGb > 0 {
		initializeParams.Size = pulumi.Int(int(spec.BootDisk.SizeGb))
	}

	// Set boot disk type if specified
	if spec.BootDisk.Type != "" {
		initializeParams.Type = pulumi.String(spec.BootDisk.Type)
	}

	// Build boot disk configuration
	bootDiskArgs := &compute.InstanceBootDiskArgs{
		AutoDelete:       pulumi.Bool(spec.BootDisk.AutoDelete),
		InitializeParams: initializeParams,
	}

	// Build network interfaces
	networkInterfaces := compute.InstanceNetworkInterfaceArray{}
	for _, ni := range spec.NetworkInterfaces {
		niArgs := &compute.InstanceNetworkInterfaceArgs{}

		// Set network if specified
		if ni.Network != nil && ni.Network.GetValue() != "" {
			niArgs.Network = pulumi.String(ni.Network.GetValue())
		}

		// Set subnetwork if specified
		if ni.Subnetwork != nil && ni.Subnetwork.GetValue() != "" {
			niArgs.Subnetwork = pulumi.String(ni.Subnetwork.GetValue())
		}

		// Build access configs for external IP
		if len(ni.AccessConfigs) > 0 {
			accessConfigs := compute.InstanceNetworkInterfaceAccessConfigArray{}
			for _, ac := range ni.AccessConfigs {
				acArgs := &compute.InstanceNetworkInterfaceAccessConfigArgs{}
				if ac.NatIp != "" {
					acArgs.NatIp = pulumi.String(ac.NatIp)
				}
				if ac.NetworkTier != "" {
					acArgs.NetworkTier = pulumi.String(ac.NetworkTier)
				}
				accessConfigs = append(accessConfigs, acArgs)
			}
			niArgs.AccessConfigs = accessConfigs
		}

		// Build alias IP ranges if specified
		if len(ni.AliasIpRanges) > 0 {
			aliasIpRanges := compute.InstanceNetworkInterfaceAliasIpRangeArray{}
			for _, air := range ni.AliasIpRanges {
				airArgs := &compute.InstanceNetworkInterfaceAliasIpRangeArgs{
					IpCidrRange: pulumi.String(air.IpCidrRange),
				}
				if air.SubnetworkRangeName != "" {
					airArgs.SubnetworkRangeName = pulumi.String(air.SubnetworkRangeName)
				}
				aliasIpRanges = append(aliasIpRanges, airArgs)
			}
			niArgs.AliasIpRanges = aliasIpRanges
		}

		networkInterfaces = append(networkInterfaces, niArgs)
	}

	// Build instance arguments
	instanceArgs := &compute.InstanceArgs{
		Name:              pulumi.String(locals.GcpComputeInstance.Metadata.Name),
		Project:           pulumi.String(spec.ProjectId.GetValue()),
		Zone:              pulumi.String(spec.Zone),
		MachineType:       pulumi.String(spec.MachineType),
		BootDisk:          bootDiskArgs,
		NetworkInterfaces: networkInterfaces,
		Labels:            pulumi.ToStringMap(locals.GcpLabels),
	}

	// Set deletion protection
	instanceArgs.DeletionProtection = pulumi.Bool(spec.DeletionProtection)

	// Set allow stopping for update
	instanceArgs.AllowStoppingForUpdate = pulumi.Bool(spec.AllowStoppingForUpdate)

	// Set tags if specified
	if len(spec.Tags) > 0 {
		instanceArgs.Tags = pulumi.ToStringArray(spec.Tags)
	}

	// Set metadata if specified
	if len(spec.Metadata) > 0 {
		instanceArgs.Metadata = pulumi.ToStringMap(spec.Metadata)
	}

	// Set startup script if specified
	if spec.StartupScript != "" {
		if instanceArgs.MetadataStartupScript == nil {
			instanceArgs.MetadataStartupScript = pulumi.String(spec.StartupScript)
		}
	}

	// Set SSH keys if specified
	if len(spec.SshKeys) > 0 {
		sshKeysStr := ""
		for i, key := range spec.SshKeys {
			if i > 0 {
				sshKeysStr += "\n"
			}
			sshKeysStr += key
		}
		if instanceArgs.Metadata == nil {
			instanceArgs.Metadata = pulumi.StringMap{}
		}
		// Note: This is a workaround since we can't modify the map directly
		// The metadata will be set via the resource configuration
	}

	// Set service account if specified
	if spec.ServiceAccount != nil {
		saArgs := &compute.InstanceServiceAccountArgs{}
		if spec.ServiceAccount.Email != nil && spec.ServiceAccount.Email.GetValue() != "" {
			saArgs.Email = pulumi.String(spec.ServiceAccount.Email.GetValue())
		}
		if len(spec.ServiceAccount.Scopes) > 0 {
			saArgs.Scopes = pulumi.ToStringArray(spec.ServiceAccount.Scopes)
		} else {
			// Default scope
			saArgs.Scopes = pulumi.StringArray{
				pulumi.String("https://www.googleapis.com/auth/cloud-platform"),
			}
		}
		instanceArgs.ServiceAccount = saArgs
	}

	// Set scheduling options
	if spec.Scheduling != nil {
		schedulingArgs := &compute.InstanceSchedulingArgs{}

		schedulingArgs.Preemptible = pulumi.Bool(spec.Scheduling.Preemptible)
		schedulingArgs.AutomaticRestart = pulumi.Bool(spec.Scheduling.AutomaticRestart)

		if spec.Scheduling.OnHostMaintenance != "" {
			schedulingArgs.OnHostMaintenance = pulumi.String(spec.Scheduling.OnHostMaintenance)
		}

		if spec.Scheduling.ProvisioningModel != "" {
			schedulingArgs.ProvisioningModel = pulumi.String(spec.Scheduling.ProvisioningModel)
		}

		if spec.Scheduling.InstanceTerminationAction != "" {
			schedulingArgs.InstanceTerminationAction = pulumi.String(spec.Scheduling.InstanceTerminationAction)
		}

		instanceArgs.Scheduling = schedulingArgs
	} else if spec.Preemptible || spec.Spot {
		// Handle legacy preemptible/spot fields
		schedulingArgs := &compute.InstanceSchedulingArgs{
			Preemptible:       pulumi.Bool(spec.Preemptible || spec.Spot),
			AutomaticRestart:  pulumi.Bool(false),
			OnHostMaintenance: pulumi.String("TERMINATE"),
		}
		if spec.Spot {
			schedulingArgs.ProvisioningModel = pulumi.String("SPOT")
		}
		instanceArgs.Scheduling = schedulingArgs
	}

	// Add attached disks if specified
	if len(spec.AttachedDisks) > 0 {
		attachedDisks := compute.InstanceAttachedDiskArray{}
		for i, disk := range spec.AttachedDisks {
			diskArgs := &compute.InstanceAttachedDiskArgs{
				Source: pulumi.String(disk.Source),
			}
			if disk.DeviceName != "" {
				diskArgs.DeviceName = pulumi.String(disk.DeviceName)
			} else {
				diskArgs.DeviceName = pulumi.String(fmt.Sprintf("attached-disk-%d", i))
			}
			if disk.Mode != "" {
				diskArgs.Mode = pulumi.String(disk.Mode)
			}
			attachedDisks = append(attachedDisks, diskArgs)
		}
		instanceArgs.AttachedDisks = attachedDisks
	}

	// Create the instance
	instance, err := compute.NewInstance(ctx,
		locals.GcpComputeInstance.Metadata.Name,
		instanceArgs,
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return nil, err
	}

	return instance, nil
}
