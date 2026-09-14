package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/workbench"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func workbenchInstance(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiNotebook.Spec

	// Enable the Notebooks API — the control plane that owns Workbench
	// instances. DisableOnDestroy stays false: tearing down one notebook
	// must never disable the API for everything else in the project.
	notebooksApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("notebooks.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	// The Workbench VM itself is a Compute Engine instance provisioned by
	// the Notebooks service agent, so a fresh project also needs Compute.
	computeApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project.
	if spec.ProjectId.GetValue() != "" {
		notebooksApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		computeApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdNotebooksApi, err := projects.NewService(ctx,
		"gcpvnb-notebooks.googleapis.com", notebooksApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable notebooks.googleapis.com api")
	}
	createdComputeApi, err := projects.NewService(ctx,
		"gcpvnb-compute.googleapis.com", computeApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	// The Workbench instance — a managed JupyterLab VM. `Name` is the
	// instance ID GCP uses in the create call and MUST be set explicitly:
	// leaving it unset would let the engine auto-generate a suffixed name,
	// so the cloud-side instance would not match the spec's naming
	// contract (instance_name, falling back to metadata.name). The
	// provider's separate instance_id attribute is vestigial at create
	// time and deliberately not sent.
	args := &workbench.InstanceArgs{
		Name:     pulumi.StringPtr(locals.InstanceName),
		Location: pulumi.String(spec.Location),
		Labels:   pulumi.ToStringMap(locals.GcpLabels),
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// Client-side destroy behavior (DELETE deletes the instance and its
	// disks; PREVENT refuses; ABANDON drops from state but keeps the VM
	// running). Empty follows the provider default (DELETE) — mirrored
	// zero-vs-omit with Terraform.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// desired_state drives declarative stop/start: ACTIVE boots the VM,
	// STOPPED suspends compute billing while keeping disks.
	if spec.DesiredState != "" {
		args.DesiredState = pulumi.StringPtr(spec.DesiredState)
	}

	if spec.DisableProxyAccess {
		args.DisableProxyAccess = pulumi.BoolPtr(true)
	}

	if len(spec.InstanceOwners) > 0 {
		args.InstanceOwners = pulumi.ToStringArray(spec.InstanceOwners)
	}

	// Managed end-user credentials: JupyterLab runs as the accessing
	// user's own identity instead of the VM service account.
	if spec.EnableManagedEuc {
		args.EnableManagedEuc = pulumi.BoolPtr(true)
	}
	if spec.EnableThirdPartyIdentity {
		args.EnableThirdPartyIdentity = pulumi.BoolPtr(true)
	}

	// Build the GCE setup block from our flattened spec fields.
	gceSetup := &workbench.InstanceGceSetupArgs{
		MachineType: pulumi.StringPtr(spec.MachineType),
	}

	// Boot disk. Disk encryption is derived, never spec-set: presence of
	// a KMS key means CMEK, absence means Google-managed encryption.
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

	// Data disk (the API supports exactly one).
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

	// GPU accelerator (the API supports exactly one configuration).
	if spec.AcceleratorConfig != nil && spec.AcceleratorConfig.Type != "" {
		accelArgs := workbench.InstanceGceSetupAcceleratorConfigArgs{
			Type: pulumi.StringPtr(spec.AcceleratorConfig.Type),
		}
		if spec.AcceleratorConfig.CoreCount != 0 {
			accelArgs.CoreCount = pulumi.StringPtr(fmt.Sprintf("%d", spec.AcceleratorConfig.CoreCount))
		}
		gceSetup.AcceleratorConfigs = workbench.InstanceGceSetupAcceleratorConfigArray{accelArgs}
	}

	// Network interface (the API supports exactly one). An external_ip
	// pins a static address (ONE_TO_ONE_NAT); without it — and with
	// public IP not disabled — GCP assigns an ephemeral external address
	// that changes across stop/start cycles.
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
		if spec.NetworkInterface.ExternalIp != nil && spec.NetworkInterface.ExternalIp.GetValue() != "" {
			niArgs.AccessConfigs = workbench.InstanceGceSetupNetworkInterfaceAccessConfigArray{
				workbench.InstanceGceSetupNetworkInterfaceAccessConfigArgs{
					ExternalIp: pulumi.String(spec.NetworkInterface.ExternalIp.GetValue()),
				},
			}
		}
		gceSetup.NetworkInterfaces = workbench.InstanceGceSetupNetworkInterfaceArray{niArgs}
	}

	if spec.DisablePublicIp {
		gceSetup.DisablePublicIp = pulumi.BoolPtr(true)
	}

	if spec.EnableIpForwarding {
		gceSetup.EnableIpForwarding = pulumi.BoolPtr(true)
	}

	// VM identity (the API supports exactly one service account); scopes
	// are fixed to cloud-platform by the Workbench service.
	if spec.ServiceAccount != nil && spec.ServiceAccount.GetValue() != "" {
		saArgs := workbench.InstanceGceSetupServiceAccountArgs{
			Email: pulumi.StringPtr(spec.ServiceAccount.GetValue()),
		}
		gceSetup.ServiceAccounts = workbench.InstanceGceSetupServiceAccountArray{saArgs}
	}

	if len(spec.Tags) > 0 {
		gceSetup.Tags = pulumi.ToStringArray(spec.Tags)
	}

	if len(spec.Metadata) > 0 {
		gceSetup.Metadata = pulumi.ToStringMap(spec.Metadata)
	}

	// VM image (mutually exclusive with container image — enforced
	// pre-deploy by the spec's CEL rule).
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

	// Shielded VM posture (rootkit/bootkit protection). Each switch is
	// presence-tracked: an unset switch is omitted so GCP's default holds,
	// and an explicit value (true or false) is sent as authored.
	if shielded := spec.ShieldedInstanceConfig; shielded != nil {
		shieldedArgs := &workbench.InstanceGceSetupShieldedInstanceConfigArgs{}
		if shielded.EnableSecureBoot != nil {
			shieldedArgs.EnableSecureBoot = pulumi.BoolPtr(shielded.GetEnableSecureBoot())
		}
		if shielded.EnableVtpm != nil {
			shieldedArgs.EnableVtpm = pulumi.BoolPtr(shielded.GetEnableVtpm())
		}
		if shielded.EnableIntegrityMonitoring != nil {
			shieldedArgs.EnableIntegrityMonitoring = pulumi.BoolPtr(shielded.GetEnableIntegrityMonitoring())
		}
		gceSetup.ShieldedInstanceConfig = shieldedArgs
	}

	// Confidential Computing (AMD SEV): guest memory encrypted in use.
	// Requires an SEV-capable machine type (n2d family). The block's switch
	// (on by default once declared, filled by the platform before this runs)
	// decides whether it renders.
	if spec.ConfidentialInstanceConfig != nil && spec.ConfidentialInstanceConfig.GetEnabled() {
		confidentialArgs := &workbench.InstanceGceSetupConfidentialInstanceConfigArgs{}
		if spec.ConfidentialInstanceConfig.ConfidentialInstanceType != "" {
			confidentialArgs.ConfidentialInstanceType = pulumi.StringPtr(spec.ConfidentialInstanceConfig.ConfidentialInstanceType)
		}
		gceSetup.ConfidentialInstanceConfig = confidentialArgs
	}

	// Reservation affinity: consume pre-purchased Compute capacity — how
	// organizations guarantee GPU availability for ML workloads.
	if spec.ReservationAffinity != nil {
		reservationArgs := &workbench.InstanceGceSetupReservationAffinityArgs{}
		if spec.ReservationAffinity.ConsumeReservationType != "" {
			reservationArgs.ConsumeReservationType = pulumi.StringPtr(spec.ReservationAffinity.ConsumeReservationType)
		}
		if spec.ReservationAffinity.Key != "" {
			reservationArgs.Key = pulumi.StringPtr(spec.ReservationAffinity.Key)
		}
		if len(spec.ReservationAffinity.Values) > 0 {
			reservationArgs.Values = pulumi.ToStringArray(spec.ReservationAffinity.Values)
		}
		gceSetup.ReservationAffinity = reservationArgs
	}

	args.GceSetup = gceSetup

	createdInstance, err := workbench.NewInstance(ctx, "workbench-instance", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdNotebooksApi, createdComputeApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create workbench instance")
	}

	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpInstanceName, createdInstance.Name)
	ctx.Export(OpProxyUri, createdInstance.ProxyUri)
	ctx.Export(OpState, createdInstance.State)
	ctx.Export(OpCreator, createdInstance.Creator)
	ctx.Export(OpCreateTime, createdInstance.CreateTime)
	ctx.Export(OpHealthState, createdInstance.HealthState)
	ctx.Export(OpUpdateTime, createdInstance.UpdateTime)

	return nil
}
