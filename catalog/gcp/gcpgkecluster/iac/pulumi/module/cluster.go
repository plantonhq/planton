package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpgkeclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkecluster/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// cluster provisions a GKE cluster — the managed Kubernetes control plane
// plus cluster-wide configuration. Node pools are separate GcpGkeNodePool
// resources: the default node pool is always removed at create time on
// Standard clusters, so every pool is an explicitly managed, first-class
// node. Autopilot clusters manage nodes themselves and take no node pools.
//
// Lifecycle notes the API enforces:
//   - name, location, description, network, subnetwork, the whole
//     ip_allocation_policy, datapath_provider, default_max_pods_per_node,
//     confidential_nodes, enable_autopilot, and the private control-plane
//     placement fields are immutable — changing any of them replaces the
//     cluster (and everything running on it).
//   - enable_l4_ilb_subsetting is one-way: it can be turned on in place but
//     never off.
//   - deletion_protection is an engine-side guard: while true, a destroy
//     preview fails before touching the cluster.
func cluster(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpGkeCluster.Spec

	// Enable the Kubernetes Engine API first so a fresh project works on the
	// first deploy. disable_on_destroy stays false: tearing down one cluster
	// must never disable the API for everything else in the project.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("container.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"gke-container.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable container.googleapis.com api")
	}

	// The Workload Identity pool name is fixed by the API to
	// PROJECT_ID.svc.id.goog; when the spec omits project_id the ambient
	// provider project fills it in (the ambient-project contract).
	workloadPoolProject := spec.ProjectId.GetValue()
	if workloadPoolProject == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to resolve ambient project from provider client config")
		}
		workloadPoolProject = clientConfig.Project
	}
	workloadPool := fmt.Sprintf("%s.svc.id.goog", workloadPoolProject)

	args := &container.ClusterArgs{
		Name:     pulumi.String(locals.ClusterName),
		Location: pulumi.StringPtr(spec.Location),

		Network:    pulumi.StringPtr(spec.Network.GetValue()),
		Subnetwork: pulumi.StringPtr(spec.Subnetwork.GetValue()),

		DeletionProtection: pulumi.BoolPtr(spec.GetDeletionProtection()),

		ResourceLabels: pulumi.ToStringMap(locals.GcpLabels),

		ReleaseChannel: &container.ClusterReleaseChannelArgs{
			Channel: pulumi.String(locals.ReleaseChannel),
		},
	}

	// An empty project falls back to the provider's default project — the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// Engine-side destroy stance, layered UNDER deletion_protection: the
	// GKE-native guard blocks the API call; deletion_policy governs what
	// the engine even attempts (PREVENT fails the preview, ABANDON drops
	// state).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// Read-side performance switches for large clusters: skip the per-pool
	// IGM node-count queries and the inline node-pool refresh. The refresh
	// skip is safe in this composition because pools are always separate
	// GcpGkeNodePool resources, never inline blocks here.
	if spec.IgnoreNodeCountChanges {
		args.IgnoreNodeCountChanges = pulumi.BoolPtr(true)
	}
	if spec.SkipNodePoolRefresh {
		args.SkipNodePoolRefresh = pulumi.BoolPtr(true)
	}

	// ALPHA cluster: every alpha feature gate on, no SLA, deleted by GKE
	// after 30 days. Strictly for short-lived evaluation.
	if spec.EnableKubernetesAlpha {
		args.EnableKubernetesAlpha = pulumi.BoolPtr(true)
	}

	if len(spec.K8SBetaApis) > 0 {
		args.EnableK8sBetaApis = &container.ClusterEnableK8sBetaApisArgs{
			EnabledApis: pulumi.ToStringArray(spec.K8SBetaApis),
		}
	}

	if spec.DataplaneOptimizationMode != "" {
		args.DataplaneOptimizationMode = pulumi.StringPtr(spec.DataplaneOptimizationMode)
	}

	if len(spec.AutopilotPrivilegedAdmissionPaths) > 0 {
		args.AutopilotPrivilegedAdmissions = pulumi.ToStringArray(spec.AutopilotPrivilegedAdmissionPaths)
	}

	// Nodes may span fewer (regional) or more (zonal) zones than the control
	// plane; empty defers to GKE's location defaults.
	if len(spec.NodeLocations) > 0 {
		args.NodeLocations = pulumi.ToStringArray(spec.NodeLocations)
	}

	// Clusters are always VPC-native (alias IP). The ip_allocation_policy
	// block is emitted even when the spec omits ip_allocation: an empty
	// block tells GKE to create and manage the pod/service secondary ranges
	// itself, while named ranges pin allocation to planned subnetwork
	// ranges.
	args.NetworkingMode = pulumi.StringPtr("VPC_NATIVE")
	ipAllocationArgs := &container.ClusterIpAllocationPolicyArgs{}
	if spec.IpAllocation != nil {
		if spec.IpAllocation.ClusterSecondaryRangeName.GetValue() != "" {
			ipAllocationArgs.ClusterSecondaryRangeName = pulumi.StringPtr(spec.IpAllocation.ClusterSecondaryRangeName.GetValue())
		}
		if spec.IpAllocation.ServicesSecondaryRangeName.GetValue() != "" {
			ipAllocationArgs.ServicesSecondaryRangeName = pulumi.StringPtr(spec.IpAllocation.ServicesSecondaryRangeName.GetValue())
		}
		if spec.IpAllocation.ClusterIpv4CidrBlock != "" {
			ipAllocationArgs.ClusterIpv4CidrBlock = pulumi.StringPtr(spec.IpAllocation.ClusterIpv4CidrBlock)
		}
		if spec.IpAllocation.ServicesIpv4CidrBlock != "" {
			ipAllocationArgs.ServicesIpv4CidrBlock = pulumi.StringPtr(spec.IpAllocation.ServicesIpv4CidrBlock)
		}
		ipAllocationArgs.StackType = pulumi.StringPtr(spec.IpAllocation.GetStackType())
		if len(spec.IpAllocation.AdditionalPodRangeNames) > 0 {
			ipAllocationArgs.AdditionalPodRangesConfig = &container.ClusterIpAllocationPolicyAdditionalPodRangesConfigArgs{
				PodRangeNames: pulumi.ToStringArray(spec.IpAllocation.AdditionalPodRangeNames),
			}
		}
		if spec.IpAllocation.PodCidrOverprovisionDisabled {
			ipAllocationArgs.PodCidrOverprovisionConfig = &container.ClusterIpAllocationPolicyPodCidrOverprovisionConfigArgs{
				Disabled: pulumi.Bool(true),
			}
		}
		if len(spec.IpAllocation.AdditionalIpRanges) > 0 {
			ranges := container.ClusterIpAllocationPolicyAdditionalIpRangesConfigArray{}
			for _, entry := range spec.IpAllocation.AdditionalIpRanges {
				rangeArgs := &container.ClusterIpAllocationPolicyAdditionalIpRangesConfigArgs{
					Subnetwork: pulumi.String(entry.Subnetwork.GetValue()),
				}
				if len(entry.PodIpv4RangeNames) > 0 {
					rangeArgs.PodIpv4RangeNames = pulumi.ToStringArray(entry.PodIpv4RangeNames)
				}
				if entry.Status != "" {
					rangeArgs.Status = pulumi.StringPtr(entry.Status)
				}
				ranges = append(ranges, rangeArgs)
			}
			ipAllocationArgs.AdditionalIpRangesConfigs = ranges
		}
		if spec.IpAllocation.AutoIpamEnabled {
			ipAllocationArgs.AutoIpamConfig = &container.ClusterIpAllocationPolicyAutoIpamConfigArgs{
				Enabled: pulumi.Bool(true),
			}
		}
		if spec.IpAllocation.NetworkTier != "" {
			ipAllocationArgs.NetworkTierConfig = &container.ClusterIpAllocationPolicyNetworkTierConfigArgs{
				NetworkTier: pulumi.String(spec.IpAllocation.NetworkTier),
			}
		}
	}
	args.IpAllocationPolicy = ipAllocationArgs

	// Standard clusters: drop the API-mandated default node pool
	// immediately — node pools are composed as GcpGkeNodePool resources.
	// Autopilot rejects both fields (GKE owns node management there).
	if spec.EnableAutopilot {
		args.EnableAutopilot = pulumi.BoolPtr(true)
		if spec.AllowNetAdmin {
			args.AllowNetAdmin = pulumi.BoolPtr(true)
		}
	} else {
		args.RemoveDefaultNodePool = pulumi.BoolPtr(true)
		args.InitialNodeCount = pulumi.IntPtr(1)
	}

	if spec.DatapathProvider != "" {
		args.DatapathProvider = pulumi.StringPtr(spec.DatapathProvider)
	}
	if spec.DefaultMaxPodsPerNode != nil {
		args.DefaultMaxPodsPerNode = pulumi.IntPtr(int(spec.GetDefaultMaxPodsPerNode()))
	}
	if spec.EnableIntranodeVisibility {
		args.EnableIntranodeVisibility = pulumi.BoolPtr(true)
	}
	if spec.EnableL4IlbSubsetting {
		args.EnableL4IlbSubsetting = pulumi.BoolPtr(true)
	}
	if spec.EnableFqdnNetworkPolicy {
		args.EnableFqdnNetworkPolicy = pulumi.BoolPtr(true)
	}
	if spec.EnableCiliumClusterwideNetworkPolicy {
		args.EnableCiliumClusterwideNetworkPolicy = pulumi.BoolPtr(true)
	}
	if spec.EnableMultiNetworking {
		args.EnableMultiNetworking = pulumi.BoolPtr(true)
	}
	if spec.PrivateIpv6GoogleAccess != "" {
		args.PrivateIpv6GoogleAccess = pulumi.StringPtr(spec.PrivateIpv6GoogleAccess)
	}
	if spec.InTransitEncryption != "" {
		args.InTransitEncryptionConfig = pulumi.StringPtr(spec.InTransitEncryption)
	}
	if spec.DisableL4LbFirewallReconciliation {
		args.DisableL4LbFirewallReconciliation = pulumi.BoolPtr(true)
	}

	// Calico NetworkPolicy enforcement is two coupled settings: the
	// enforcement block here and the addon below. Both follow the single
	// spec toggle so they can never drift apart. Omitted entirely on
	// Autopilot (which enforces NetworkPolicy natively via Dataplane V2).
	if spec.EnableNetworkPolicy {
		args.NetworkPolicy = &container.ClusterNetworkPolicyArgs{
			Enabled:  pulumi.Bool(true),
			Provider: pulumi.StringPtr("CALICO"),
		}
	}

	if spec.DisableDefaultSnat {
		args.DefaultSnatStatus = &container.ClusterDefaultSnatStatusArgs{
			Disabled: pulumi.Bool(true),
		}
	}

	if spec.TotalEgressBandwidthTier != "" {
		args.NetworkPerformanceConfig = &container.ClusterNetworkPerformanceConfigArgs{
			TotalEgressBandwidthTier: pulumi.String(spec.TotalEgressBandwidthTier),
		}
	}

	if spec.DnsConfig != nil {
		dnsArgs := &container.ClusterDnsConfigArgs{}
		if spec.DnsConfig.ClusterDns != "" {
			dnsArgs.ClusterDns = pulumi.StringPtr(spec.DnsConfig.ClusterDns)
		}
		if spec.DnsConfig.ClusterDnsScope != "" {
			dnsArgs.ClusterDnsScope = pulumi.StringPtr(spec.DnsConfig.ClusterDnsScope)
		}
		if spec.DnsConfig.ClusterDnsDomain != "" {
			dnsArgs.ClusterDnsDomain = pulumi.StringPtr(spec.DnsConfig.ClusterDnsDomain)
		}
		if spec.DnsConfig.AdditiveVpcScopeDnsDomain != "" {
			dnsArgs.AdditiveVpcScopeDnsDomain = pulumi.StringPtr(spec.DnsConfig.AdditiveVpcScopeDnsDomain)
		}
		args.DnsConfig = dnsArgs
	}

	if spec.GatewayApiChannel != "" {
		args.GatewayApiConfig = &container.ClusterGatewayApiConfigArgs{
			Channel: pulumi.String(spec.GatewayApiChannel),
		}
	}

	if spec.EnableServiceExternalIps {
		args.ServiceExternalIpsConfig = &container.ClusterServiceExternalIpsConfigArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	// Private topology: private nodes need Cloud NAT (a GcpRouterNat on the
	// network) for outbound internet; a private-only endpoint removes
	// public kubectl access entirely.
	if spec.PrivateCluster != nil {
		privateArgs := &container.ClusterPrivateClusterConfigArgs{
			EnablePrivateNodes:    pulumi.BoolPtr(spec.PrivateCluster.EnablePrivateNodes),
			EnablePrivateEndpoint: pulumi.BoolPtr(spec.PrivateCluster.EnablePrivateEndpoint),
		}
		if spec.PrivateCluster.MasterIpv4CidrBlock != "" {
			privateArgs.MasterIpv4CidrBlock = pulumi.StringPtr(spec.PrivateCluster.MasterIpv4CidrBlock)
		}
		if spec.PrivateCluster.PrivateEndpointSubnetwork.GetValue() != "" {
			privateArgs.PrivateEndpointSubnetwork = pulumi.StringPtr(spec.PrivateCluster.PrivateEndpointSubnetwork.GetValue())
		}
		if spec.PrivateCluster.EnableMasterGlobalAccess {
			privateArgs.MasterGlobalAccessConfig = &container.ClusterPrivateClusterConfigMasterGlobalAccessConfigArgs{
				Enabled: pulumi.Bool(true),
			}
		}
		args.PrivateClusterConfig = privateArgs
	}

	if spec.MasterAuthorizedNetworks != nil {
		manArgs := &container.ClusterMasterAuthorizedNetworksConfigArgs{}
		if spec.MasterAuthorizedNetworks.GcpPublicCidrsAccessEnabled != nil {
			manArgs.GcpPublicCidrsAccessEnabled = pulumi.BoolPtr(spec.MasterAuthorizedNetworks.GetGcpPublicCidrsAccessEnabled())
		}
		if spec.MasterAuthorizedNetworks.PrivateEndpointEnforcementEnabled != nil {
			manArgs.PrivateEndpointEnforcementEnabled = pulumi.BoolPtr(spec.MasterAuthorizedNetworks.GetPrivateEndpointEnforcementEnabled())
		}
		cidrBlocks := container.ClusterMasterAuthorizedNetworksConfigCidrBlockArray{}
		for _, cidr := range spec.MasterAuthorizedNetworks.CidrBlocks {
			cidrArgs := &container.ClusterMasterAuthorizedNetworksConfigCidrBlockArgs{
				CidrBlock: pulumi.String(cidr.CidrBlock),
			}
			if cidr.DisplayName != "" {
				cidrArgs.DisplayName = pulumi.StringPtr(cidr.DisplayName)
			}
			cidrBlocks = append(cidrBlocks, cidrArgs)
		}
		manArgs.CidrBlocks = cidrBlocks
		args.MasterAuthorizedNetworksConfig = manArgs
	}

	if spec.ControlPlaneEndpoints != nil {
		dnsEndpointArgs := &container.ClusterControlPlaneEndpointsConfigDnsEndpointConfigArgs{
			AllowExternalTraffic: pulumi.BoolPtr(spec.ControlPlaneEndpoints.DnsEndpointAllowExternalTraffic),
		}
		if spec.ControlPlaneEndpoints.EnableK8STokensViaDns != nil {
			dnsEndpointArgs.EnableK8sTokensViaDns = pulumi.BoolPtr(*spec.ControlPlaneEndpoints.EnableK8STokensViaDns)
		}
		if spec.ControlPlaneEndpoints.EnableK8SCertsViaDns != nil {
			dnsEndpointArgs.EnableK8sCertsViaDns = pulumi.BoolPtr(*spec.ControlPlaneEndpoints.EnableK8SCertsViaDns)
		}
		args.ControlPlaneEndpointsConfig = &container.ClusterControlPlaneEndpointsConfigArgs{
			DnsEndpointConfig: dnsEndpointArgs,
			IpEndpointsConfig: &container.ClusterControlPlaneEndpointsConfigIpEndpointsConfigArgs{
				Enabled: pulumi.BoolPtr(spec.ControlPlaneEndpoints.GetIpEndpointsEnabled()),
			},
		}
	}

	// Legacy client-certificate issuance: certificate auth bypasses IAM
	// and cannot be revoked short of rotating the cluster CA — forwarded
	// only when a manifest explicitly takes a stance.
	if spec.IssueClientCertificate != nil {
		args.MasterAuth = &container.ClusterMasterAuthArgs{
			ClientCertificateConfig: &container.ClusterMasterAuthClientCertificateConfigArgs{
				IssueClientCertificate: pulumi.Bool(*spec.IssueClientCertificate),
			},
		}
	}

	if spec.NodeCreationMode != "" {
		args.NodeCreationConfig = &container.ClusterNodeCreationConfigArgs{
			NodeCreationMode: pulumi.String(spec.NodeCreationMode),
		}
	}

	if spec.GkeAutoUpgradePatchMode != "" {
		args.GkeAutoUpgradeConfig = &container.ClusterGkeAutoUpgradeConfigArgs{
			PatchMode: pulumi.String(spec.GkeAutoUpgradePatchMode),
		}
	}

	if rbac := spec.RbacBindingConfig; rbac != nil {
		rbacArgs := &container.ClusterRbacBindingConfigArgs{}
		if rbac.EnableInsecureBindingSystemAuthenticated != nil {
			rbacArgs.EnableInsecureBindingSystemAuthenticated = pulumi.BoolPtr(*rbac.EnableInsecureBindingSystemAuthenticated)
		}
		if rbac.EnableInsecureBindingSystemUnauthenticated != nil {
			rbacArgs.EnableInsecureBindingSystemUnauthenticated = pulumi.BoolPtr(*rbac.EnableInsecureBindingSystemUnauthenticated)
		}
		args.RbacBindingConfig = rbacArgs
	}

	if policy := spec.AutopilotPolicy; policy != nil {
		policyArgs := &container.ClusterAutopilotClusterPolicyConfigArgs{}
		if policy.NoStandardNodePools != nil {
			policyArgs.NoStandardNodePools = pulumi.BoolPtr(*policy.NoStandardNodePools)
		}
		if policy.NoSystemImpersonation != nil {
			policyArgs.NoSystemImpersonation = pulumi.BoolPtr(*policy.NoSystemImpersonation)
		}
		if policy.NoSystemMutation != nil {
			policyArgs.NoSystemMutation = pulumi.BoolPtr(*policy.NoSystemMutation)
		}
		if policy.NoUnsafeWebhooks != nil {
			policyArgs.NoUnsafeWebhooks = pulumi.BoolPtr(*policy.NoUnsafeWebhooks)
		}
		args.AutopilotClusterPolicyConfig = policyArgs
	}

	// Node settings GKE applies to the pools IT manages on Autopilot — the
	// Autopilot counterpart of per-pool node_config.
	if npac := spec.NodePoolAutoConfig; npac != nil {
		npacArgs := &container.ClusterNodePoolAutoConfigArgs{}
		if len(npac.NetworkTags) > 0 {
			npacArgs.NetworkTags = &container.ClusterNodePoolAutoConfigNetworkTagsArgs{
				Tags: pulumi.ToStringArray(npac.NetworkTags),
			}
		}
		if len(npac.ResourceManagerTags) > 0 {
			npacArgs.ResourceManagerTags = pulumi.ToStringMap(npac.ResourceManagerTags)
		}
		if npac.CgroupMode != "" || npac.NodeKernelModuleLoadingPolicy != "" {
			linuxArgs := &container.ClusterNodePoolAutoConfigLinuxNodeConfigArgs{}
			if npac.CgroupMode != "" {
				linuxArgs.CgroupMode = pulumi.StringPtr(npac.CgroupMode)
			}
			if npac.NodeKernelModuleLoadingPolicy != "" {
				linuxArgs.NodeKernelModuleLoading = &container.ClusterNodePoolAutoConfigLinuxNodeConfigNodeKernelModuleLoadingArgs{
					Policy: pulumi.StringPtr(npac.NodeKernelModuleLoadingPolicy),
				}
			}
			npacArgs.LinuxNodeConfig = linuxArgs
		}
		if npac.InsecureKubeletReadonlyPortEnabled != "" {
			npacArgs.NodeKubeletConfig = &container.ClusterNodePoolAutoConfigNodeKubeletConfigArgs{
				InsecureKubeletReadonlyPortEnabled: pulumi.StringPtr(npac.InsecureKubeletReadonlyPortEnabled),
			}
		}
		args.NodePoolAutoConfig = npacArgs
	}

	// Creation-time defaults inherited by every node pool on a Standard
	// cluster; a pool's own node_config overrides these.
	if npd := spec.NodePoolDefaults; npd != nil {
		defaultsArgs := &container.ClusterNodePoolDefaultsNodeConfigDefaultsArgs{}
		if npd.GcfsEnabled != nil {
			defaultsArgs.GcfsConfig = &container.ClusterNodePoolDefaultsNodeConfigDefaultsGcfsConfigArgs{
				Enabled: pulumi.Bool(*npd.GcfsEnabled),
			}
		}
		if npd.InsecureKubeletReadonlyPortEnabled != "" {
			defaultsArgs.InsecureKubeletReadonlyPortEnabled = pulumi.StringPtr(npd.InsecureKubeletReadonlyPortEnabled)
		}
		if npd.LoggingVariant != "" {
			defaultsArgs.LoggingVariant = pulumi.StringPtr(npd.LoggingVariant)
		}
		if containerd := npd.ContainerdConfig; containerd != nil {
			defaultsArgs.ContainerdConfig = buildClusterContainerdDefaults(containerd)
		}
		args.NodePoolDefaults = &container.ClusterNodePoolDefaultsArgs{
			NodeConfigDefaults: defaultsArgs,
		}
	}

	// Bring-your-own control-plane trust: customer CAs and KMS keys for
	// the control plane's disks and ServiceAccount JWT signing.
	if keys := spec.UserManagedKeys; keys != nil {
		keysArgs := &container.ClusterUserManagedKeysConfigArgs{}
		if keys.ClusterCa != "" {
			keysArgs.ClusterCa = pulumi.StringPtr(keys.ClusterCa)
		}
		if keys.EtcdApiCa != "" {
			keysArgs.EtcdApiCa = pulumi.StringPtr(keys.EtcdApiCa)
		}
		if keys.EtcdPeerCa != "" {
			keysArgs.EtcdPeerCa = pulumi.StringPtr(keys.EtcdPeerCa)
		}
		if keys.AggregationCa != "" {
			keysArgs.AggregationCa = pulumi.StringPtr(keys.AggregationCa)
		}
		if keys.ControlPlaneDiskEncryptionKey.GetValue() != "" {
			keysArgs.ControlPlaneDiskEncryptionKey = pulumi.StringPtr(keys.ControlPlaneDiskEncryptionKey.GetValue())
		}
		if keys.GkeopsEtcdBackupEncryptionKey.GetValue() != "" {
			keysArgs.GkeopsEtcdBackupEncryptionKey = pulumi.StringPtr(keys.GkeopsEtcdBackupEncryptionKey.GetValue())
		}
		if len(keys.ServiceAccountSigningKeys) > 0 {
			keysArgs.ServiceAccountSigningKeys = pulumi.ToStringArray(keys.ServiceAccountSigningKeys)
		}
		if len(keys.ServiceAccountVerificationKeys) > 0 {
			keysArgs.ServiceAccountVerificationKeys = pulumi.ToStringArray(keys.ServiceAccountVerificationKeys)
		}
		args.UserManagedKeysConfig = keysArgs
	}

	if spec.MinMasterVersion != "" {
		args.MinMasterVersion = pulumi.StringPtr(spec.MinMasterVersion)
	}

	if spec.MaintenancePolicy != nil {
		maintenanceArgs := &container.ClusterMaintenancePolicyArgs{}
		if spec.MaintenancePolicy.DailyWindow != nil {
			maintenanceArgs.DailyMaintenanceWindow = &container.ClusterMaintenancePolicyDailyMaintenanceWindowArgs{
				StartTime: pulumi.String(spec.MaintenancePolicy.DailyWindow.StartTime),
			}
		}
		if spec.MaintenancePolicy.RecurringWindow != nil {
			maintenanceArgs.RecurringWindow = &container.ClusterMaintenancePolicyRecurringWindowArgs{
				StartTime:  pulumi.String(spec.MaintenancePolicy.RecurringWindow.StartTime),
				EndTime:    pulumi.String(spec.MaintenancePolicy.RecurringWindow.EndTime),
				Recurrence: pulumi.String(spec.MaintenancePolicy.RecurringWindow.Recurrence),
			}
		}
		if len(spec.MaintenancePolicy.Exclusions) > 0 {
			exclusions := container.ClusterMaintenancePolicyMaintenanceExclusionArray{}
			for _, exclusion := range spec.MaintenancePolicy.Exclusions {
				exclusionArgs := &container.ClusterMaintenancePolicyMaintenanceExclusionArgs{
					ExclusionName: pulumi.String(exclusion.ExclusionName),
					StartTime:     pulumi.String(exclusion.StartTime),
					EndTime:       pulumi.String(exclusion.EndTime),
				}
				if exclusion.Scope != "" {
					optionsArgs := &container.ClusterMaintenancePolicyMaintenanceExclusionExclusionOptionsArgs{
						Scope: pulumi.String(exclusion.Scope),
					}
					if exclusion.EndTimeBehavior != "" {
						optionsArgs.EndTimeBehavior = pulumi.StringPtr(exclusion.EndTimeBehavior)
					}
					exclusionArgs.ExclusionOptions = optionsArgs
				}
				exclusions = append(exclusions, exclusionArgs)
			}
			maintenanceArgs.MaintenanceExclusions = exclusions
		}
		if budget := spec.MaintenancePolicy.DisruptionBudget; budget != nil {
			budgetArgs := &container.ClusterMaintenancePolicyDisruptionBudgetArgs{}
			if budget.MinorVersionDisruptionInterval != "" {
				budgetArgs.MinorVersionDisruptionInterval = pulumi.StringPtr(budget.MinorVersionDisruptionInterval)
			}
			if budget.PatchVersionDisruptionInterval != "" {
				budgetArgs.PatchVersionDisruptionInterval = pulumi.StringPtr(budget.PatchVersionDisruptionInterval)
			}
			maintenanceArgs.DisruptionBudget = budgetArgs
		}
		args.MaintenancePolicy = maintenanceArgs
	}

	// Node auto-provisioning: GKE creates/deletes node pools within the
	// resource limits — bounded by spec-level validation so an enabled NAP
	// always carries limits (an unbounded NAP is an unbounded bill).
	if spec.ClusterAutoscaling != nil {
		autoscalingArgs := &container.ClusterClusterAutoscalingArgs{
			Enabled:            pulumi.BoolPtr(spec.ClusterAutoscaling.Enabled),
			AutoscalingProfile: pulumi.StringPtr(spec.ClusterAutoscaling.GetAutoscalingProfile()),
		}
		if spec.ClusterAutoscaling.DefaultComputeClassEnabled != nil {
			autoscalingArgs.DefaultComputeClassEnabled = pulumi.BoolPtr(*spec.ClusterAutoscaling.DefaultComputeClassEnabled)
		}
		if len(spec.ClusterAutoscaling.AutoProvisioningLocations) > 0 {
			autoscalingArgs.AutoProvisioningLocations = pulumi.ToStringArray(spec.ClusterAutoscaling.AutoProvisioningLocations)
		}
		if len(spec.ClusterAutoscaling.ResourceLimits) > 0 {
			limits := container.ClusterClusterAutoscalingResourceLimitArray{}
			for _, limit := range spec.ClusterAutoscaling.ResourceLimits {
				limits = append(limits, &container.ClusterClusterAutoscalingResourceLimitArgs{
					ResourceType: pulumi.String(limit.ResourceType),
					Minimum:      pulumi.IntPtr(int(limit.Minimum)),
					Maximum:      pulumi.Int(int(limit.Maximum)),
				})
			}
			autoscalingArgs.ResourceLimits = limits
		}
		if defaults := spec.ClusterAutoscaling.AutoProvisioningDefaults; defaults != nil {
			defaultsArgs := &container.ClusterClusterAutoscalingAutoProvisioningDefaultsArgs{
				ShieldedInstanceConfig: &container.ClusterClusterAutoscalingAutoProvisioningDefaultsShieldedInstanceConfigArgs{
					EnableSecureBoot:          pulumi.BoolPtr(defaults.EnableSecureBoot),
					EnableIntegrityMonitoring: pulumi.BoolPtr(defaults.GetEnableIntegrityMonitoring()),
				},
				Management: &container.ClusterClusterAutoscalingAutoProvisioningDefaultsManagementArgs{
					AutoUpgrade: pulumi.BoolPtr(defaults.GetAutoUpgrade()),
					AutoRepair:  pulumi.BoolPtr(defaults.GetAutoRepair()),
				},
			}
			if defaults.ServiceAccount.GetValue() != "" {
				defaultsArgs.ServiceAccount = pulumi.StringPtr(defaults.ServiceAccount.GetValue())
			}
			if len(defaults.OauthScopes) > 0 {
				defaultsArgs.OauthScopes = pulumi.ToStringArray(defaults.OauthScopes)
			}
			if defaults.DiskSizeGb != nil {
				defaultsArgs.DiskSize = pulumi.IntPtr(int(defaults.GetDiskSizeGb()))
			}
			if defaults.DiskType != "" {
				defaultsArgs.DiskType = pulumi.StringPtr(defaults.DiskType)
			}
			if defaults.ImageType != "" {
				defaultsArgs.ImageType = pulumi.StringPtr(defaults.ImageType)
			}
			if defaults.MinCpuPlatform != "" {
				defaultsArgs.MinCpuPlatform = pulumi.StringPtr(defaults.MinCpuPlatform)
			}
			if defaults.BootDiskKmsKey.GetValue() != "" {
				defaultsArgs.BootDiskKmsKey = pulumi.StringPtr(defaults.BootDiskKmsKey.GetValue())
			}
			if upgrade := defaults.UpgradeSettings; upgrade != nil {
				upgradeArgs := &container.ClusterClusterAutoscalingAutoProvisioningDefaultsUpgradeSettingsArgs{}
				if upgrade.MaxSurge != nil {
					upgradeArgs.MaxSurge = pulumi.IntPtr(int(*upgrade.MaxSurge))
				}
				if upgrade.MaxUnavailable != nil {
					upgradeArgs.MaxUnavailable = pulumi.IntPtr(int(*upgrade.MaxUnavailable))
				}
				if upgrade.Strategy != "" {
					upgradeArgs.Strategy = pulumi.StringPtr(upgrade.Strategy)
				}
				if blueGreen := upgrade.BlueGreenSettings; blueGreen != nil {
					blueGreenArgs := &container.ClusterClusterAutoscalingAutoProvisioningDefaultsUpgradeSettingsBlueGreenSettingsArgs{}
					if blueGreen.NodePoolSoakDuration != "" {
						blueGreenArgs.NodePoolSoakDuration = pulumi.StringPtr(blueGreen.NodePoolSoakDuration)
					}
					if rollout := blueGreen.StandardRolloutPolicy; rollout != nil {
						rolloutArgs := &container.ClusterClusterAutoscalingAutoProvisioningDefaultsUpgradeSettingsBlueGreenSettingsStandardRolloutPolicyArgs{}
						if rollout.BatchPercentage != nil {
							rolloutArgs.BatchPercentage = pulumi.Float64Ptr(float64(*rollout.BatchPercentage))
						}
						if rollout.BatchNodeCount != nil {
							rolloutArgs.BatchNodeCount = pulumi.IntPtr(int(*rollout.BatchNodeCount))
						}
						if rollout.BatchSoakDuration != "" {
							rolloutArgs.BatchSoakDuration = pulumi.StringPtr(rollout.BatchSoakDuration)
						}
						blueGreenArgs.StandardRolloutPolicy = rolloutArgs
					}
					upgradeArgs.BlueGreenSettings = blueGreenArgs
				}
				defaultsArgs.UpgradeSettings = upgradeArgs
			}
			autoscalingArgs.AutoProvisioningDefaults = defaultsArgs
		}
		args.ClusterAutoscaling = autoscalingArgs
	}

	if spec.EnableVerticalPodAutoscaling {
		args.VerticalPodAutoscaling = &container.ClusterVerticalPodAutoscalingArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	if spec.HpaProfile != "" {
		args.PodAutoscaling = &container.ClusterPodAutoscalingArgs{
			HpaProfile: pulumi.String(spec.HpaProfile),
		}
	}

	// Workload Identity Federation for GKE: the pool name is fixed by the
	// API to PROJECT_ID.svc.id.goog. Autopilot clusters have it always on —
	// the block is suppressed there to avoid a permanent diff.
	workloadIdentityEnabled := spec.GetWorkloadIdentityEnabled()
	if workloadIdentityEnabled && !spec.EnableAutopilot {
		args.WorkloadIdentityConfig = &container.ClusterWorkloadIdentityConfigArgs{
			WorkloadPool: pulumi.StringPtr(workloadPool),
		}
	}

	// Shielded GKE nodes: only forwarded when the spec sets it (GCP default
	// is true; Autopilot rejects the field and spec validation blocks it
	// there).
	if spec.EnableShieldedNodes != nil {
		args.EnableShieldedNodes = pulumi.BoolPtr(spec.GetEnableShieldedNodes())
	}

	if spec.DatabaseEncryption != nil {
		encryptionArgs := &container.ClusterDatabaseEncryptionArgs{
			State: pulumi.String(spec.DatabaseEncryption.State),
		}
		if spec.DatabaseEncryption.KeyName.GetValue() != "" {
			encryptionArgs.KeyName = pulumi.StringPtr(spec.DatabaseEncryption.KeyName.GetValue())
		}
		args.DatabaseEncryption = encryptionArgs
	}

	if spec.BinaryAuthorizationEvaluationMode != "" {
		args.BinaryAuthorization = &container.ClusterBinaryAuthorizationArgs{
			EvaluationMode: pulumi.StringPtr(spec.BinaryAuthorizationEvaluationMode),
		}
	}

	if spec.SecurityPosture != nil {
		postureArgs := &container.ClusterSecurityPostureConfigArgs{}
		if spec.SecurityPosture.Mode != "" {
			postureArgs.Mode = pulumi.StringPtr(spec.SecurityPosture.Mode)
		}
		if spec.SecurityPosture.VulnerabilityMode != "" {
			postureArgs.VulnerabilityMode = pulumi.StringPtr(spec.SecurityPosture.VulnerabilityMode)
		}
		args.SecurityPostureConfig = postureArgs
	}

	if spec.AuthenticatorSecurityGroup != "" {
		args.AuthenticatorGroupsConfig = &container.ClusterAuthenticatorGroupsConfigArgs{
			SecurityGroup: pulumi.String(spec.AuthenticatorSecurityGroup),
		}
	}

	if spec.EnableLegacyAbac {
		args.EnableLegacyAbac = pulumi.BoolPtr(true)
	}

	if spec.EnableMeshCertificates {
		args.MeshCertificates = &container.ClusterMeshCertificatesArgs{
			EnableCertificates: pulumi.Bool(true),
		}
	}

	if spec.EnableSecretManagerCsi {
		secretManagerArgs := &container.ClusterSecretManagerConfigArgs{
			Enabled: pulumi.Bool(true),
		}
		if rotation := spec.SecretManagerRotation; rotation != nil {
			rotationArgs := &container.ClusterSecretManagerConfigRotationConfigArgs{
				Enabled: pulumi.Bool(rotation.Enabled),
			}
			if rotation.RotationInterval != "" {
				rotationArgs.RotationInterval = pulumi.StringPtr(rotation.RotationInterval)
			}
			secretManagerArgs.RotationConfig = rotationArgs
		}
		args.SecretManagerConfig = secretManagerArgs
	}

	// The Secret Manager SYNC add-on (secrets into Kubernetes Secret
	// objects) — a separate add-on from the CSI mount path above.
	if sync := spec.SecretSync; sync != nil {
		syncArgs := &container.ClusterSecretSyncConfigArgs{
			Enabled: pulumi.Bool(sync.Enabled),
		}
		if sync.RotationEnabled || sync.RotationInterval != "" {
			rotationArgs := &container.ClusterSecretSyncConfigRotationConfigArgs{
				Enabled: pulumi.Bool(sync.RotationEnabled),
			}
			if sync.RotationInterval != "" {
				rotationArgs.RotationInterval = pulumi.StringPtr(sync.RotationInterval)
			}
			syncArgs.RotationConfig = rotationArgs
		}
		args.SecretSyncConfig = syncArgs
	}

	if spec.ConfidentialNodes != nil {
		confidentialArgs := &container.ClusterConfidentialNodesArgs{
			Enabled: pulumi.Bool(spec.ConfidentialNodes.Enabled),
		}
		if spec.ConfidentialNodes.ConfidentialInstanceType != "" {
			confidentialArgs.ConfidentialInstanceType = pulumi.StringPtr(spec.ConfidentialNodes.ConfidentialInstanceType)
		}
		args.ConfidentialNodes = confidentialArgs
	}

	if spec.AnonymousAuthenticationMode != "" {
		args.AnonymousAuthenticationConfig = &container.ClusterAnonymousAuthenticationConfigArgs{
			Mode: pulumi.String(spec.AnonymousAuthenticationMode),
		}
	}

	if spec.EnableIdentityService {
		args.IdentityServiceConfig = &container.ClusterIdentityServiceConfigArgs{
			Enabled: pulumi.BoolPtr(true),
		}
	}

	// Logging switched off is written as the empty component list (GKE's
	// spelling of "no Cloud Logging integration"); the API already refuses a
	// switched-on block with no components. Monitoring's presence carries
	// nothing by itself (an empty component list is GKE's default there).
	if spec.Logging != nil {
		components := []string{}
		if spec.Logging.GetEnabled() {
			components = spec.Logging.Components
		}
		args.LoggingConfig = &container.ClusterLoggingConfigArgs{
			EnableComponents: pulumi.ToStringArray(components),
		}
	}

	if spec.Monitoring != nil {
		monitoringArgs := &container.ClusterMonitoringConfigArgs{}
		if len(spec.Monitoring.Components) > 0 {
			monitoringArgs.EnableComponents = pulumi.ToStringArray(spec.Monitoring.Components)
		}
		prometheusArgs := &container.ClusterMonitoringConfigManagedPrometheusArgs{
			Enabled: pulumi.Bool(spec.Monitoring.GetManagedPrometheusEnabled()),
		}
		if spec.Monitoring.AutoMonitoringScope != "" {
			prometheusArgs.AutoMonitoringConfig = &container.ClusterMonitoringConfigManagedPrometheusAutoMonitoringConfigArgs{
				Scope: pulumi.String(spec.Monitoring.AutoMonitoringScope),
			}
		}
		monitoringArgs.ManagedPrometheus = prometheusArgs
		if spec.Monitoring.AdvancedDatapathMetricsEnabled || spec.Monitoring.AdvancedDatapathRelayEnabled {
			monitoringArgs.AdvancedDatapathObservabilityConfig = &container.ClusterMonitoringConfigAdvancedDatapathObservabilityConfigArgs{
				EnableMetrics: pulumi.Bool(spec.Monitoring.AdvancedDatapathMetricsEnabled),
				EnableRelay:   pulumi.Bool(spec.Monitoring.AdvancedDatapathRelayEnabled),
			}
		}
		args.MonitoringConfig = monitoringArgs
	}

	if spec.NotificationPubsub != nil {
		pubsubArgs := &container.ClusterNotificationConfigPubsubArgs{
			Enabled: pulumi.Bool(spec.NotificationPubsub.Enabled),
		}
		if spec.NotificationPubsub.Topic.GetValue() != "" {
			pubsubArgs.Topic = pulumi.StringPtr(spec.NotificationPubsub.Topic.GetValue())
		}
		if len(spec.NotificationPubsub.EventTypes) > 0 {
			pubsubArgs.Filter = &container.ClusterNotificationConfigPubsubFilterArgs{
				EventTypes: pulumi.ToStringArray(spec.NotificationPubsub.EventTypes),
			}
		}
		args.NotificationConfig = &container.ClusterNotificationConfigArgs{
			Pubsub: pubsubArgs,
		}
	}

	if spec.EnableCostManagement {
		args.CostManagementConfig = &container.ClusterCostManagementConfigArgs{
			Enabled: pulumi.Bool(true),
		}
	}

	if spec.ResourceUsageExport != nil {
		args.ResourceUsageExportConfig = &container.ClusterResourceUsageExportConfigArgs{
			EnableNetworkEgressMetering:       pulumi.BoolPtr(spec.ResourceUsageExport.EnableNetworkEgressMetering),
			EnableResourceConsumptionMetering: pulumi.BoolPtr(spec.ResourceUsageExport.GetEnableResourceConsumptionMetering()),
			BigqueryDestination: &container.ClusterResourceUsageExportConfigBigqueryDestinationArgs{
				DatasetId: pulumi.String(spec.ResourceUsageExport.BigqueryDatasetId.GetValue()),
			},
		}
	}

	// Addons: emitted when the spec configures them or when Calico network
	// policy needs its companion addon. The network-policy addon toggle
	// always mirrors enable_network_policy (never set on Autopilot).
	if spec.Addons != nil || (spec.EnableNetworkPolicy && !spec.EnableAutopilot) {
		addonsArgs := &container.ClusterAddonsConfigArgs{}
		if !spec.EnableAutopilot {
			addonsArgs.NetworkPolicyConfig = &container.ClusterAddonsConfigNetworkPolicyConfigArgs{
				Disabled: pulumi.Bool(!spec.EnableNetworkPolicy),
			}
		}
		httpLoadBalancingEnabled := true
		horizontalPodAutoscalingEnabled := true
		pdCsiEnabled := true
		if spec.Addons != nil {
			httpLoadBalancingEnabled = spec.Addons.GetHttpLoadBalancingEnabled()
			horizontalPodAutoscalingEnabled = spec.Addons.GetHorizontalPodAutoscalingEnabled()
			pdCsiEnabled = spec.Addons.GetGcePersistentDiskCsiDriverEnabled()
		}
		addonsArgs.HttpLoadBalancing = &container.ClusterAddonsConfigHttpLoadBalancingArgs{
			Disabled: pulumi.Bool(!httpLoadBalancingEnabled),
		}
		addonsArgs.HorizontalPodAutoscaling = &container.ClusterAddonsConfigHorizontalPodAutoscalingArgs{
			Disabled: pulumi.Bool(!horizontalPodAutoscalingEnabled),
		}
		addonsArgs.GcePersistentDiskCsiDriverConfig = &container.ClusterAddonsConfigGcePersistentDiskCsiDriverConfigArgs{
			Enabled: pulumi.Bool(pdCsiEnabled),
		}
		if spec.Addons != nil {
			if spec.Addons.GcpFilestoreCsiDriverEnabled {
				addonsArgs.GcpFilestoreCsiDriverConfig = &container.ClusterAddonsConfigGcpFilestoreCsiDriverConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.GcsFuseCsiDriverEnabled {
				addonsArgs.GcsFuseCsiDriverConfig = &container.ClusterAddonsConfigGcsFuseCsiDriverConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.GkeBackupAgentEnabled {
				addonsArgs.GkeBackupAgentConfig = &container.ClusterAddonsConfigGkeBackupAgentConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.DnsCacheEnabled {
				addonsArgs.DnsCacheConfig = &container.ClusterAddonsConfigDnsCacheConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.ConfigConnectorEnabled {
				addonsArgs.ConfigConnectorConfig = &container.ClusterAddonsConfigConfigConnectorConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.StatefulHaEnabled {
				addonsArgs.StatefulHaConfig = &container.ClusterAddonsConfigStatefulHaConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.RayOperatorEnabled {
				rayArgs := &container.ClusterAddonsConfigRayOperatorConfigArgs{
					Enabled: pulumi.Bool(true),
				}
				if spec.Addons.RayClusterLoggingEnabled {
					rayArgs.RayClusterLoggingConfig = &container.ClusterAddonsConfigRayOperatorConfigRayClusterLoggingConfigArgs{
						Enabled: pulumi.Bool(true),
					}
				}
				if spec.Addons.RayClusterMonitoringEnabled {
					rayArgs.RayClusterMonitoringConfig = &container.ClusterAddonsConfigRayOperatorConfigRayClusterMonitoringConfigArgs{
						Enabled: pulumi.Bool(true),
					}
				}
				addonsArgs.RayOperatorConfigs = container.ClusterAddonsConfigRayOperatorConfigArray{rayArgs}
			}
			// The Cloud Run addon's argument is inverted (disabled) — the
			// spec keeps the affirmative form every other addon uses.
			if spec.Addons.CloudrunEnabled {
				cloudrunArgs := &container.ClusterAddonsConfigCloudrunConfigArgs{
					Disabled: pulumi.Bool(false),
				}
				if spec.Addons.CloudrunLoadBalancerType != "" {
					cloudrunArgs.LoadBalancerType = pulumi.StringPtr(spec.Addons.CloudrunLoadBalancerType)
				}
				addonsArgs.CloudrunConfig = cloudrunArgs
			}
			if spec.Addons.ParallelstoreCsiDriverEnabled {
				addonsArgs.ParallelstoreCsiDriverConfig = &container.ClusterAddonsConfigParallelstoreCsiDriverConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.LustreCsiDriverEnabled {
				lustreArgs := &container.ClusterAddonsConfigLustreCsiDriverConfigArgs{
					Enabled: pulumi.Bool(true),
				}
				if spec.Addons.LustreCsiLegacyPortEnabled {
					lustreArgs.EnableLegacyLustrePort = pulumi.BoolPtr(true)
				}
				if spec.Addons.LustreCsiDisableMultiNic {
					lustreArgs.DisableMultiNic = pulumi.BoolPtr(true)
				}
				addonsArgs.LustreCsiDriverConfig = lustreArgs
			}
			if spec.Addons.PodSnapshotEnabled {
				addonsArgs.PodSnapshotConfig = &container.ClusterAddonsConfigPodSnapshotConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.AgentSandboxEnabled {
				addonsArgs.AgentSandboxConfig = &container.ClusterAddonsConfigAgentSandboxConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.SliceControllerEnabled {
				addonsArgs.SliceControllerConfig = &container.ClusterAddonsConfigSliceControllerConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
			if spec.Addons.SlurmOperatorEnabled {
				addonsArgs.SlurmOperatorConfig = &container.ClusterAddonsConfigSlurmOperatorConfigArgs{
					Enabled: pulumi.Bool(true),
				}
			}
		}
		args.AddonsConfig = addonsArgs
	}

	if spec.FleetProject != "" || spec.FleetMembershipType != "" {
		fleetArgs := &container.ClusterFleetArgs{}
		if spec.FleetProject != "" {
			fleetArgs.Project = pulumi.StringPtr(spec.FleetProject)
		}
		if spec.FleetMembershipType != "" {
			fleetArgs.MembershipType = pulumi.StringPtr(spec.FleetMembershipType)
		}
		args.Fleet = fleetArgs
	}

	createdCluster, err := container.NewCluster(ctx, "cluster", args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create container cluster")
	}

	ctx.Export(OpEndpoint, createdCluster.Endpoint)
	// The CA certificate is public trust material (clients install it as a
	// trust anchor), not a secret.
	ctx.Export(OpClusterCaCertificate, createdCluster.MasterAuth.ClusterCaCertificate())
	if workloadIdentityEnabled || spec.EnableAutopilot {
		ctx.Export(OpWorkloadIdentityPool, pulumi.String(workloadPool))
	} else {
		ctx.Export(OpWorkloadIdentityPool, pulumi.String(""))
	}
	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpName, createdCluster.Name)
	// The plain spec location (not the provider's computed attribute) so
	// both engines emit the identical value for API callers.
	ctx.Export(OpLocation, pulumi.String(spec.Location))
	ctx.Export(OpSelfLink, createdCluster.SelfLink)
	ctx.Export(OpMasterVersion, createdCluster.MasterVersion)

	return nil
}

// buildClusterContainerdDefaults translates the cluster-level containerd
// defaults (node_pool_defaults.node_config_defaults.containerd_config):
// private-CA registry trust, per-registry host overrides, and writable
// cgroups — containerd hosts.toml semantics carried through the GKE API.
func buildClusterContainerdDefaults(containerd *gcpgkeclusterv1alpha1.GcpGkeClusterContainerdDefaults) *container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigArgs {
	containerdArgs := &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigArgs{}

	if access := containerd.PrivateRegistryAccess; access != nil {
		accessArgs := &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigPrivateRegistryAccessConfigArgs{
			Enabled: pulumi.Bool(access.Enabled),
		}
		if len(access.CertificateAuthorityDomains) > 0 {
			domains := container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigPrivateRegistryAccessConfigCertificateAuthorityDomainConfigArray{}
			for _, domain := range access.CertificateAuthorityDomains {
				domains = append(domains, &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigPrivateRegistryAccessConfigCertificateAuthorityDomainConfigArgs{
					Fqdns: pulumi.ToStringArray(domain.Fqdns),
					GcpSecretManagerCertificateConfig: &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigPrivateRegistryAccessConfigCertificateAuthorityDomainConfigGcpSecretManagerCertificateConfigArgs{
						SecretUri: pulumi.String(domain.GcpSecretManagerCertificateUri),
					},
				})
			}
			accessArgs.CertificateAuthorityDomainConfigs = domains
		}
		containerdArgs.PrivateRegistryAccessConfig = accessArgs
	}

	if len(containerd.RegistryHosts) > 0 {
		registries := container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostArray{}
		for _, registry := range containerd.RegistryHosts {
			hosts := container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostArray{}
			for _, endpoint := range registry.Hosts {
				hostArgs := &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostArgs{
					Host: pulumi.String(endpoint.Host),
				}
				if len(endpoint.Capabilities) > 0 {
					hostArgs.Capabilities = pulumi.ToStringArray(endpoint.Capabilities)
				}
				if endpoint.DialTimeout != "" {
					hostArgs.DialTimeout = pulumi.StringPtr(endpoint.DialTimeout)
				}
				if endpoint.OverridePath != nil {
					hostArgs.OverridePath = pulumi.BoolPtr(*endpoint.OverridePath)
				}
				// The bridged SDK models ca/client as single-element lists
				// and header values as a list; the semantic payload is
				// identical to the Terraform module's single-value form.
				if endpoint.CaSecretUri != "" {
					hostArgs.Cas = container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostCaArray{
						&container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostCaArgs{
							GcpSecretManagerSecretUri: pulumi.StringPtr(endpoint.CaSecretUri),
						},
					}
				}
				if endpoint.ClientCertSecretUri != "" || endpoint.ClientKeySecretUri != "" {
					clientArgs := &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostClientArgs{
						// The SDK requires the cert entry; a key-only client
						// is meaningless TLS anyway.
						Cert: &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostClientCertArgs{
							GcpSecretManagerSecretUri: pulumi.StringPtr(endpoint.ClientCertSecretUri),
						},
					}
					if endpoint.ClientKeySecretUri != "" {
						clientArgs.Key = &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostClientKeyArgs{
							GcpSecretManagerSecretUri: pulumi.StringPtr(endpoint.ClientKeySecretUri),
						}
					}
					hostArgs.Clients = container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostClientArray{clientArgs}
				}
				if len(endpoint.Headers) > 0 {
					headers := container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostHeaderArray{}
					for key, value := range endpoint.Headers {
						headers = append(headers, &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostHostHeaderArgs{
							Key:    pulumi.String(key),
							Values: pulumi.ToStringArray([]string{value}),
						})
					}
					hostArgs.Headers = headers
				}
				hosts = append(hosts, hostArgs)
			}
			registries = append(registries, &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigRegistryHostArgs{
				Server: pulumi.String(registry.Server),
				Hosts:  hosts,
			})
		}
		containerdArgs.RegistryHosts = registries
	}

	if containerd.WritableCgroupsEnabled != nil {
		containerdArgs.WritableCgroups = &container.ClusterNodePoolDefaultsNodeConfigDefaultsContainerdConfigWritableCgroupsArgs{
			Enabled: pulumi.Bool(*containerd.WritableCgroupsEnabled),
		}
	}

	return containerdArgs
}
