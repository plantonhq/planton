package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ocicomputeinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocicomputeinstance/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func instance(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciComputeInstance.Spec

	args := &core.InstanceArgs{
		CompartmentId:      pulumi.String(spec.CompartmentId.GetValue()),
		AvailabilityDomain: pulumi.String(spec.AvailabilityDomain),
		Shape:              pulumi.StringPtr(spec.Shape),
		DisplayName:        pulumi.StringPtr(locals.DisplayName),
		FreeformTags:       pulumi.ToStringMap(locals.FreeformTags),
		SourceDetails:      buildSourceDetails(spec.SourceDetails),
		CreateVnicDetails:  buildCreateVnicDetails(spec.CreateVnicDetails),
	}

	if len(spec.Metadata) > 0 {
		args.Metadata = pulumi.ToStringMap(spec.Metadata)
	}

	if spec.FaultDomain != "" {
		args.FaultDomain = pulumi.StringPtr(spec.FaultDomain)
	}

	if spec.IsPvEncryptionInTransitEnabled != nil {
		args.IsPvEncryptionInTransitEnabled = pulumi.BoolPtr(*spec.IsPvEncryptionInTransitEnabled)
	}

	if spec.ShapeConfig != nil {
		args.ShapeConfig = buildShapeConfig(spec.ShapeConfig)
	}

	if spec.AgentConfig != nil {
		args.AgentConfig = buildAgentConfig(spec.AgentConfig)
	}

	if spec.AvailabilityConfig != nil {
		args.AvailabilityConfig = buildAvailabilityConfig(spec.AvailabilityConfig)
	}

	if spec.LaunchOptions != nil {
		args.LaunchOptions = buildLaunchOptions(spec.LaunchOptions)
	}

	if spec.InstanceOptions != nil {
		args.InstanceOptions = buildInstanceOptions(spec.InstanceOptions)
	}

	if spec.PreemptibleInstanceConfig != nil {
		args.PreemptibleInstanceConfig = buildPreemptibleConfig(spec.PreemptibleInstanceConfig)
	}

	if spec.CapacityReservationId != nil {
		args.CapacityReservationId = pulumi.StringPtr(spec.CapacityReservationId.GetValue())
	}

	if spec.DedicatedVmHostId != nil {
		args.DedicatedVmHostId = pulumi.StringPtr(spec.DedicatedVmHostId.GetValue())
	}

	if spec.PlatformConfig != nil {
		args.PlatformConfig = buildPlatformConfig(spec.PlatformConfig)
	}

	createdInstance, err := core.NewInstance(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci compute instance")
	}

	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpPrivateIp, createdInstance.PrivateIp)
	ctx.Export(OpPublicIp, createdInstance.PublicIp)
	ctx.Export(OpBootVolumeId, createdInstance.BootVolumeId)
	ctx.Export(OpAvailabilityDomain, createdInstance.AvailabilityDomain)

	return nil
}

func buildSourceDetails(sd *ocicomputeinstancev1.OciComputeInstanceSpec_SourceDetails) core.InstanceSourceDetailsArgs {
	args := core.InstanceSourceDetailsArgs{
		SourceType: pulumi.String(sourceTypeString(sd.SourceType)),
		SourceId:   pulumi.StringPtr(sd.SourceId),
	}

	if sd.BootVolumeSizeInGbs != nil {
		args.BootVolumeSizeInGbs = pulumi.StringPtr(fmt.Sprintf("%d", *sd.BootVolumeSizeInGbs))
	}

	if sd.BootVolumeVpusPerGb != nil {
		args.BootVolumeVpusPerGb = pulumi.StringPtr(fmt.Sprintf("%d", *sd.BootVolumeVpusPerGb))
	}

	if sd.KmsKeyId != nil {
		args.KmsKeyId = pulumi.StringPtr(sd.KmsKeyId.GetValue())
	}

	return args
}

