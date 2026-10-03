package module

import (
	"github.com/pkg/errors"
	azureaksnodepoolv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureaksnodepool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates an AKS node pool: a scale set of worker nodes attached
// to an existing cluster by ARM ID.
//
// Design and lifecycle notes worth knowing before operating this resource:
//   - The pool is the unit of compute shape: general, memory-optimized,
//     GPU, spot, or Windows pools each live as their own resource with an
//     independent lifecycle. The cluster carries only its mandatory default
//     (system) pool.
//   - Spot pools (priority SPOT) trade 30-90% cost savings for evictability
//     and automatically carry the scalesetpriority spot taint; AKS replaces
//     evicted nodes as capacity returns. Eviction policy and max price are
//     fixed at creation.
//   - Many shape changes (vm_size, os_disk_type, fips, host encryption)
//     rotate the pool. Setting temporary_name_for_rotation lets AKS stand
//     up a replacement pool first instead of tearing this one down --
//     set it proactively on production pools.
//   - Node Kubernetes versions may lag the control plane by up to two
//     minor versions: orchestrator_version is the seam for canarying node
//     upgrades pool by pool.
func Resources(ctx *pulumi.Context, iacInput *azureaksnodepoolv1alpha1.AzureAksNodePoolIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureAksNodePool.Spec

	poolArgs := &containerservice.KubernetesClusterNodePoolArgs{
		Name:                  pulumi.String(spec.Name),
		KubernetesClusterId:   pulumi.String(locals.KubernetesClusterId),
		VmSize:                pulumi.String(spec.VmSize),
		Mode:                  pulumi.String(modeToArm(spec.Mode)),
		OsType:                pulumi.String(osTypeToArm(spec.OsType)),
		NodeCount:             pulumi.Int(int(spec.NodeCount)),
		AutoScalingEnabled:    pulumi.Bool(spec.AutoScalingEnabled),
		FipsEnabled:           pulumi.Bool(spec.FipsEnabled),
		HostEncryptionEnabled: pulumi.Bool(spec.HostEncryptionEnabled),
		NodePublicIpEnabled:   pulumi.Bool(spec.NodePublicIpEnabled),
		UltraSsdEnabled:       pulumi.Bool(spec.UltraSsdEnabled),
		Tags:                  pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.AutoScalingEnabled {
		poolArgs.MinCount = pulumi.Int(int(spec.MinCount))
		poolArgs.MaxCount = pulumi.Int(int(spec.MaxCount))
	}

	if spec.MaxPods > 0 {
		poolArgs.MaxPods = pulumi.Int(int(spec.MaxPods))
	}

	if len(spec.Zones) > 0 {
		poolArgs.Zones = pulumi.ToStringArray(spec.Zones)
	}

	if v := osSkuToArm(spec.OsSku); v != "" {
		poolArgs.OsSku = pulumi.String(v)
	}

	// Spot economics only render for SPOT pools; ARM rejects them on
	// Regular pools (spec-level validation enforces the pairing too).
	if spec.Priority == azureaksnodepoolv1alpha1.AzureAksNodePoolPriority_SPOT {
		poolArgs.Priority = pulumi.String("Spot")
		if v := evictionPolicyToArm(spec.EvictionPolicy); v != "" {
			poolArgs.EvictionPolicy = pulumi.String(v)
		}
		if spec.SpotMaxPrice != 0 {
			poolArgs.SpotMaxPrice = pulumi.Float64(spec.SpotMaxPrice)
		} else {
			poolArgs.SpotMaxPrice = pulumi.Float64(-1)
		}
	} else {
		poolArgs.Priority = pulumi.String("Regular")
	}

	if len(spec.NodeLabels) > 0 {
		poolArgs.NodeLabels = pulumi.ToStringMap(spec.NodeLabels)
	}
	if len(spec.NodeTaints) > 0 {
		poolArgs.NodeTaints = pulumi.ToStringArray(spec.NodeTaints)
	}

	if subnetId := spec.VnetSubnetId.GetValue(); subnetId != "" {
		poolArgs.VnetSubnetId = pulumi.String(subnetId)
	}
	if podSubnetId := spec.PodSubnetId.GetValue(); podSubnetId != "" {
		poolArgs.PodSubnetId = pulumi.String(podSubnetId)
	}

	if spec.OrchestratorVersion != "" {
		poolArgs.OrchestratorVersion = pulumi.String(spec.OrchestratorVersion)
	}

	if spec.OsDiskSizeGb > 0 {
		poolArgs.OsDiskSizeGb = pulumi.Int(int(spec.OsDiskSizeGb))
	}
	if v := osDiskTypeToArm(spec.OsDiskType); v != "" {
		poolArgs.OsDiskType = pulumi.String(v)
	} else {
		poolArgs.OsDiskType = pulumi.String("Managed")
	}
	if v := kubeletDiskTypeToArm(spec.KubeletDiskType); v != "" {
		poolArgs.KubeletDiskType = pulumi.String(v)
	}

	if prefixId := spec.NodePublicIpPrefixId.GetValue(); prefixId != "" {
		poolArgs.NodePublicIpPrefixId = pulumi.String(prefixId)
	}

	if v := gpuInstanceToArm(spec.GpuInstance); v != "" {
		poolArgs.GpuInstance = pulumi.String(v)
	}
	if v := gpuDriverToArm(spec.GpuDriver); v != "" {
		poolArgs.GpuDriver = pulumi.String(v)
	}

	if spec.ProximityPlacementGroupId != "" {
		poolArgs.ProximityPlacementGroupId = pulumi.String(spec.ProximityPlacementGroupId)
	}
	if spec.HostGroupId != "" {
		poolArgs.HostGroupId = pulumi.String(spec.HostGroupId)
	}
	if spec.CapacityReservationGroupId != "" {
		poolArgs.CapacityReservationGroupId = pulumi.String(spec.CapacityReservationGroupId)
	}

	if v := scaleDownModeToArm(spec.ScaleDownMode); v != "" {
		poolArgs.ScaleDownMode = pulumi.String(v)
	} else {
		poolArgs.ScaleDownMode = pulumi.String("Delete")
	}
	if spec.SnapshotId != "" {
		poolArgs.SnapshotId = pulumi.String(spec.SnapshotId)
	}
	if v := workloadRuntimeToArm(spec.WorkloadRuntime); v != "" {
		poolArgs.WorkloadRuntime = pulumi.String(v)
	}

	if spec.TemporaryNameForRotation != "" {
		poolArgs.TemporaryNameForRotation = pulumi.String(spec.TemporaryNameForRotation)
	}

	if spec.KubeletConfig != nil {
		poolArgs.KubeletConfig = buildKubeletConfig(spec.KubeletConfig)
	}
	if spec.LinuxOsConfig != nil {
		poolArgs.LinuxOsConfig = buildLinuxOsConfig(spec.LinuxOsConfig)
	}
	if spec.NodeNetworkProfile != nil {
		poolArgs.NodeNetworkProfile = buildNodeNetworkProfile(spec.NodeNetworkProfile)
	}
	if spec.UpgradeSettings != nil {
		poolArgs.UpgradeSettings = buildUpgradeSettings(spec.UpgradeSettings)
	}
	if spec.WindowsProfile != nil {
		poolArgs.WindowsProfile = buildWindowsProfile(spec.WindowsProfile)
	}

	nodePool, err := containerservice.NewKubernetesClusterNodePool(
		ctx,
		locals.AzureAksNodePool.Metadata.Name,
		poolArgs,
		pulumi.Provider(azureProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create AKS node pool")
	}

	ctx.Export(OpNodePoolId, nodePool.ID())
	ctx.Export(OpNodePoolName, nodePool.Name)
	ctx.Export(OpNodeImageVersion, nodePool.NodeImageVersion)

	return nil
}

func modeToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolMode) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolMode_SYSTEM:
		return "System"
	default:
		return "User"
	}
}

func osTypeToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolOsType) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsType_WINDOWS:
		return "Windows"
	default:
		return "Linux"
	}
}

func evictionPolicyToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolEvictionPolicy) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolEvictionPolicy_EVICTION_DEALLOCATE:
		return "Deallocate"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolEvictionPolicy_EVICTION_DELETE:
		return "Delete"
	default:
		return ""
	}
}

// buildUpgradeSettings passes the spec's rollout settings straight through:
// spec-level CEL already forbids surge/unavailable values on spot pools, so
// both engines send identical payloads without module-side filtering.
func buildUpgradeSettings(cfg *azureaksnodepoolv1alpha1.AzureAksNodePoolUpgradeSettings) *containerservice.KubernetesClusterNodePoolUpgradeSettingsArgs {
	args := &containerservice.KubernetesClusterNodePoolUpgradeSettingsArgs{}
	if cfg.MaxSurge != "" {
		args.MaxSurge = pulumi.String(cfg.MaxSurge)
	}
	if cfg.MaxUnavailable != "" {
		args.MaxUnavailable = pulumi.String(cfg.MaxUnavailable)
	}
	if cfg.DrainTimeoutInMinutes > 0 {
		args.DrainTimeoutInMinutes = pulumi.Int(int(cfg.DrainTimeoutInMinutes))
	}
	if cfg.NodeSoakDurationInMinutes > 0 {
		args.NodeSoakDurationInMinutes = pulumi.Int(int(cfg.NodeSoakDurationInMinutes))
	}
	if v := undrainableNodeBehaviorToArm(cfg.UndrainableNodeBehavior); v != "" {
		args.UndrainableNodeBehavior = pulumi.String(v)
	}
	return args
}

