package module

import (
	"encoding/base64"
	"strconv"

	"github.com/pkg/errors"
	azureaksclusterv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureakscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the AKS managed cluster: the control plane, its
// identity and access model, its network fabric, the mandatory default
// (system) node pool, and the Azure-managed add-ons.
//
// Design and lifecycle notes worth knowing before operating this resource:
//   - The cluster carries exactly ONE node pool -- the default pool Azure
//     requires at creation. Every additional pool is a standalone
//     AzureAksNodePool resource referencing this cluster's ID; pools have
//     independent lifecycles and coupling them here would force cluster
//     updates for pool changes.
//   - The network fabric (plugin, mode, CIDRs, outbound type) mostly
//     replaces the cluster when changed -- decide the network model up
//     front. The module writes the modern AKS default explicitly (Azure
//     CNI in overlay mode) when the spec leaves it unspecified, because
//     kubenet -- the provider's implicit fallback -- is deprecated and
//     retires in 2028.
//   - The OIDC issuer defaults ON (the spec's default, deliberately above
//     Azure's provisioning default): it is the trust anchor for workload
//     identity federation (AzureFederatedIdentityCredential consumes the
//     oidc_issuer_url output), costs nothing, and disabling it after
//     enabling forces cluster replacement.
//   - The legacy service-principal auth block is deliberately not modeled:
//     managed identity is Azure's own stated direction, and a client
//     secret in cluster config is exactly the credential class the
//     platform is built to eliminate.
func Resources(ctx *pulumi.Context, stackInput *azureaksclusterv1alpha1.AzureAksClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureAksCluster.Spec

	clusterArgs := &containerservice.KubernetesClusterArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// ARM requires exactly one DNS prefix flavor. When the spec sets
	// neither, derive the public prefix from the cluster name -- the
	// behavior nearly everyone wants -- unless this is a private cluster
	// carrying its own private prefix.
	if spec.DnsPrefix != "" {
		clusterArgs.DnsPrefix = pulumi.String(spec.DnsPrefix)
	} else if spec.DnsPrefixPrivateCluster != "" {
		clusterArgs.DnsPrefixPrivateCluster = pulumi.String(spec.DnsPrefixPrivateCluster)
	} else {
		clusterArgs.DnsPrefix = pulumi.String(spec.Name)
	}

	// Unset provisions the latest AKS-recommended GA version; production
	// clusters pin a version in the spec so upgrades are deliberate.
	if spec.KubernetesVersion != "" {
		clusterArgs.KubernetesVersion = pulumi.String(spec.KubernetesVersion)
	}

	// Written explicitly (rather than leaving the provider defaults) so
	// both engines and ARM's returned state stay aligned.
	if v := skuTierToArm(spec.SkuTier); v != "" {
		clusterArgs.SkuTier = pulumi.String(v)
	} else {
		clusterArgs.SkuTier = pulumi.String("Free")
	}
	if v := supportPlanToArm(spec.SupportPlan); v != "" {
		clusterArgs.SupportPlan = pulumi.String(v)
	} else {
		clusterArgs.SupportPlan = pulumi.String("KubernetesOfficial")
	}

	if v := upgradeChannelToArm(spec.AutomaticUpgradeChannel); v != "" {
		clusterArgs.AutomaticUpgradeChannel = pulumi.String(v)
	}
	if v := nodeOsUpgradeChannelToArm(spec.NodeOsUpgradeChannel); v != "" {
		clusterArgs.NodeOsUpgradeChannel = pulumi.String(v)
	} else {
		clusterArgs.NodeOsUpgradeChannel = pulumi.String("NodeImage")
	}

	// Presence-guarded: the optional bool defaults to true in the proto,
	// but the getter returns false when the field is unset. An absent spec
	// value falls back to true -- the same value the Terraform module's
	// optional(bool, true) encodes -- so an untouched manifest keeps the
	// OIDC issuer (the workload-identity trust anchor) on.
	if spec.OidcIssuerEnabled != nil {
		clusterArgs.OidcIssuerEnabled = pulumi.Bool(spec.GetOidcIssuerEnabled())
	} else {
		clusterArgs.OidcIssuerEnabled = pulumi.Bool(true)
	}
	clusterArgs.WorkloadIdentityEnabled = pulumi.Bool(spec.WorkloadIdentityEnabled)

	clusterArgs.PrivateClusterEnabled = pulumi.Bool(spec.PrivateClusterEnabled)
	if privateDnsZoneId := spec.PrivateDnsZoneId.GetValue(); privateDnsZoneId != "" {
		clusterArgs.PrivateDnsZoneId = pulumi.String(privateDnsZoneId)
	}
	clusterArgs.PrivateClusterPublicFqdnEnabled = pulumi.Bool(spec.PrivateClusterPublicFqdnEnabled)

	if spec.RoleBasedAccessControlEnabled != nil {
		clusterArgs.RoleBasedAccessControlEnabled = pulumi.Bool(spec.GetRoleBasedAccessControlEnabled())
	} else {
		clusterArgs.RoleBasedAccessControlEnabled = pulumi.Bool(true)
	}
	clusterArgs.LocalAccountDisabled = pulumi.Bool(spec.LocalAccountDisabled)

	clusterArgs.AzurePolicyEnabled = pulumi.Bool(spec.AzurePolicyEnabled)

	clusterArgs.ImageCleanerEnabled = pulumi.Bool(spec.ImageCleanerEnabled)
	if spec.ImageCleanerIntervalHours > 0 {
		clusterArgs.ImageCleanerIntervalHours = pulumi.Int(int(spec.ImageCleanerIntervalHours))
	}
	clusterArgs.CostAnalysisEnabled = pulumi.Bool(spec.CostAnalysisEnabled)
	if spec.RunCommandEnabled != nil {
		clusterArgs.RunCommandEnabled = pulumi.Bool(spec.GetRunCommandEnabled())
	} else {
		clusterArgs.RunCommandEnabled = pulumi.Bool(true)
	}

	// The disk encryption set reference resolves to a literal ARM ID before
	// the module runs, so GetValue() returns the resolved id for both a
	// literal and a valueFrom reference.
	if spec.DiskEncryptionSetId != nil && spec.DiskEncryptionSetId.GetValue() != "" {
		clusterArgs.DiskEncryptionSetId = pulumi.String(spec.DiskEncryptionSetId.GetValue())
	}
	if spec.EdgeZone != "" {
		clusterArgs.EdgeZone = pulumi.String(spec.EdgeZone)
	}
	if spec.NodeResourceGroup != "" {
		clusterArgs.NodeResourceGroup = pulumi.String(spec.NodeResourceGroup)
	}
	if len(spec.CustomCaTrustCertificatesBase64) > 0 {
		clusterArgs.CustomCaTrustCertificatesBase64s = pulumi.ToStringArray(spec.CustomCaTrustCertificatesBase64)
	}
	clusterArgs.AiToolchainOperatorEnabled = pulumi.Bool(spec.AiToolchainOperatorEnabled)

	clusterArgs.DefaultNodePool = buildDefaultNodePool(spec, locals)
	clusterArgs.Identity = buildIdentity(spec)

	// The kubelet identity nodes use for image pulls and Azure access. All
	// three fields describe the same user-assigned identity and require a
	// user-assigned cluster identity.
	if spec.KubeletIdentity != nil {
		clusterArgs.KubeletIdentity = &containerservice.KubernetesClusterKubeletIdentityArgs{
			ClientId:               pulumi.String(spec.KubeletIdentity.ClientId),
			ObjectId:               pulumi.String(spec.KubeletIdentity.ObjectId),
			UserAssignedIdentityId: pulumi.String(spec.KubeletIdentity.UserAssignedIdentityId.GetValue()),
		}
	}

	if spec.ApiServerAccessProfile != nil {
		apiServerArgs := &containerservice.KubernetesClusterApiServerAccessProfileArgs{
			VirtualNetworkIntegrationEnabled: pulumi.Bool(spec.ApiServerAccessProfile.VirtualNetworkIntegrationEnabled),
		}
		if len(spec.ApiServerAccessProfile.AuthorizedIpRanges) > 0 {
			apiServerArgs.AuthorizedIpRanges = pulumi.ToStringArray(spec.ApiServerAccessProfile.AuthorizedIpRanges)
		}
		if subnetId := spec.ApiServerAccessProfile.SubnetId.GetValue(); subnetId != "" {
			apiServerArgs.SubnetId = pulumi.String(subnetId)
		}
		clusterArgs.ApiServerAccessProfile = apiServerArgs
	}

	// Entra ID (Azure AD) integration -- cluster admission by AAD group
	// membership, optionally with Azure RBAC as the authorization source.
	if aad := spec.AzureActiveDirectoryRoleBasedAccessControl; aad != nil {
		aadArgs := &containerservice.KubernetesClusterAzureActiveDirectoryRoleBasedAccessControlArgs{
			AzureRbacEnabled: pulumi.Bool(aad.AzureRbacEnabled),
		}
		if aad.TenantId != "" {
			aadArgs.TenantId = pulumi.String(aad.TenantId)
		}
		if len(aad.AdminGroupObjectIds) > 0 {
			aadArgs.AdminGroupObjectIds = pulumi.ToStringArray(aad.AdminGroupObjectIds)
		}
		clusterArgs.AzureActiveDirectoryRoleBasedAccessControl = aadArgs
	}

	clusterArgs.NetworkProfile = buildNetworkProfile(spec)

	if spec.AutoScalerProfile != nil {
		clusterArgs.AutoScalerProfile = buildAutoScalerProfile(spec.AutoScalerProfile)
	}

	buildMaintenanceWindows(spec, clusterArgs)
	buildAddons(spec, clusterArgs)
	buildPlatformProfiles(spec, clusterArgs)

	createdCluster, err := containerservice.NewKubernetesCluster(ctx,
		spec.Name,
		clusterArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create aks cluster %s", spec.Name)
	}

	// Export stack outputs. cluster_id is the parent seam every standalone
	// AzureAksNodePool consumes; oidc_issuer_url is the trust anchor an
	// AzureFederatedIdentityCredential binds to for workload identity.
	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpClusterName, createdCluster.Name)
	ctx.Export(OpFqdn, createdCluster.Fqdn)
	ctx.Export(OpPrivateFqdn, createdCluster.PrivateFqdn)
	ctx.Export(OpPortalFqdn, createdCluster.PortalFqdn)
	ctx.Export(OpOidcIssuerUrl, createdCluster.OidcIssuerUrl)
	ctx.Export(OpNodeResourceGroup, createdCluster.NodeResourceGroup)
	ctx.Export(OpNodeResourceGroupId, createdCluster.NodeResourceGroupId)
	// Base64-encode the raw kubeconfig so both engines export the same
	// shape (the Terraform module base64encodes kube_config_raw).
	ctx.Export(OpClusterKubeconfig, pulumi.ToSecret(createdCluster.KubeConfigRaw.ApplyT(func(raw string) string {
		return base64.StdEncoding.EncodeToString([]byte(raw))
	}).(pulumi.StringOutput)))
	ctx.Export(OpClusterIdentityPrincipalId, createdCluster.Identity.PrincipalId().Elem())
	ctx.Export(OpKubeletIdentityObjectId, createdCluster.KubeletIdentity.ObjectId().Elem())
	ctx.Export(OpKubeletIdentityClientId, createdCluster.KubeletIdentity.ClientId().Elem())
	ctx.Export(OpCurrentKubernetesVersion, createdCluster.CurrentKubernetesVersion)
	// The CA certificate is public cluster identity (the TLS trust anchor), not
	// credential material. The secret flag it inherits from the provider's
	// sensitive kube_config attribute is deliberately unwrapped so the
	// platform's cluster-connection materializer can read it as a plain stack
	// output -- the same posture as the EKS/GKE CA outputs.
	ctx.Export(OpClusterCaCertificate, pulumi.Unsecret(
		createdCluster.KubeConfigs.Index(pulumi.Int(0)).ClusterCaCertificate().Elem()))
	// The applicability gate for token-based cluster connections: only an
	// Entra-integrated API server honors short-lived Entra bearer tokens, so
	// the materializer publishes a connection only when this reads "true".
	ctx.Export(OpEntraIntegrationEnabled,
		pulumi.String(strconv.FormatBool(spec.AzureActiveDirectoryRoleBasedAccessControl != nil)))

	return nil
}