func buildCreateVnicDetails(vnic *ocicomputeinstancev1.OciComputeInstanceSpec_CreateVnicDetails) core.InstanceCreateVnicDetailsArgs {
	args := core.InstanceCreateVnicDetailsArgs{
		SubnetId: pulumi.StringPtr(vnic.SubnetId.GetValue()),
	}

	if len(vnic.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(vnic.NsgIds))
		for i, nsg := range vnic.NsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.NsgIds = nsgIds
	}

	if vnic.AssignPublicIp != nil {
		if *vnic.AssignPublicIp {
			args.AssignPublicIp = pulumi.StringPtr("true")
		} else {
			args.AssignPublicIp = pulumi.StringPtr("false")
		}
	}

	if vnic.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(vnic.DisplayName)
	}

	if vnic.HostnameLabel != "" {
		args.HostnameLabel = pulumi.StringPtr(vnic.HostnameLabel)
	}

	if vnic.PrivateIp != "" {
		args.PrivateIp = pulumi.StringPtr(vnic.PrivateIp)
	}

	if vnic.SkipSourceDestCheck != nil {
		args.SkipSourceDestCheck = pulumi.BoolPtr(*vnic.SkipSourceDestCheck)
	}

	if vnic.AssignPrivateDnsRecord != nil {
		args.AssignPrivateDnsRecord = pulumi.BoolPtr(*vnic.AssignPrivateDnsRecord)
	}

	return args
}

func buildShapeConfig(sc *ocicomputeinstancev1.OciComputeInstanceSpec_ShapeConfig) core.InstanceShapeConfigArgs {
	args := core.InstanceShapeConfigArgs{}

	if sc.Ocpus != nil {
		args.Ocpus = pulumi.Float64Ptr(float64(*sc.Ocpus))
	}

	if sc.MemoryInGbs != nil {
		args.MemoryInGbs = pulumi.Float64Ptr(float64(*sc.MemoryInGbs))
	}

	if sc.BaselineOcpuUtilization != "" {
		args.BaselineOcpuUtilization = pulumi.StringPtr(sc.BaselineOcpuUtilization)
	}

	if sc.Nvmes != nil {
		args.Nvmes = pulumi.IntPtr(int(*sc.Nvmes))
	}

	return args
}

func buildAgentConfig(ac *ocicomputeinstancev1.OciComputeInstanceSpec_AgentConfig) core.InstanceAgentConfigArgs {
	args := core.InstanceAgentConfigArgs{}

	if ac.AreAllPluginsDisabled != nil {
		args.AreAllPluginsDisabled = pulumi.BoolPtr(*ac.AreAllPluginsDisabled)
	}

	if ac.IsManagementDisabled != nil {
		args.IsManagementDisabled = pulumi.BoolPtr(*ac.IsManagementDisabled)
	}

	if ac.IsMonitoringDisabled != nil {
		args.IsMonitoringDisabled = pulumi.BoolPtr(*ac.IsMonitoringDisabled)
	}

	if len(ac.PluginsConfig) > 0 {
		plugins := make(core.InstanceAgentConfigPluginsConfigArray, len(ac.PluginsConfig))
		for i, pc := range ac.PluginsConfig {
			plugins[i] = core.InstanceAgentConfigPluginsConfigArgs{
				Name:         pulumi.String(pc.Name),
				DesiredState: pulumi.String(strings.ToUpper(pc.DesiredState.String())),
			}
		}
		args.PluginsConfigs = plugins
	}

	return args
}

func buildAvailabilityConfig(ac *ocicomputeinstancev1.OciComputeInstanceSpec_AvailabilityConfig) core.InstanceAvailabilityConfigArgs {
	args := core.InstanceAvailabilityConfigArgs{}

	if ac.IsLiveMigrationPreferred != nil {
		args.IsLiveMigrationPreferred = pulumi.BoolPtr(*ac.IsLiveMigrationPreferred)
	}

	if ac.RecoveryAction != ocicomputeinstancev1.OciComputeInstanceSpec_AvailabilityConfig_recovery_action_unspecified {
		args.RecoveryAction = pulumi.StringPtr(strings.ToUpper(ac.RecoveryAction.String()))
	}

	return args
}

func buildLaunchOptions(lo *ocicomputeinstancev1.OciComputeInstanceSpec_LaunchOptions) core.InstanceLaunchOptionsArgs {
	args := core.InstanceLaunchOptionsArgs{}

	if lo.BootVolumeType != "" {
		args.BootVolumeType = pulumi.StringPtr(lo.BootVolumeType)
	}

	if lo.NetworkType != "" {
		args.NetworkType = pulumi.StringPtr(lo.NetworkType)
	}

	if lo.Firmware != ocicomputeinstancev1.OciComputeInstanceSpec_LaunchOptions_firmware_unspecified {
		args.Firmware = pulumi.StringPtr(strings.ToUpper(lo.Firmware.String()))
	}

	if lo.IsPvEncryptionInTransitEnabled != nil {
		args.IsPvEncryptionInTransitEnabled = pulumi.BoolPtr(*lo.IsPvEncryptionInTransitEnabled)
	}

	if lo.IsConsistentVolumeNamingEnabled != nil {
		args.IsConsistentVolumeNamingEnabled = pulumi.BoolPtr(*lo.IsConsistentVolumeNamingEnabled)
	}

	return args
}