func buildWindowsProfile(cfg *azureaksnodepoolv1alpha1.AzureAksNodePoolWindowsProfile) *containerservice.KubernetesClusterNodePoolWindowsProfileArgs {
	args := &containerservice.KubernetesClusterNodePoolWindowsProfileArgs{}
	if cfg.OutboundNatEnabled != nil {
		args.OutboundNatEnabled = pulumi.Bool(cfg.GetOutboundNatEnabled())
	} else {
		args.OutboundNatEnabled = pulumi.Bool(true)
	}
	return args
}

func buildKubeletConfig(cfg *azureaksnodepoolv1alpha1.AzureAksNodePoolKubeletConfig) *containerservice.KubernetesClusterNodePoolKubeletConfigArgs {
	args := &containerservice.KubernetesClusterNodePoolKubeletConfigArgs{}
	if v := cpuManagerPolicyToArm(cfg.CpuManagerPolicy); v != "" {
		args.CpuManagerPolicy = pulumi.String(v)
	}
	if cfg.CpuCfsQuotaEnabled != nil {
		args.CpuCfsQuotaEnabled = pulumi.Bool(cfg.GetCpuCfsQuotaEnabled())
	} else {
		args.CpuCfsQuotaEnabled = pulumi.Bool(true)
	}
	if cfg.CpuCfsQuotaPeriod != "" {
		args.CpuCfsQuotaPeriod = pulumi.String(cfg.CpuCfsQuotaPeriod)
	}
	if cfg.ImageGcHighThreshold > 0 {
		args.ImageGcHighThreshold = pulumi.Int(int(cfg.ImageGcHighThreshold))
	}
	if cfg.ImageGcLowThreshold > 0 {
		args.ImageGcLowThreshold = pulumi.Int(int(cfg.ImageGcLowThreshold))
	}
	if v := topologyManagerPolicyToArm(cfg.TopologyManagerPolicy); v != "" {
		args.TopologyManagerPolicy = pulumi.String(v)
	}
	if len(cfg.AllowedUnsafeSysctls) > 0 {
		args.AllowedUnsafeSysctls = pulumi.ToStringArray(cfg.AllowedUnsafeSysctls)
	}
	if cfg.ContainerLogMaxSizeMb > 0 {
		args.ContainerLogMaxSizeMb = pulumi.Int(int(cfg.ContainerLogMaxSizeMb))
	}
	if cfg.ContainerLogMaxFiles > 0 {
		args.ContainerLogMaxFiles = pulumi.Int(int(cfg.ContainerLogMaxFiles))
	}
	if cfg.PodMaxPid > 0 {
		args.PodMaxPid = pulumi.Int(int(cfg.PodMaxPid))
	}
	return args
}

