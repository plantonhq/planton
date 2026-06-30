package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/workbench"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func workbenchInstance(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiNotebook.Spec

	args := &workbench.InstanceArgs{
		Location:   pulumi.String(spec.Location),
		Project:    pulumi.StringPtr(spec.ProjectId.GetValue()),
		InstanceId: pulumi.StringPtr(locals.InstanceName),
		Labels:     pulumi.ToStringMap(locals.GcpLabels),
	}

	// Desired state (ACTIVE or STOPPED).
	if spec.DesiredState != "" {
		args.DesiredState = pulumi.StringPtr(spec.DesiredState)
	}

	// Disable proxy access.
	if spec.DisableProxyAccess {
		args.DisableProxyAccess = pulumi.BoolPtr(true)
	}

	// Instance owners.
	if len(spec.InstanceOwners) > 0 {
		args.InstanceOwners = pulumi.ToStringArray(spec.InstanceOwners)
	}

	// Build the GCE setup block from our flattened spec fields.
	gceSetup := &workbench.InstanceGceSetupArgs{}
	hasGceSetup := false

	// Machine type (required in our spec).
	gceSetup.MachineType = pulumi.StringPtr(spec.MachineType)
	hasGceSetup = true

	// Boot disk.
	if spec.BootDisk != nil {
		bootDiskArgs := &workbench.InstanceGceSetupBootDiskArgs{}
		if spec.BootDisk.DiskType != "" {
			bootDiskArgs.DiskType = pulumi.StringPtr(spec.BootDisk.DiskType)
		}
		if spec.BootDisk.DiskSizeGb != 0 {
			bootDiskArgs.DiskSizeGb = pulumi.StringPtr(fmt.Sprintf("%d", spec.BootDisk.DiskSizeGb))
		}
		if spec.BootDisk.KmsKey != nil && spec.BootDisk.KmsKey.GetValue() != "" {
			bootDiskArgs.DiskEncryption = pulumi.StringPtr("CMEK")
			bootDiskArgs.KmsKey = pulumi.StringPtr(spec.BootDisk.KmsKey.GetValue())
		}
		gceSetup.BootDisk = bootDiskArgs
	}

	// Data disk.
	if spec.DataDisk != nil {
		dataDiskArgs := &workbench.InstanceGceSetupDataDisksArgs{}
		if spec.DataDisk.DiskType != "" {
			dataDiskArgs.DiskType = pulumi.StringPtr(spec.DataDisk.DiskType)
		}
		if spec.DataDisk.DiskSizeGb != 0 {
			dataDiskArgs.DiskSizeGb = pulumi.StringPtr(fmt.Sprintf("%d", spec.DataDisk.DiskSizeGb))
		}
		if spec.DataDisk.KmsKey != nil && spec.DataDisk.KmsKey.GetValue() != "" {
			dataDiskArgs.DiskEncryption = pulumi.StringPtr("CMEK")
			dataDiskArgs.KmsKey = pulumi.StringPtr(spec.DataDisk.KmsKey.GetValue())
		}
		gceSetup.DataDisks = dataDiskArgs
	}

	// Accelerator config.
	if spec.AcceleratorConfig != nil && spec.AcceleratorConfig.Type != "" {
		accelArgs := workbench.InstanceGceSetupAcceleratorConfigArgs{}
		accelArgs.Type = pulumi.StringPtr(spec.AcceleratorConfig.Type)
		if spec.AcceleratorConfig.CoreCount != 0 {
			accelArgs.CoreCount = pulumi.StringPtr(fmt.Sprintf("%d", spec.AcceleratorConfig.CoreCount))
		}
		gceSetup.AcceleratorConfigs = workbench.InstanceGceSetupAcceleratorConfigArray{accelArgs}
	}

	// Network interface.
	if spec.NetworkInterface != nil {
		niArgs := workbench.InstanceGceSetupNetworkInterfaceArgs{}
		if spec.NetworkInterface.Network != nil && spec.NetworkInterface.Network.GetValue() != "" {
			niArgs.Network = pulumi.StringPtr(spec.NetworkInterface.Network.GetValue())
		}
		if spec.NetworkInterface.Subnet != nil && spec.NetworkInterface.Subnet.GetValue() != "" {
			niArgs.Subnet = pulumi.StringPtr(spec.NetworkInterface.Subnet.GetValue())
		}
		if spec.NetworkInterface.NicType != "" {
			niArgs.NicType = pulumi.StringPtr(spec.NetworkInterface.NicType)
		}
		gceSetup.NetworkInterfaces = workbench.InstanceGceSetupNetworkInterfaceArray{niArgs}
	}

	// Disable public IP.
	if spec.DisablePublicIp {
		gceSetup.DisablePublicIp = pulumi.BoolPtr(true)
	}

	// Enable IP forwarding.
	if spec.EnableIpForwarding {
		gceSetup.EnableIpForwarding = pulumi.BoolPtr(true)
	}

	// Service account.
	if spec.ServiceAccount != nil && spec.ServiceAccount.GetValue() != "" {
		saArgs := workbench.InstanceGceSetupServiceAccountArgs{
			Email: pulumi.StringPtr(spec.ServiceAccount.GetValue()),
		}
		gceSetup.ServiceAccounts = workbench.InstanceGceSetupServiceAccountArray{saArgs}
	}

	// Tags.
	if len(spec.Tags) > 0 {
		gceSetup.Tags = pulumi.ToStringArray(spec.Tags)
	}

	// Metadata.
	if len(spec.Metadata) > 0 {
		gceSetup.Metadata = pulumi.ToStringMap(spec.Metadata)
	}

	// VM image (mutually exclusive with container image).
	if spec.VmImage != nil {
		vmImageArgs := &workbench.InstanceGceSetupVmImageArgs{}
		if spec.VmImage.Project != "" {
			vmImageArgs.Project = pulumi.StringPtr(spec.VmImage.Project)
		}
		if spec.VmImage.Family != "" {
			vmImageArgs.Family = pulumi.StringPtr(spec.VmImage.Family)
		}
		if spec.VmImage.Name != "" {
			vmImageArgs.Name = pulumi.StringPtr(spec.VmImage.Name)
		}
		gceSetup.VmImage = vmImageArgs
	}

	// Container image (mutually exclusive with VM image).
	if spec.ContainerImage != nil {
		containerImageArgs := &workbench.InstanceGceSetupContainerImageArgs{
			Repository: pulumi.String(spec.ContainerImage.Repository),
		}
		if spec.ContainerImage.Tag != "" {
			containerImageArgs.Tag = pulumi.StringPtr(spec.ContainerImage.Tag)
		}
		gceSetup.ContainerImage = containerImageArgs
	}

	// Shielded instance config.
	if spec.ShieldedInstanceConfig != nil {
		shieldedArgs := &workbench.InstanceGceSetupShieldedInstanceConfigArgs{}
		if spec.ShieldedInstanceConfig.EnableSecureBoot {
			shieldedArgs.EnableSecureBoot = pulumi.BoolPtr(true)
		}
		if spec.ShieldedInstanceConfig.EnableVtpm {
			shieldedArgs.EnableVtpm = pulumi.BoolPtr(true)
		}
		if spec.ShieldedInstanceConfig.EnableIntegrityMonitoring {
			shieldedArgs.EnableIntegrityMonitoring = pulumi.BoolPtr(true)
		}
		gceSetup.ShieldedInstanceConfig = shieldedArgs
	}

	if hasGceSetup {
		args.GceSetup = gceSetup
	}

	createdInstance, err := workbench.NewInstance(ctx, "workbench-instance", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create workbench instance")
	}

	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpInstanceName, createdInstance.Name)
	ctx.Export(OpProxyUri, createdInstance.ProxyUri)
	ctx.Export(OpState, createdInstance.State)
	ctx.Export(OpCreator, createdInstance.Creator)
	ctx.Export(OpCreateTime, createdInstance.CreateTime)

	return nil
}