// buildDefaultNodePool maps the spec's default (system) pool -- always
// Linux, always System mode, which is why it carries no os_type/mode/spot
// knobs. Its field shape deliberately matches the standalone
// AzureAksNodePool kind so moving a workload pool out to its own resource
// is a mechanical copy.
func buildDefaultNodePool(spec *azureaksclusterv1alpha1.AzureAksClusterSpec, locals *Locals) *containerservice.KubernetesClusterDefaultNodePoolArgs {
	pool := spec.DefaultNodePool

	poolArgs := &containerservice.KubernetesClusterDefaultNodePoolArgs{
		Name:   pulumi.String(pool.Name),
		VmSize: pulumi.String(pool.VmSize),
	}

	// With autoscaling, ARM owns node_count after creation; without it,
	// node_count is the pool's fixed size (spec validation requires it).
	if pool.NodeCount > 0 {
		poolArgs.NodeCount = pulumi.Int(int(pool.NodeCount))
	}
	poolArgs.AutoScalingEnabled = pulumi.Bool(pool.AutoScalingEnabled)
	if pool.AutoScalingEnabled {
		poolArgs.MinCount = pulumi.Int(int(pool.MinCount))
		poolArgs.MaxCount = pulumi.Int(int(pool.MaxCount))
	}

	if pool.MaxPods > 0 {
		poolArgs.MaxPods = pulumi.Int(int(pool.MaxPods))
	}
	if len(pool.Zones) > 0 {
		poolArgs.Zones = pulumi.ToStringArray(pool.Zones)
	}

	// Unset deploys AKS-managed networking (Azure's default); a BYO
	// subnet requires the cluster identity to hold Network Contributor.
	if subnetId := pool.VnetSubnetId.GetValue(); subnetId != "" {
		poolArgs.VnetSubnetId = pulumi.String(subnetId)
	}
	if podSubnetId := pool.PodSubnetId.GetValue(); podSubnetId != "" {
		poolArgs.PodSubnetId = pulumi.String(podSubnetId)
	}

	if pool.OsDiskSizeGb > 0 {
		poolArgs.OsDiskSizeGb = pulumi.Int(int(pool.OsDiskSizeGb))
	}
	if v := osDiskTypeToArm(pool.OsDiskType); v != "" {
		poolArgs.OsDiskType = pulumi.String(v)
	} else {
		poolArgs.OsDiskType = pulumi.String("Managed")
	}
	if v := kubeletDiskTypeToArm(pool.KubeletDiskType); v != "" {
		poolArgs.KubeletDiskType = pulumi.String(v)
	}
	if v := osSkuToArm(pool.OsSku); v != "" {
		poolArgs.OsSku = pulumi.String(v)
	}
	if pool.OrchestratorVersion != "" {
		poolArgs.OrchestratorVersion = pulumi.String(pool.OrchestratorVersion)
	}

	if len(pool.NodeLabels) > 0 {
		poolArgs.NodeLabels = pulumi.ToStringMap(pool.NodeLabels)
	}
	poolArgs.OnlyCriticalAddonsEnabled = pulumi.Bool(pool.OnlyCriticalAddonsEnabled)

	poolArgs.FipsEnabled = pulumi.Bool(pool.FipsEnabled)
	poolArgs.HostEncryptionEnabled = pulumi.Bool(pool.HostEncryptionEnabled)

	poolArgs.NodePublicIpEnabled = pulumi.Bool(pool.NodePublicIpEnabled)
	if prefixId := pool.NodePublicIpPrefixId.GetValue(); prefixId != "" {
		poolArgs.NodePublicIpPrefixId = pulumi.String(prefixId)
	}

	if v := gpuInstanceToArm(pool.GpuInstance); v != "" {
		poolArgs.GpuInstance = pulumi.String(v)
	}
	if v := gpuDriverToArm(pool.GpuDriver); v != "" {
		poolArgs.GpuDriver = pulumi.String(v)
	}

	if pool.ProximityPlacementGroupId != "" {
		poolArgs.ProximityPlacementGroupId = pulumi.String(pool.ProximityPlacementGroupId)
	}
	if pool.HostGroupId != "" {
		poolArgs.HostGroupId = pulumi.String(pool.HostGroupId)
	}
	if pool.CapacityReservationGroupId != "" {
		poolArgs.CapacityReservationGroupId = pulumi.String(pool.CapacityReservationGroupId)
	}

	if v := scaleDownModeToArm(pool.ScaleDownMode); v != "" {
		poolArgs.ScaleDownMode = pulumi.String(v)
	} else {
		poolArgs.ScaleDownMode = pulumi.String("Delete")
	}
	if pool.SnapshotId != "" {
		poolArgs.SnapshotId = pulumi.String(pool.SnapshotId)
	}
	if v := workloadRuntimeToArm(pool.WorkloadRuntime); v != "" {
		poolArgs.WorkloadRuntime = pulumi.String(v)
	}

	poolArgs.UltraSsdEnabled = pulumi.Bool(pool.UltraSsdEnabled)

	// A stand-in pool AKS rotates through otherwise replace-forcing
	// changes (vm_size, os_disk_type...) -- set proactively in production.
	if pool.TemporaryNameForRotation != "" {
		poolArgs.TemporaryNameForRotation = pulumi.String(pool.TemporaryNameForRotation)
	}

	if pool.KubeletConfig != nil {
		poolArgs.KubeletConfig = buildDefaultPoolKubeletConfig(pool.KubeletConfig)
	}
	if pool.LinuxOsConfig != nil {
		poolArgs.LinuxOsConfig = buildDefaultPoolLinuxOsConfig(pool.LinuxOsConfig)
	}
	if pool.NodeNetworkProfile != nil {
		poolArgs.NodeNetworkProfile = buildDefaultPoolNodeNetworkProfile(pool.NodeNetworkProfile)
	}
	if pool.UpgradeSettings != nil {
		upgradeArgs := &containerservice.KubernetesClusterDefaultNodePoolUpgradeSettingsArgs{
			MaxSurge: pulumi.String(pool.UpgradeSettings.MaxSurge),
		}
		if pool.UpgradeSettings.DrainTimeoutInMinutes > 0 {
			upgradeArgs.DrainTimeoutInMinutes = pulumi.Int(int(pool.UpgradeSettings.DrainTimeoutInMinutes))
		}
		if pool.UpgradeSettings.NodeSoakDurationInMinutes > 0 {
			upgradeArgs.NodeSoakDurationInMinutes = pulumi.Int(int(pool.UpgradeSettings.NodeSoakDurationInMinutes))
		}
		if v := undrainableNodeBehaviorToArm(pool.UpgradeSettings.UndrainableNodeBehavior); v != "" {
			upgradeArgs.UndrainableNodeBehavior = pulumi.String(v)
		}
		poolArgs.UpgradeSettings = upgradeArgs
	}

	// Pool tags: the cluster-wide merged tags, with the pool's own user
	// tags merged over them (pool tags win) -- mirroring the Terraform
	// module's merge order.
	poolTags := map[string]string{}
	for k, v := range locals.AzureTags {
		poolTags[k] = v
	}
	for k, v := range pool.Tags {
		poolTags[k] = v
	}
	poolArgs.Tags = pulumi.ToStringMap(poolTags)

	return poolArgs
}