func buildLinuxOsConfig(cfg *azureaksnodepoolv1alpha1.AzureAksNodePoolLinuxOsConfig) *containerservice.KubernetesClusterNodePoolLinuxOsConfigArgs {
	args := &containerservice.KubernetesClusterNodePoolLinuxOsConfigArgs{}
	if v := transparentHugePageToArm(cfg.TransparentHugePage); v != "" {
		args.TransparentHugePage = pulumi.String(v)
	}
	if v := transparentHugePageDefragToArm(cfg.TransparentHugePageDefrag); v != "" {
		args.TransparentHugePageDefrag = pulumi.String(v)
	}
	if cfg.SwapFileSizeMb > 0 {
		args.SwapFileSizeMb = pulumi.Int(int(cfg.SwapFileSizeMb))
	}
	if cfg.SysctlConfig != nil {
		args.SysctlConfig = buildSysctlConfig(cfg.SysctlConfig)
	}
	return args
}

func buildSysctlConfig(cfg *azureaksnodepoolv1alpha1.AzureAksNodePoolSysctlConfig) *containerservice.KubernetesClusterNodePoolLinuxOsConfigSysctlConfigArgs {
	args := &containerservice.KubernetesClusterNodePoolLinuxOsConfigSysctlConfigArgs{}
	if cfg.FsAioMaxNr > 0 {
		args.FsAioMaxNr = pulumi.Int(int(cfg.FsAioMaxNr))
	}
	if cfg.FsFileMax > 0 {
		args.FsFileMax = pulumi.Int(int(cfg.FsFileMax))
	}
	if cfg.FsInotifyMaxUserWatches > 0 {
		args.FsInotifyMaxUserWatches = pulumi.Int(int(cfg.FsInotifyMaxUserWatches))
	}
	if cfg.FsNrOpen > 0 {
		args.FsNrOpen = pulumi.Int(int(cfg.FsNrOpen))
	}
	if cfg.KernelThreadsMax > 0 {
		args.KernelThreadsMax = pulumi.Int(int(cfg.KernelThreadsMax))
	}
	if cfg.NetCoreNetdevMaxBacklog > 0 {
		args.NetCoreNetdevMaxBacklog = pulumi.Int(int(cfg.NetCoreNetdevMaxBacklog))
	}
	if cfg.NetCoreOptmemMax > 0 {
		args.NetCoreOptmemMax = pulumi.Int(int(cfg.NetCoreOptmemMax))
	}
	if cfg.NetCoreRmemDefault > 0 {
		args.NetCoreRmemDefault = pulumi.Int(int(cfg.NetCoreRmemDefault))
	}
	if cfg.NetCoreRmemMax > 0 {
		args.NetCoreRmemMax = pulumi.Int(int(cfg.NetCoreRmemMax))
	}
	if cfg.NetCoreSomaxconn > 0 {
		args.NetCoreSomaxconn = pulumi.Int(int(cfg.NetCoreSomaxconn))
	}
	if cfg.NetCoreWmemDefault > 0 {
		args.NetCoreWmemDefault = pulumi.Int(int(cfg.NetCoreWmemDefault))
	}
	if cfg.NetCoreWmemMax > 0 {
		args.NetCoreWmemMax = pulumi.Int(int(cfg.NetCoreWmemMax))
	}
	if cfg.NetIpv4IpLocalPortRangeMin > 0 {
		args.NetIpv4IpLocalPortRangeMin = pulumi.Int(int(cfg.NetIpv4IpLocalPortRangeMin))
	}
	if cfg.NetIpv4IpLocalPortRangeMax > 0 {
		args.NetIpv4IpLocalPortRangeMax = pulumi.Int(int(cfg.NetIpv4IpLocalPortRangeMax))
	}
	if cfg.NetIpv4NeighDefaultGcThresh1 > 0 {
		args.NetIpv4NeighDefaultGcThresh1 = pulumi.Int(int(cfg.NetIpv4NeighDefaultGcThresh1))
	}
	if cfg.NetIpv4NeighDefaultGcThresh2 > 0 {
		args.NetIpv4NeighDefaultGcThresh2 = pulumi.Int(int(cfg.NetIpv4NeighDefaultGcThresh2))
	}
	if cfg.NetIpv4NeighDefaultGcThresh3 > 0 {
		args.NetIpv4NeighDefaultGcThresh3 = pulumi.Int(int(cfg.NetIpv4NeighDefaultGcThresh3))
	}
	if cfg.NetIpv4TcpFinTimeout > 0 {
		args.NetIpv4TcpFinTimeout = pulumi.Int(int(cfg.NetIpv4TcpFinTimeout))
	}
	if cfg.NetIpv4TcpKeepaliveIntvl > 0 {
		args.NetIpv4TcpKeepaliveIntvl = pulumi.Int(int(cfg.NetIpv4TcpKeepaliveIntvl))
	}
	if cfg.NetIpv4TcpKeepaliveProbes > 0 {
		args.NetIpv4TcpKeepaliveProbes = pulumi.Int(int(cfg.NetIpv4TcpKeepaliveProbes))
	}
	if cfg.NetIpv4TcpKeepaliveTime > 0 {
		args.NetIpv4TcpKeepaliveTime = pulumi.Int(int(cfg.NetIpv4TcpKeepaliveTime))
	}
	if cfg.NetIpv4TcpMaxSynBacklog > 0 {
		args.NetIpv4TcpMaxSynBacklog = pulumi.Int(int(cfg.NetIpv4TcpMaxSynBacklog))
	}
	if cfg.NetIpv4TcpMaxTwBuckets > 0 {
		args.NetIpv4TcpMaxTwBuckets = pulumi.Int(int(cfg.NetIpv4TcpMaxTwBuckets))
	}
	if cfg.NetIpv4TcpTwReuse {
		args.NetIpv4TcpTwReuse = pulumi.Bool(true)
	}
	if cfg.NetNetfilterNfConntrackBuckets > 0 {
		args.NetNetfilterNfConntrackBuckets = pulumi.Int(int(cfg.NetNetfilterNfConntrackBuckets))
	}
	if cfg.NetNetfilterNfConntrackMax > 0 {
		args.NetNetfilterNfConntrackMax = pulumi.Int(int(cfg.NetNetfilterNfConntrackMax))
	}
	if cfg.VmMaxMapCount > 0 {
		args.VmMaxMapCount = pulumi.Int(int(cfg.VmMaxMapCount))
	}
	if cfg.VmSwappiness > 0 {
		args.VmSwappiness = pulumi.Int(int(cfg.VmSwappiness))
	}
	if cfg.VmVfsCachePressure > 0 {
		args.VmVfsCachePressure = pulumi.Int(int(cfg.VmVfsCachePressure))
	}
	return args
}