func buildInstanceOptions(io *ocicomputeinstancev1.OciComputeInstanceSpec_InstanceOptions) core.InstanceInstanceOptionsArgs {
	args := core.InstanceInstanceOptionsArgs{}

	if io.AreLegacyImdsEndpointsDisabled != nil {
		args.AreLegacyImdsEndpointsDisabled = pulumi.BoolPtr(*io.AreLegacyImdsEndpointsDisabled)
	}

	return args
}

func buildPreemptibleConfig(pc *ocicomputeinstancev1.OciComputeInstanceSpec_PreemptibleInstanceConfig) core.InstancePreemptibleInstanceConfigArgs {
	preemptionAction := core.InstancePreemptibleInstanceConfigPreemptionActionArgs{
		Type: pulumi.String("TERMINATE"),
	}

	if pc.PreserveBootVolume != nil {
		preemptionAction.PreserveBootVolume = pulumi.BoolPtr(*pc.PreserveBootVolume)
	}

	return core.InstancePreemptibleInstanceConfigArgs{
		PreemptionAction: preemptionAction,
	}
}

func buildPlatformConfig(pc *ocicomputeinstancev1.OciComputeInstanceSpec_PlatformConfig) core.InstancePlatformConfigArgs {
	args := core.InstancePlatformConfigArgs{
		Type: pulumi.String(strings.ToUpper(pc.Type.String())),
	}

	if pc.IsSecureBootEnabled != nil {
		args.IsSecureBootEnabled = pulumi.BoolPtr(*pc.IsSecureBootEnabled)
	}

	if pc.IsMeasuredBootEnabled != nil {
		args.IsMeasuredBootEnabled = pulumi.BoolPtr(*pc.IsMeasuredBootEnabled)
	}

	if pc.IsTrustedPlatformModuleEnabled != nil {
		args.IsTrustedPlatformModuleEnabled = pulumi.BoolPtr(*pc.IsTrustedPlatformModuleEnabled)
	}

	if pc.IsMemoryEncryptionEnabled != nil {
		args.IsMemoryEncryptionEnabled = pulumi.BoolPtr(*pc.IsMemoryEncryptionEnabled)
	}

	if pc.IsSymmetricMultiThreadingEnabled != nil {
		args.IsSymmetricMultiThreadingEnabled = pulumi.BoolPtr(*pc.IsSymmetricMultiThreadingEnabled)
	}

	if pc.AreVirtualInstructionsEnabled != nil {
		args.AreVirtualInstructionsEnabled = pulumi.BoolPtr(*pc.AreVirtualInstructionsEnabled)
	}

	if pc.IsAccessControlServiceEnabled != nil {
		args.IsAccessControlServiceEnabled = pulumi.BoolPtr(*pc.IsAccessControlServiceEnabled)
	}

	if pc.IsInputOutputMemoryManagementUnitEnabled != nil {
		args.IsInputOutputMemoryManagementUnitEnabled = pulumi.BoolPtr(*pc.IsInputOutputMemoryManagementUnitEnabled)
	}

	if pc.NumaNodesPerSocket != "" {
		args.NumaNodesPerSocket = pulumi.StringPtr(pc.NumaNodesPerSocket)
	}

	if pc.PercentageOfCoresEnabled != nil {
		args.PercentageOfCoresEnabled = pulumi.IntPtr(int(*pc.PercentageOfCoresEnabled))
	}

	return args
}

func sourceTypeString(st ocicomputeinstancev1.OciComputeInstanceSpec_SourceDetails_SourceType) string {
	switch st {
	case ocicomputeinstancev1.OciComputeInstanceSpec_SourceDetails_image:
		return "image"
	case ocicomputeinstancev1.OciComputeInstanceSpec_SourceDetails_boot_volume:
		return "bootVolume"
	default:
		return "image"
	}
}