func buildDefaultPoolKubeletConfig(cfg *azureaksclusterv1alpha1.AzureAksClusterKubeletConfig) *containerservice.KubernetesClusterDefaultNodePoolKubeletConfigArgs {
	args := &containerservice.KubernetesClusterDefaultNodePoolKubeletConfigArgs{}
	if v := cpuManagerPolicyToArm(cfg.CpuManagerPolicy); v != "" {
		args.CpuManagerPolicy = pulumi.String(v)
	}
	// Presence-guarded: the optional bool defaults to true in the proto;
	// an absent value falls back to true -- matching the Terraform
	// module's optional(bool, true).
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

func buildDefaultPoolLinuxOsConfig(cfg *azureaksclusterv1alpha1.AzureAksClusterLinuxOsConfig) *containerservice.KubernetesClusterDefaultNodePoolLinuxOsConfigArgs {
	args := &containerservice.KubernetesClusterDefaultNodePoolLinuxOsConfigArgs{}
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
		args.SysctlConfig = buildDefaultPoolSysctlConfig(cfg.SysctlConfig)
	}
	return args
}

// buildDefaultPoolSysctlConfig maps only the sysctls the spec sets (zero
// means "keep the AKS default"), so an untouched manifest never overrides
// node kernel settings.
func buildDefaultPoolSysctlConfig(cfg *azureaksclusterv1alpha1.AzureAksClusterSysctlConfig) *containerservice.KubernetesClusterDefaultNodePoolLinuxOsConfigSysctlConfigArgs {
	args := &containerservice.KubernetesClusterDefaultNodePoolLinuxOsConfigSysctlConfigArgs{}
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

func buildDefaultPoolNodeNetworkProfile(cfg *azureaksclusterv1alpha1.AzureAksClusterNodeNetworkProfile) *containerservice.KubernetesClusterDefaultNodePoolNodeNetworkProfileArgs {
	args := &containerservice.KubernetesClusterDefaultNodePoolNodeNetworkProfileArgs{}
	if len(cfg.AllowedHostPorts) > 0 {
		hostPorts := containerservice.KubernetesClusterDefaultNodePoolNodeNetworkProfileAllowedHostPortArray{}
		for _, hostPort := range cfg.AllowedHostPorts {
			hostPortArgs := containerservice.KubernetesClusterDefaultNodePoolNodeNetworkProfileAllowedHostPortArgs{}
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

// buildIdentity maps the cluster's managed identity: system-assigned
// unless the spec brings a user-assigned identity (needed when grants
// must pre-exist -- BYO private DNS zone, BYO subnet).
func buildIdentity(spec *azureaksclusterv1alpha1.AzureAksClusterSpec) *containerservice.KubernetesClusterIdentityArgs {
	if spec.Identity != nil && spec.Identity.Type == azureaksclusterv1alpha1.AzureAksClusterIdentityType_USER_ASSIGNED {
		identityIds := []string{}
		for _, identityId := range spec.Identity.IdentityIds {
			identityIds = append(identityIds, identityId.GetValue())
		}
		return &containerservice.KubernetesClusterIdentityArgs{
			Type:        pulumi.String("UserAssigned"),
			IdentityIds: pulumi.ToStringArray(identityIds),
		}
	}
	return &containerservice.KubernetesClusterIdentityArgs{
		Type: pulumi.String("SystemAssigned"),
	}
}

// buildNetworkProfile writes the network fabric explicitly even when the
// spec leaves it unset: the provider's implicit fallback is deprecated
// kubenet, while the modern AKS default is Azure CNI overlay -- the
// module makes the good default the actual default.
func buildNetworkProfile(spec *azureaksclusterv1alpha1.AzureAksClusterSpec) *containerservice.KubernetesClusterNetworkProfileArgs {
	profile := spec.NetworkProfile

	networkPlugin := "azure"
	if profile != nil {
		switch profile.NetworkPlugin {
		case azureaksclusterv1alpha1.AzureAksClusterNetworkPlugin_KUBENET:
			networkPlugin = "kubenet"
		case azureaksclusterv1alpha1.AzureAksClusterNetworkPlugin_NETWORK_PLUGIN_NONE:
			networkPlugin = "none"
		}
	}

	args := &containerservice.KubernetesClusterNetworkProfileArgs{
		NetworkPlugin: pulumi.String(networkPlugin),
	}

	// Overlay applies when explicitly chosen, or by default for Azure CNI
	// when no pod subnet points at traditional dynamic allocation.
	if networkPlugin == "azure" {
		explicitOverlay := profile != nil && profile.NetworkPluginMode == azureaksclusterv1alpha1.AzureAksClusterNetworkPluginMode_OVERLAY
		defaultOverlay := (profile == nil || profile.NetworkPluginMode == azureaksclusterv1alpha1.AzureAksClusterNetworkPluginMode(0)) &&
			spec.DefaultNodePool.PodSubnetId.GetValue() == ""
		if explicitOverlay || defaultOverlay {
			args.NetworkPluginMode = pulumi.String("overlay")
		}
	}

	outboundType := "loadBalancer"
	if profile != nil {
		switch profile.NetworkPolicy {
		case azureaksclusterv1alpha1.AzureAksClusterNetworkPolicy_NETWORK_POLICY_AZURE:
			args.NetworkPolicy = pulumi.String("azure")
		case azureaksclusterv1alpha1.AzureAksClusterNetworkPolicy_CALICO:
			args.NetworkPolicy = pulumi.String("calico")
		case azureaksclusterv1alpha1.AzureAksClusterNetworkPolicy_NETWORK_POLICY_CILIUM:
			args.NetworkPolicy = pulumi.String("cilium")
		}

		switch profile.NetworkDataPlane {
		case azureaksclusterv1alpha1.AzureAksClusterNetworkDataPlane_DATA_PLANE_AZURE:
			args.NetworkDataPlane = pulumi.String("azure")
		case azureaksclusterv1alpha1.AzureAksClusterNetworkDataPlane_DATA_PLANE_CILIUM:
			args.NetworkDataPlane = pulumi.String("cilium")
		}

		if profile.DnsServiceIp != "" {
			args.DnsServiceIp = pulumi.String(profile.DnsServiceIp)
		}
		if profile.ServiceCidr != "" {
			args.ServiceCidr = pulumi.String(profile.ServiceCidr)
		}
		if len(profile.ServiceCidrs) > 0 {
			args.ServiceCidrs = pulumi.ToStringArray(profile.ServiceCidrs)
		}
		if profile.PodCidr != "" {
			args.PodCidr = pulumi.String(profile.PodCidr)
		}
		if len(profile.PodCidrs) > 0 {
			args.PodCidrs = pulumi.ToStringArray(profile.PodCidrs)
		}
		if len(profile.IpVersions) > 0 {
			ipVersions := []string{}
			for _, ipVersion := range profile.IpVersions {
				if ipVersion == azureaksclusterv1alpha1.AzureAksClusterIpVersion_IPV6 {
					ipVersions = append(ipVersions, "IPv6")
				} else {
					ipVersions = append(ipVersions, "IPv4")
				}
			}
			args.IpVersions = pulumi.ToStringArray(ipVersions)
		}

		switch profile.OutboundType {
		case azureaksclusterv1alpha1.AzureAksClusterOutboundType_MANAGED_NAT_GATEWAY:
			outboundType = "managedNATGateway"
		case azureaksclusterv1alpha1.AzureAksClusterOutboundType_USER_ASSIGNED_NAT_GATEWAY:
			outboundType = "userAssignedNATGateway"
		case azureaksclusterv1alpha1.AzureAksClusterOutboundType_USER_DEFINED_ROUTING:
			outboundType = "userDefinedRouting"
		case azureaksclusterv1alpha1.AzureAksClusterOutboundType_OUTBOUND_NONE:
			outboundType = "none"
		}

		if lb := profile.LoadBalancerProfile; lb != nil {
			lbArgs := &containerservice.KubernetesClusterNetworkProfileLoadBalancerProfileArgs{}
			if lb.OutboundPortsAllocated > 0 {
				lbArgs.OutboundPortsAllocated = pulumi.Int(int(lb.OutboundPortsAllocated))
			}
			if lb.IdleTimeoutInMinutes > 0 {
				lbArgs.IdleTimeoutInMinutes = pulumi.Int(int(lb.IdleTimeoutInMinutes))
			}
			if lb.ManagedOutboundIpCount > 0 {
				lbArgs.ManagedOutboundIpCount = pulumi.Int(int(lb.ManagedOutboundIpCount))
			}
			if lb.ManagedOutboundIpv6Count > 0 {
				lbArgs.ManagedOutboundIpv6Count = pulumi.Int(int(lb.ManagedOutboundIpv6Count))
			}
			if len(lb.OutboundIpPrefixIds) > 0 {
				prefixIds := []string{}
				for _, prefixId := range lb.OutboundIpPrefixIds {
					prefixIds = append(prefixIds, prefixId.GetValue())
				}
				lbArgs.OutboundIpPrefixIds = pulumi.ToStringArray(prefixIds)
			}
			if len(lb.OutboundIpAddressIds) > 0 {
				addressIds := []string{}
				for _, addressId := range lb.OutboundIpAddressIds {
					addressIds = append(addressIds, addressId.GetValue())
				}
				lbArgs.OutboundIpAddressIds = pulumi.ToStringArray(addressIds)
			}
			switch lb.BackendPoolType {
			case azureaksclusterv1alpha1.AzureAksClusterLoadBalancerBackendPoolType_NODE_IP_CONFIGURATION:
				lbArgs.BackendPoolType = pulumi.String("NodeIPConfiguration")
			case azureaksclusterv1alpha1.AzureAksClusterLoadBalancerBackendPoolType_NODE_IP:
				lbArgs.BackendPoolType = pulumi.String("NodeIP")
			}
			args.LoadBalancerProfile = lbArgs
		}

		if nat := profile.NatGatewayProfile; nat != nil {
			natArgs := &containerservice.KubernetesClusterNetworkProfileNatGatewayProfileArgs{}
			if nat.IdleTimeoutInMinutes > 0 {
				natArgs.IdleTimeoutInMinutes = pulumi.Int(int(nat.IdleTimeoutInMinutes))
			}
			if nat.ManagedOutboundIpCount > 0 {
				natArgs.ManagedOutboundIpCount = pulumi.Int(int(nat.ManagedOutboundIpCount))
			}
			args.NatGatewayProfile = natArgs
		}

		if advancedNetworking := profile.AdvancedNetworking; advancedNetworking != nil {
			args.AdvancedNetworking = &containerservice.KubernetesClusterNetworkProfileAdvancedNetworkingArgs{
				ObservabilityEnabled: pulumi.Bool(advancedNetworking.ObservabilityEnabled),
				SecurityEnabled:      pulumi.Bool(advancedNetworking.SecurityEnabled),
			}
		}
	}
	args.OutboundType = pulumi.String(outboundType)

	return args
}

func buildAutoScalerProfile(profile *azureaksclusterv1alpha1.AzureAksClusterAutoScalerProfile) *containerservice.KubernetesClusterAutoScalerProfileArgs {
	args := &containerservice.KubernetesClusterAutoScalerProfileArgs{
		BalanceSimilarNodeGroups:              pulumi.Bool(profile.BalanceSimilarNodeGroups),
		DaemonsetEvictionForEmptyNodesEnabled: pulumi.Bool(profile.DaemonsetEvictionForEmptyNodesEnabled),
		IgnoreDaemonsetsUtilizationEnabled:    pulumi.Bool(profile.IgnoreDaemonsetsUtilizationEnabled),
		SkipNodesWithLocalStorage:             pulumi.Bool(profile.SkipNodesWithLocalStorage),
	}
	// Presence-guarded true-default optional bools: an absent value falls
	// back to Azure's default (true), matching the Terraform module.
	if profile.DaemonsetEvictionForOccupiedNodesEnabled != nil {
		args.DaemonsetEvictionForOccupiedNodesEnabled = pulumi.Bool(profile.GetDaemonsetEvictionForOccupiedNodesEnabled())
	} else {
		args.DaemonsetEvictionForOccupiedNodesEnabled = pulumi.Bool(true)
	}
	if profile.SkipNodesWithSystemPods != nil {
		args.SkipNodesWithSystemPods = pulumi.Bool(profile.GetSkipNodesWithSystemPods())
	} else {
		args.SkipNodesWithSystemPods = pulumi.Bool(true)
	}

	switch profile.Expander {
	case azureaksclusterv1alpha1.AzureAksClusterAutoscalerExpander_LEAST_WASTE:
		args.Expander = pulumi.String("least-waste")
	case azureaksclusterv1alpha1.AzureAksClusterAutoscalerExpander_MOST_PODS:
		args.Expander = pulumi.String("most-pods")
	case azureaksclusterv1alpha1.AzureAksClusterAutoscalerExpander_PRIORITY:
		args.Expander = pulumi.String("priority")
	case azureaksclusterv1alpha1.AzureAksClusterAutoscalerExpander_RANDOM:
		args.Expander = pulumi.String("random")
	}

	if profile.MaxGracefulTerminationSec > 0 {
		args.MaxGracefulTerminationSec = pulumi.String(itoa(profile.MaxGracefulTerminationSec))
	}
	if profile.MaxNodeProvisioningTime != "" {
		args.MaxNodeProvisioningTime = pulumi.String(profile.MaxNodeProvisioningTime)
	}
	if profile.MaxUnreadyNodes > 0 {
		args.MaxUnreadyNodes = pulumi.Int(int(profile.MaxUnreadyNodes))
	}
	if profile.MaxUnreadyPercentage > 0 {
		args.MaxUnreadyPercentage = pulumi.Float64(float64(profile.MaxUnreadyPercentage))
	}
	if profile.NewPodScaleUpDelay != "" {
		args.NewPodScaleUpDelay = pulumi.String(profile.NewPodScaleUpDelay)
	}
	if profile.ScanInterval != "" {
		args.ScanInterval = pulumi.String(profile.ScanInterval)
	}
	if profile.ScaleDownDelayAfterAdd != "" {
		args.ScaleDownDelayAfterAdd = pulumi.String(profile.ScaleDownDelayAfterAdd)
	}
	if profile.ScaleDownDelayAfterDelete != "" {
		args.ScaleDownDelayAfterDelete = pulumi.String(profile.ScaleDownDelayAfterDelete)
	}
	if profile.ScaleDownDelayAfterFailure != "" {
		args.ScaleDownDelayAfterFailure = pulumi.String(profile.ScaleDownDelayAfterFailure)
	}
	if profile.ScaleDownUnneeded != "" {
		args.ScaleDownUnneeded = pulumi.String(profile.ScaleDownUnneeded)
	}
	if profile.ScaleDownUnready != "" {
		args.ScaleDownUnready = pulumi.String(profile.ScaleDownUnready)
	}
	if profile.ScaleDownUtilizationThreshold != "" {
		args.ScaleDownUtilizationThreshold = pulumi.String(profile.ScaleDownUtilizationThreshold)
	}
	if profile.EmptyBulkDeleteMax > 0 {
		args.EmptyBulkDeleteMax = pulumi.String(itoa(profile.EmptyBulkDeleteMax))
	}
	return args
}

// buildMaintenanceWindows maps the three maintenance surfaces: the legacy
// hour-of-week window for routine maintenance, and the two schedule-based
// windows that govern WHEN the upgrade channels apply their work.
func buildMaintenanceWindows(spec *azureaksclusterv1alpha1.AzureAksClusterSpec, clusterArgs *containerservice.KubernetesClusterArgs) {
	if mw := spec.MaintenanceWindow; mw != nil {
		mwArgs := &containerservice.KubernetesClusterMaintenanceWindowArgs{}
		alloweds := containerservice.KubernetesClusterMaintenanceWindowAllowedArray{}
		for _, allowed := range mw.Allowed {
			hours := []int{}
			for _, h := range allowed.Hours {
				hours = append(hours, int(h))
			}
			alloweds = append(alloweds, containerservice.KubernetesClusterMaintenanceWindowAllowedArgs{
				Day:   pulumi.String(weekDayToArm(allowed.Day)),
				Hours: pulumi.ToIntArray(hours),
			})
		}
		if len(alloweds) > 0 {
			mwArgs.Alloweds = alloweds
		}
		notAlloweds := containerservice.KubernetesClusterMaintenanceWindowNotAllowedArray{}
		for _, notAllowed := range mw.NotAllowed {
			notAlloweds = append(notAlloweds, containerservice.KubernetesClusterMaintenanceWindowNotAllowedArgs{
				Start: pulumi.String(notAllowed.Start),
				End:   pulumi.String(notAllowed.End),
			})
		}
		if len(notAlloweds) > 0 {
			mwArgs.NotAlloweds = notAlloweds
		}
		clusterArgs.MaintenanceWindow = mwArgs
	}

	if schedule := spec.MaintenanceWindowAutoUpgrade; schedule != nil {
		scheduleArgs := &containerservice.KubernetesClusterMaintenanceWindowAutoUpgradeArgs{
			Frequency: pulumi.String(frequencyToArm(schedule.Frequency)),
			Interval:  pulumi.Int(int(schedule.Interval)),
			Duration:  pulumi.Int(int(schedule.Duration)),
		}
		if v := weekDayToArm(schedule.DayOfWeek); v != "" {
			scheduleArgs.DayOfWeek = pulumi.String(v)
		}
		if v := weekIndexToArm(schedule.WeekIndex); v != "" {
			scheduleArgs.WeekIndex = pulumi.String(v)
		}
		if schedule.DayOfMonth > 0 {
			scheduleArgs.DayOfMonth = pulumi.Int(int(schedule.DayOfMonth))
		}
		if schedule.StartDate != "" {
			scheduleArgs.StartDate = pulumi.String(schedule.StartDate)
		}
		if schedule.StartTime != "" {
			scheduleArgs.StartTime = pulumi.String(schedule.StartTime)
		}
		if schedule.UtcOffset != "" {
			scheduleArgs.UtcOffset = pulumi.String(schedule.UtcOffset)
		}
		notAlloweds := containerservice.KubernetesClusterMaintenanceWindowAutoUpgradeNotAllowedArray{}
		for _, notAllowed := range schedule.NotAllowed {
			notAlloweds = append(notAlloweds, containerservice.KubernetesClusterMaintenanceWindowAutoUpgradeNotAllowedArgs{
				Start: pulumi.String(notAllowed.Start),
				End:   pulumi.String(notAllowed.End),
			})
		}
		if len(notAlloweds) > 0 {
			scheduleArgs.NotAlloweds = notAlloweds
		}
		clusterArgs.MaintenanceWindowAutoUpgrade = scheduleArgs
	}

	if schedule := spec.MaintenanceWindowNodeOs; schedule != nil {
		scheduleArgs := &containerservice.KubernetesClusterMaintenanceWindowNodeOsArgs{
			Frequency: pulumi.String(frequencyToArm(schedule.Frequency)),
			Interval:  pulumi.Int(int(schedule.Interval)),
			Duration:  pulumi.Int(int(schedule.Duration)),
		}
		if v := weekDayToArm(schedule.DayOfWeek); v != "" {
			scheduleArgs.DayOfWeek = pulumi.String(v)
		}
		if v := weekIndexToArm(schedule.WeekIndex); v != "" {
			scheduleArgs.WeekIndex = pulumi.String(v)
		}
		if schedule.DayOfMonth > 0 {
			scheduleArgs.DayOfMonth = pulumi.Int(int(schedule.DayOfMonth))
		}
		if schedule.StartDate != "" {
			scheduleArgs.StartDate = pulumi.String(schedule.StartDate)
		}
		if schedule.StartTime != "" {
			scheduleArgs.StartTime = pulumi.String(schedule.StartTime)
		}
		if schedule.UtcOffset != "" {
			scheduleArgs.UtcOffset = pulumi.String(schedule.UtcOffset)
		}
		notAlloweds := containerservice.KubernetesClusterMaintenanceWindowNodeOsNotAllowedArray{}
		for _, notAllowed := range schedule.NotAllowed {
			notAlloweds = append(notAlloweds, containerservice.KubernetesClusterMaintenanceWindowNodeOsNotAllowedArgs{
				Start: pulumi.String(notAllowed.Start),
				End:   pulumi.String(notAllowed.End),
			})
		}
		if len(notAlloweds) > 0 {
			scheduleArgs.NotAlloweds = notAlloweds
		}
		clusterArgs.MaintenanceWindowNodeOs = scheduleArgs
	}
}

// buildAddons maps the Azure-managed add-ons. Each block only renders
// when the spec configures it, so an unset spec and Azure's addon-off
// default deploy identically.
func buildAddons(spec *azureaksclusterv1alpha1.AzureAksClusterSpec, clusterArgs *containerservice.KubernetesClusterArgs) {
	if oms := spec.OmsAgent; oms != nil {
		clusterArgs.OmsAgent = &containerservice.KubernetesClusterOmsAgentArgs{
			LogAnalyticsWorkspaceId:     pulumi.String(oms.LogAnalyticsWorkspaceId.GetValue()),
			MsiAuthForMonitoringEnabled: pulumi.Bool(oms.MsiAuthForMonitoringEnabled),
		}
	}

	// Each add-on block carries its own switch (on by default once declared,
	// filled by the platform before this runs); the switch decides whether
	// the add-on renders.
	if kvProvider := spec.KeyVaultSecretsProvider; kvProvider != nil && kvProvider.GetEnabled() {
		kvArgs := &containerservice.KubernetesClusterKeyVaultSecretsProviderArgs{
			SecretRotationEnabled: pulumi.Bool(kvProvider.SecretRotationEnabled),
		}
		if kvProvider.SecretRotationInterval != "" {
			kvArgs.SecretRotationInterval = pulumi.String(kvProvider.SecretRotationInterval)
		}
		clusterArgs.KeyVaultSecretsProvider = kvArgs
	}

	if defender := spec.MicrosoftDefender; defender != nil {
		clusterArgs.MicrosoftDefender = &containerservice.KubernetesClusterMicrosoftDefenderArgs{
			LogAnalyticsWorkspaceId: pulumi.String(defender.LogAnalyticsWorkspaceId.GetValue()),
		}
	}

	if metrics := spec.MonitorMetrics; metrics != nil && metrics.GetEnabled() {
		metricsArgs := &containerservice.KubernetesClusterMonitorMetricsArgs{}
		if metrics.AnnotationsAllowed != "" {
			metricsArgs.AnnotationsAllowed = pulumi.String(metrics.AnnotationsAllowed)
		}
		if metrics.LabelsAllowed != "" {
			metricsArgs.LabelsAllowed = pulumi.String(metrics.LabelsAllowed)
		}
		clusterArgs.MonitorMetrics = metricsArgs
	}

	if agic := spec.IngressApplicationGateway; agic != nil {
		agicArgs := &containerservice.KubernetesClusterIngressApplicationGatewayArgs{}
		if gatewayId := agic.GatewayId.GetValue(); gatewayId != "" {
			agicArgs.GatewayId = pulumi.String(gatewayId)
		}
		if agic.GatewayName != "" {
			agicArgs.GatewayName = pulumi.String(agic.GatewayName)
		}
		if agic.SubnetCidr != "" {
			agicArgs.SubnetCidr = pulumi.String(agic.SubnetCidr)
		}
		if subnetId := agic.SubnetId.GetValue(); subnetId != "" {
			agicArgs.SubnetId = pulumi.String(subnetId)
		}
		clusterArgs.IngressApplicationGateway = agicArgs
	}

	if aci := spec.AciConnectorLinux; aci != nil {
		clusterArgs.AciConnectorLinux = &containerservice.KubernetesClusterAciConnectorLinuxArgs{
			SubnetName: pulumi.String(aci.SubnetName),
		}
	}

	if confidentialComputing := spec.ConfidentialComputing; confidentialComputing != nil && confidentialComputing.GetEnabled() {
		clusterArgs.ConfidentialComputing = &containerservice.KubernetesClusterConfidentialComputingArgs{
			SgxQuoteHelperEnabled: pulumi.Bool(confidentialComputing.SgxQuoteHelperEnabled),
		}
	}

	if webAppRouting := spec.WebAppRouting; webAppRouting != nil {
		dnsZoneIds := []string{}
		for _, dnsZoneId := range webAppRouting.DnsZoneIds {
			dnsZoneIds = append(dnsZoneIds, dnsZoneId.GetValue())
		}
		routingArgs := &containerservice.KubernetesClusterWebAppRoutingArgs{
			DnsZoneIds: pulumi.ToStringArray(dnsZoneIds),
		}
		if v := nginxControllerToArm(webAppRouting.DefaultNginxController); v != "" {
			routingArgs.DefaultNginxController = pulumi.String(v)
		}
		clusterArgs.WebAppRouting = routingArgs
	}
}

// buildPlatformProfiles maps the platform-level profiles: service mesh,
// storage drivers, workload autoscalers, KMS etcd encryption, proxy,
// node access credentials, bootstrap source, node auto-provisioning, and
// the upgrade override.
func buildPlatformProfiles(spec *azureaksclusterv1alpha1.AzureAksClusterSpec, clusterArgs *containerservice.KubernetesClusterArgs) {
	if mesh := spec.ServiceMeshProfile; mesh != nil {
		meshArgs := &containerservice.KubernetesClusterServiceMeshProfileArgs{
			Mode:                          pulumi.String("Istio"),
			Revisions:                     pulumi.ToStringArray(mesh.Revisions),
			InternalIngressGatewayEnabled: pulumi.Bool(mesh.InternalIngressGatewayEnabled),
			ExternalIngressGatewayEnabled: pulumi.Bool(mesh.ExternalIngressGatewayEnabled),
		}
		if ca := mesh.CertificateAuthority; ca != nil {
			meshArgs.CertificateAuthority = &containerservice.KubernetesClusterServiceMeshProfileCertificateAuthorityArgs{
				KeyVaultId:          pulumi.String(ca.KeyVaultId.GetValue()),
				RootCertObjectName:  pulumi.String(ca.RootCertObjectName),
				CertChainObjectName: pulumi.String(ca.CertChainObjectName),
				CertObjectName:      pulumi.String(ca.CertObjectName),
				KeyObjectName:       pulumi.String(ca.KeyObjectName),
			}
		}
		clusterArgs.ServiceMeshProfile = meshArgs
	}

	if storage := spec.StorageProfile; storage != nil {
		storageArgs := &containerservice.KubernetesClusterStorageProfileArgs{
			BlobDriverEnabled: pulumi.Bool(storage.BlobDriverEnabled),
		}
		// Presence-guarded true-default optional bools: an absent value
		// falls back to Azure's default (true), matching the Terraform
		// module's optional(bool, true).
		if storage.DiskDriverEnabled != nil {
			storageArgs.DiskDriverEnabled = pulumi.Bool(storage.GetDiskDriverEnabled())
		} else {
			storageArgs.DiskDriverEnabled = pulumi.Bool(true)
		}
		if storage.FileDriverEnabled != nil {
			storageArgs.FileDriverEnabled = pulumi.Bool(storage.GetFileDriverEnabled())
		} else {
			storageArgs.FileDriverEnabled = pulumi.Bool(true)
		}
		if storage.SnapshotControllerEnabled != nil {
			storageArgs.SnapshotControllerEnabled = pulumi.Bool(storage.GetSnapshotControllerEnabled())
		} else {
			storageArgs.SnapshotControllerEnabled = pulumi.Bool(true)
		}
		clusterArgs.StorageProfile = storageArgs
	}

	if autoscaler := spec.WorkloadAutoscalerProfile; autoscaler != nil {
		clusterArgs.WorkloadAutoscalerProfile = &containerservice.KubernetesClusterWorkloadAutoscalerProfileArgs{
			KedaEnabled:                  pulumi.Bool(autoscaler.KedaEnabled),
			VerticalPodAutoscalerEnabled: pulumi.Bool(autoscaler.VerticalPodAutoscalerEnabled),
		}
	}

	if kms := spec.KeyManagementService; kms != nil {
		kmsArgs := &containerservice.KubernetesClusterKeyManagementServiceArgs{
			KeyVaultKeyId: pulumi.String(kms.KeyVaultKeyId.GetValue()),
		}
		switch kms.KeyVaultNetworkAccess {
		case azureaksclusterv1alpha1.AzureAksClusterKeyVaultNetworkAccess_KMS_PRIVATE:
			kmsArgs.KeyVaultNetworkAccess = pulumi.String("Private")
		default:
			kmsArgs.KeyVaultNetworkAccess = pulumi.String("Public")
		}
		clusterArgs.KeyManagementService = kmsArgs
	}

	if proxy := spec.HttpProxyConfig; proxy != nil {
		proxyArgs := &containerservice.KubernetesClusterHttpProxyConfigArgs{}
		if proxy.HttpProxy != "" {
			proxyArgs.HttpProxy = pulumi.String(proxy.HttpProxy)
		}
		if proxy.HttpsProxy != "" {
			proxyArgs.HttpsProxy = pulumi.String(proxy.HttpsProxy)
		}
		if len(proxy.NoProxy) > 0 {
			proxyArgs.NoProxies = pulumi.ToStringArray(proxy.NoProxy)
		}
		if proxy.TrustedCa != "" {
			proxyArgs.TrustedCa = pulumi.String(proxy.TrustedCa)
		}
		clusterArgs.HttpProxyConfig = proxyArgs
	}

	if linux := spec.LinuxProfile; linux != nil {
		clusterArgs.LinuxProfile = &containerservice.KubernetesClusterLinuxProfileArgs{
			AdminUsername: pulumi.String(linux.AdminUsername),
			SshKey: &containerservice.KubernetesClusterLinuxProfileSshKeyArgs{
				KeyData: pulumi.String(linux.SshPublicKey),
			},
		}
	}

	// Windows credentials -- the prerequisite for any Windows
	// AzureAksNodePool joining this cluster.
	if windows := spec.WindowsProfile; windows != nil {
		windowsArgs := &containerservice.KubernetesClusterWindowsProfileArgs{
			AdminUsername: pulumi.String(windows.AdminUsername),
			AdminPassword: pulumi.String(windows.AdminPassword),
		}
		if windows.License == azureaksclusterv1alpha1.AzureAksClusterWindowsLicense_WINDOWS_SERVER {
			windowsArgs.License = pulumi.String("Windows_Server")
		}
		if gmsa := windows.Gmsa; gmsa != nil {
			windowsArgs.Gmsa = &containerservice.KubernetesClusterWindowsProfileGmsaArgs{
				DnsServer:  pulumi.String(gmsa.DnsServer),
				RootDomain: pulumi.String(gmsa.RootDomain),
			}
		}
		clusterArgs.WindowsProfile = windowsArgs
	}

	if bootstrap := spec.BootstrapProfile; bootstrap != nil {
		bootstrapArgs := &containerservice.KubernetesClusterBootstrapProfileArgs{}
		if bootstrap.ArtifactSource == azureaksclusterv1alpha1.AzureAksClusterBootstrapArtifactSource_CACHE {
			bootstrapArgs.ArtifactSource = pulumi.String("Cache")
		} else {
			bootstrapArgs.ArtifactSource = pulumi.String("Direct")
		}
		if registryId := bootstrap.ContainerRegistryId.GetValue(); registryId != "" {
			bootstrapArgs.ContainerRegistryId = pulumi.String(registryId)
		}
		clusterArgs.BootstrapProfile = bootstrapArgs
	}

	if nodeProvisioning := spec.NodeProvisioningProfile; nodeProvisioning != nil {
		provisioningArgs := &containerservice.KubernetesClusterNodeProvisioningProfileArgs{}
		if nodeProvisioning.Mode == azureaksclusterv1alpha1.AzureAksClusterNodeProvisioningMode_AUTO {
			provisioningArgs.Mode = pulumi.String("Auto")
		} else {
			provisioningArgs.Mode = pulumi.String("Manual")
		}
		if nodeProvisioning.DefaultNodePools == azureaksclusterv1alpha1.AzureAksClusterNodeProvisioningDefaultPools_NODE_POOLS_NONE {
			provisioningArgs.DefaultNodePools = pulumi.String("None")
		} else {
			provisioningArgs.DefaultNodePools = pulumi.String("Auto")
		}
		clusterArgs.NodeProvisioningProfile = provisioningArgs
	}

	if override := spec.UpgradeOverride; override != nil {
		overrideArgs := &containerservice.KubernetesClusterUpgradeOverrideArgs{
			ForceUpgradeEnabled: pulumi.Bool(override.ForceUpgradeEnabled),
		}
		if override.EffectiveUntil != "" {
			overrideArgs.EffectiveUntil = pulumi.String(override.EffectiveUntil)
		}
		clusterArgs.UpgradeOverride = overrideArgs
	}
}

// itoa converts a proto int32 to its decimal string form (a few
// autoscaler-profile fields are string-typed on the provider wire).
func itoa(v int32) string {
	return strconv.Itoa(int(v))
}
