package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurevirtualmachinev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualmachine/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys the virtual machine -- the compute shell wired to
// first-class referenced resources.
//
// The VM is deliberately just the machine (matching Azure's own model):
// network presence comes from referenced network interfaces, data volumes
// are referenced managed disks realized as attachment resources, and only
// the OS disk is inline -- unless the VM boots from an existing
// referenced OS disk.
//
// ARM models Linux and Windows VMs as separate management surfaces
// (different auth contracts, patch vocabularies, and OS settings), so the
// module deploys exactly one of the two resource types from the spec's
// explicit OS discriminator (os_profile.linux XOR os_profile.windows).
//
// Lifecycle notes worth knowing before operating this resource:
//   - Name, region, zone, image source, admin credentials, custom_data,
//     and the security/confidential posture are the VM's identity --
//     changing any of them replaces the VM (the OS disk with it; data
//     disks and NICs survive, which is exactly why they are referenced).
//   - Resizing (size) reboots in place. Spot settings are fixed at
//     creation.
func Resources(ctx *pulumi.Context, iacInput *azurevirtualmachinev1alpha1.AzureVirtualMachineIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVirtualMachine.Spec

	var vmId pulumi.StringOutput
	var vmName pulumi.StringOutput
	var vmGuid pulumi.StringOutput
	var privateIp pulumi.StringOutput
	var publicIp pulumi.StringOutput
	var computerName pulumi.StringOutput
	var principalId pulumi.StringOutput

	if locals.IsLinux {
		createdVm, err := compute.NewLinuxVirtualMachine(ctx,
			spec.Name,
			buildLinuxArgs(locals),
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create linux virtual machine %s", spec.Name)
		}
		vmId = createdVm.ID().ToStringOutput()
		vmName = createdVm.Name
		vmGuid = createdVm.VirtualMachineId
		privateIp = createdVm.PrivateIpAddress
		publicIp = createdVm.PublicIpAddress
		computerName = createdVm.ComputerName
		principalId = createdVm.Identity.ApplyT(func(identity *compute.LinuxVirtualMachineIdentity) string {
			if identity == nil || identity.PrincipalId == nil {
				return ""
			}
			return *identity.PrincipalId
		}).(pulumi.StringOutput)
	} else {
		createdVm, err := compute.NewWindowsVirtualMachine(ctx,
			spec.Name,
			buildWindowsArgs(locals),
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create windows virtual machine %s", spec.Name)
		}
		vmId = createdVm.ID().ToStringOutput()
		vmName = createdVm.Name
		vmGuid = createdVm.VirtualMachineId
		privateIp = createdVm.PrivateIpAddress
		publicIp = createdVm.PublicIpAddress
		computerName = createdVm.ComputerName
		principalId = createdVm.Identity.ApplyT(func(identity *compute.WindowsVirtualMachineIdentity) string {
			if identity == nil || identity.PrincipalId == nil {
				return ""
			}
			return *identity.PrincipalId
		}).(pulumi.StringOutput)
	}

	// Attach the referenced first-class data disks. Each attachment is
	// its own ARM operation (Azure's model): the disk -- and its data --
	// outlives the VM, and detaching is just removing the spec entry.
	for _, attachment := range spec.DataDiskAttachments {
		attachmentArgs := &compute.DataDiskAttachmentArgs{
			ManagedDiskId:    pulumi.String(attachment.ManagedDiskId.GetValue()),
			VirtualMachineId: vmId,
			Lun:              pulumi.Int(int(attachment.GetLun())),
			Caching:          pulumi.String(cachingToArm(attachment.Caching)),
		}
		if attachment.WriteAcceleratorEnabled {
			attachmentArgs.WriteAcceleratorEnabled = pulumi.Bool(true)
		}
		if _, err := compute.NewDataDiskAttachment(ctx,
			fmt.Sprintf("%s-lun-%d", spec.Name, attachment.GetLun()),
			attachmentArgs,
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to attach data disk at lun %d to virtual machine %s", attachment.Lun, spec.Name)
		}
	}

	// Export outputs from the created resource. The
	// system-assigned principal id is an empty string when the feature
	// is off, matching the Terraform module's try() fallback.
	ctx.Export(OpVmId, vmId)
	ctx.Export(OpVmName, vmName)
	ctx.Export(OpVirtualMachineGuid, vmGuid)
	ctx.Export(OpPrivateIpAddress, privateIp)
	ctx.Export(OpPublicIpAddress, publicIp)
	ctx.Export(OpComputerName, computerName)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, principalId)

	return nil
}