func buildNodeNetworkProfile(cfg *azureaksnodepoolv1alpha1.AzureAksNodePoolNodeNetworkProfile) *containerservice.KubernetesClusterNodePoolNodeNetworkProfileArgs {
	args := &containerservice.KubernetesClusterNodePoolNodeNetworkProfileArgs{}
	if len(cfg.AllowedHostPorts) > 0 {
		hostPorts := containerservice.KubernetesClusterNodePoolNodeNetworkProfileAllowedHostPortArray{}
		for _, hostPort := range cfg.AllowedHostPorts {
			hostPortArgs := containerservice.KubernetesClusterNodePoolNodeNetworkProfileAllowedHostPortArgs{}
			if hostPort.PortStart > 0 {
				hostPortArgs.PortStart = pulumi.Int(int(hostPort.PortStart))
			}
			if hostPort.PortEnd > 0 {
				hostPortArgs.PortEnd = pulumi.Int(int(hostPort.PortEnd))
			}
			if v := hostPortProtocolToArm(hostPort.Protocol); v != "" {
				hostPortArgs.Protocol = pulumi.String(v)
			}
			hostPorts = append(hostPorts, hostPortArgs)
		}
		args.AllowedHostPorts = hostPorts
	}
	if len(cfg.ApplicationSecurityGroupIds) > 0 {
		args.ApplicationSecurityGroupIds = pulumi.ToStringArray(cfg.ApplicationSecurityGroupIds)
	}
	if len(cfg.NodePublicIpTags) > 0 {
		args.NodePublicIpTags = pulumi.ToStringMap(cfg.NodePublicIpTags)
	}
	return args
}
