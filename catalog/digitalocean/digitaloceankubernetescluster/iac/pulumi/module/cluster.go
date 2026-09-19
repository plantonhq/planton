package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// cluster provisions the Kubernetes cluster and exports its outputs.
func cluster(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.KubernetesCluster, error) {
	spec := locals.DigitalOceanKubernetesCluster.Spec

	// User tags plus the standard Planton labels rendered as "key:value"
	// tags — the exact set the Terraform module applies.
	tagSet := map[string]bool{}
	var tagInputs pulumi.StringArray
	for _, t := range spec.Tags {
		if !tagSet[t] {
			tagSet[t] = true
			tagInputs = append(tagInputs, pulumi.String(t))
		}
	}
	for k, v := range locals.DigitalOceanLabels {
		t := k + ":" + v
		if !tagSet[t] {
			tagSet[t] = true
			tagInputs = append(tagInputs, pulumi.String(t))
		}
	}

	// Kubernetes node labels on the default pool: user labels over the
	// standard Planton labels — the exact map the Terraform module applies.
	poolLabels := pulumi.StringMap{}
	for k, v := range locals.DigitalOceanLabels {
		poolLabels[k] = pulumi.String(v)
	}
	for k, v := range spec.DefaultNodePool.Labels {
		poolLabels[k] = pulumi.String(v)
	}

	// The inline default node pool. Its name is synthesized -- the pool has
	// no independent identity; additional pools are the separate
	// DigitalOceanKubernetesNodePool kind.
	poolArgs := &digitalocean.KubernetesClusterNodePoolArgs{
		Name:      pulumi.String("default"),
		Size:      pulumi.String(spec.DefaultNodePool.Size),
		AutoScale: pulumi.BoolPtr(spec.DefaultNodePool.AutoScale),
		Labels:    poolLabels,
	}
	// Exactly one sizing mode owns the count -- matching the Terraform
	// module. A fixed pool sends node_count; an autoscaled pool sends only
	// the bounds and NO count, because the provider writes the live count
	// back into node_count on every read and re-applies a stated one on
	// every update, so a stated count and the autoscaler would fight forever
	// (measured: a pool that autoscaled to two nodes planned `2 -> 1`).
	// Without a count the API starts the pool at min_nodes.
	if spec.DefaultNodePool.AutoScale {
		poolArgs.MinNodes = pulumi.IntPtr(int(spec.DefaultNodePool.MinNodes))
		poolArgs.MaxNodes = pulumi.IntPtr(int(spec.DefaultNodePool.MaxNodes))
	} else {
		poolArgs.NodeCount = pulumi.IntPtr(int(spec.DefaultNodePool.NodeCount))
	}
	if len(spec.DefaultNodePool.Tags) > 0 {
		var poolTags pulumi.StringArray
		for _, t := range spec.DefaultNodePool.Tags {
			poolTags = append(poolTags, pulumi.String(t))
		}
		poolArgs.Tags = poolTags
	}
	if len(spec.DefaultNodePool.Taints) > 0 {
		var taints digitalocean.KubernetesClusterNodePoolTaintArray
		for _, t := range spec.DefaultNodePool.Taints {
			taints = append(taints, digitalocean.KubernetesClusterNodePoolTaintArgs{
				Key:    pulumi.String(t.Key),
				Value:  pulumi.String(t.Value),
				Effect: pulumi.String(t.Effect),
			})
		}
		poolArgs.Taints = taints
	}
	// GPU partitioning is create-only on the pool and only meaningful on GPU
	// sizes; unset must arrive as null, never "" (the provider rejects it).
	if spec.DefaultNodePool.GpuPartitionMode != "" {
		poolArgs.GpuPartitionMode = pulumi.StringPtr(spec.DefaultNodePool.GpuPartitionMode)
	}

	// Enum value names are exactly the DigitalOcean region slugs.
	clusterArgs := &digitalocean.KubernetesClusterArgs{
		Name:    pulumi.String(spec.ClusterName),
		Region:  pulumi.String(spec.Region.String()),
		Version: pulumi.String(spec.KubernetesVersion),
		VpcUuid: pulumi.String(spec.Vpc.GetValue()),
		// HA is one-way; an explicit false (proto3 unset) keeps the cheaper
		// single-replica control plane even on DOKS versions whose
		// server-side default is HA on.
		Ha:                  pulumi.BoolPtr(spec.HighlyAvailable),
		AutoUpgrade:         pulumi.BoolPtr(spec.AutoUpgrade),
		RegistryIntegration: pulumi.BoolPtr(spec.RegistryIntegration),
		Tags:                tagInputs,
		NodePool:            poolArgs,
	}

	// Sent only when present: unset defers to the provider's default (true),
	// matching DigitalOcean's own surge-upgrade default. Never coalesce to
	// false.
	if spec.SurgeUpgrade != nil {
		clusterArgs.SurgeUpgrade = pulumi.BoolPtr(spec.GetSurgeUpgrade())
	}

	// Day is lowercased because the provider accepts any case but reads back
	// lowercase -- mixed case would drift.
	if spec.MaintenancePolicy != nil {
		clusterArgs.MaintenancePolicy = &digitalocean.KubernetesClusterMaintenancePolicyArgs{
			Day:       pulumi.StringPtr(strings.ToLower(spec.MaintenancePolicy.Day)),
			StartTime: pulumi.StringPtr(spec.MaintenancePolicy.StartTime),
		}
	}

	if spec.ControlPlaneFirewall != nil {
		var allowedAddresses pulumi.StringArray
		for _, a := range spec.ControlPlaneFirewall.AllowedAddresses {
			allowedAddresses = append(allowedAddresses, pulumi.String(a))
		}
		clusterArgs.ControlPlaneFirewall = &digitalocean.KubernetesClusterControlPlaneFirewallArgs{
			Enabled:          pulumi.Bool(spec.ControlPlaneFirewall.GetEnabled()),
			AllowedAddresses: allowedAddresses,
		}
	}

	// Create-only network placement; the provider rejects empty strings.
	if spec.ClusterSubnet != "" {
		clusterArgs.ClusterSubnet = pulumi.StringPtr(spec.ClusterSubnet)
	}
	if spec.ServiceSubnet != "" {
		clusterArgs.ServiceSubnet = pulumi.StringPtr(spec.ServiceSubnet)
	}
	// Worker placement: an explicit VPC subnet for the nodes (create-only;
	// the provider rejects ""), and isolated workers (no public IPs on the
	// nodes) sent as the spec's bool exactly as the Terraform module does.
	if spec.WorkerSubnetUuid != "" {
		clusterArgs.WorkerSubnetUuid = pulumi.StringPtr(spec.WorkerSubnetUuid)
	}
	clusterArgs.IsolatedWorkers = pulumi.BoolPtr(spec.IsolatedWorkers)

	// Single sign-on. The SDK models the block as an array (the provider
	// reads only the first element), so the spec's single message is wrapped
	// here -- the cluster-autoscaler-configuration shape. Empty issuer/client
	// strings arrive as null, matching the Terraform module's coalescing.
	if spec.Sso != nil {
		ssoArgs := digitalocean.KubernetesClusterSsoArgs{
			Enabled:  pulumi.Bool(spec.Sso.GetEnabled()),
			Required: pulumi.BoolPtr(spec.Sso.Required),
		}
		if spec.Sso.IssuerUrl != "" {
			ssoArgs.IssuerUrl = pulumi.StringPtr(spec.Sso.IssuerUrl)
		}
		if spec.Sso.ClientId != "" {
			ssoArgs.ClientId = pulumi.StringPtr(spec.Sso.ClientId)
		}
		clusterArgs.Ssos = digitalocean.KubernetesClusterSsoArray{ssoArgs}
	}

	if spec.DestroyAllAssociatedResources {
		clusterArgs.DestroyAllAssociatedResources = pulumi.BoolPtr(true)
	}

	// 0 (proto3 unset) means DigitalOcean's 7-day default credential
	// validity; keep it out of state rather than pinning an explicit zero.
	if spec.KubeconfigExpireSeconds > 0 {
		clusterArgs.KubeconfigExpireSeconds = pulumi.IntPtr(int(spec.KubeconfigExpireSeconds))
	}

	// The SDK models the autoscaler configuration as an array; the provider
	// only ever reads the first element, so the spec carries a single
	// message wrapped here.
	if spec.ClusterAutoscalerConfiguration != nil {
		caArgs := digitalocean.KubernetesClusterClusterAutoscalerConfigurationArgs{}
		if spec.ClusterAutoscalerConfiguration.ScaleDownUtilizationThreshold != nil {
			caArgs.ScaleDownUtilizationThreshold = pulumi.Float64Ptr(spec.ClusterAutoscalerConfiguration.GetScaleDownUtilizationThreshold())
		}
		if spec.ClusterAutoscalerConfiguration.ScaleDownUnneededTime != "" {
			caArgs.ScaleDownUnneededTime = pulumi.StringPtr(spec.ClusterAutoscalerConfiguration.ScaleDownUnneededTime)
		}
		if len(spec.ClusterAutoscalerConfiguration.Expanders) > 0 {
			var expanders pulumi.StringArray
			for _, e := range spec.ClusterAutoscalerConfiguration.Expanders {
				expanders = append(expanders, pulumi.String(e))
			}
			caArgs.Expanders = expanders
		}
		clusterArgs.ClusterAutoscalerConfigurations = digitalocean.KubernetesClusterClusterAutoscalerConfigurationArray{caArgs}
	}

	// Managed addon toggles. An unset spec message sends no block, deferring
	// to DigitalOcean's own default for that addon; a set message asserts the
	// desired state, on or off -- the same contract as the Terraform module's
	// dynamic blocks. All nine addon blocks are wired; the GPU-family blocks
	// (AMD/NVIDIA device plugins and DRA drivers, RDMA) are accepted by the
	// API only on clusters with GPU node pools, and the P2P OCI registry
	// plugin only on Kubernetes 1.36.0-do.2 or later -- an older version
	// fails the whole create with a validation 422 that creates nothing.
	// The module sends what the manifest states and lets DigitalOcean's own
	// validation speak, because the version floor moves with DigitalOcean's
	// release train and a module-side check would go stale.
	if spec.RoutingAgent != nil {
		clusterArgs.RoutingAgent = &digitalocean.KubernetesClusterRoutingAgentArgs{
			Enabled: pulumi.Bool(spec.RoutingAgent.GetEnabled()),
		}
	}
	if spec.P2POciRegistryPlugin != nil {
		clusterArgs.P2pOciRegistryPlugin = &digitalocean.KubernetesClusterP2pOciRegistryPluginArgs{
			Enabled: pulumi.Bool(spec.P2POciRegistryPlugin.GetEnabled()),
		}
	}
	if spec.AmdGpuDevicePlugin != nil {
		clusterArgs.AmdGpuDevicePlugin = &digitalocean.KubernetesClusterAmdGpuDevicePluginArgs{
			Enabled: pulumi.Bool(spec.AmdGpuDevicePlugin.GetEnabled()),
		}
	}
	if spec.AmdGpuDraDriver != nil {
		clusterArgs.AmdGpuDraDriver = &digitalocean.KubernetesClusterAmdGpuDraDriverArgs{
			Enabled: pulumi.Bool(spec.AmdGpuDraDriver.GetEnabled()),
		}
	}
	if spec.AmdGpuDeviceMetricsExporterPlugin != nil {
		clusterArgs.AmdGpuDeviceMetricsExporterPlugin = &digitalocean.KubernetesClusterAmdGpuDeviceMetricsExporterPluginArgs{
			Enabled: pulumi.Bool(spec.AmdGpuDeviceMetricsExporterPlugin.GetEnabled()),
		}
	}
	if spec.NvidiaGpuDevicePlugin != nil {
		clusterArgs.NvidiaGpuDevicePlugin = &digitalocean.KubernetesClusterNvidiaGpuDevicePluginArgs{
			Enabled: pulumi.Bool(spec.NvidiaGpuDevicePlugin.GetEnabled()),
		}
	}
	if spec.NvidiaGpuDraDriver != nil {
		clusterArgs.NvidiaGpuDraDriver = &digitalocean.KubernetesClusterNvidiaGpuDraDriverArgs{
			Enabled: pulumi.Bool(spec.NvidiaGpuDraDriver.GetEnabled()),
		}
	}
	if spec.RdmaSharedDevicePlugin != nil {
		clusterArgs.RdmaSharedDevicePlugin = &digitalocean.KubernetesClusterRdmaSharedDevicePluginArgs{
			Enabled: pulumi.Bool(spec.RdmaSharedDevicePlugin.GetEnabled()),
		}
	}
	if spec.CorednsAutoscaler != nil {
		clusterArgs.CorednsAutoscaler = &digitalocean.KubernetesClusterCorednsAutoscalerArgs{
			Enabled: pulumi.Bool(spec.CorednsAutoscaler.GetEnabled()),
		}
	}

	// Auto-upgrade moves the live version ahead of the configured pin, and
	// the provider DESTROYS AND RECREATES the cluster when the configured
	// version is lower than the live one. Ignoring version drift makes the
	// pin creation-only.
	createdCluster, err := digitalocean.NewKubernetesCluster(
		ctx,
		"cluster",
		clusterArgs,
		pulumi.Provider(digitalOceanProvider),
		pulumi.IgnoreChanges([]string{"version"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean kubernetes cluster")
	}

	ctx.Export(OpClusterId, createdCluster.ID())
	// The raw kubeconfig YAML (registered as a secret output by the SDK).
	ctx.Export(OpKubeconfig, createdCluster.KubeConfigs.Index(pulumi.Int(0)).RawConfig())
	ctx.Export(OpApiServerEndpoint, createdCluster.Endpoint)
	ctx.Export(OpUrn, createdCluster.ClusterUrn)
	ctx.Export(OpIpv4Address, createdCluster.Ipv4Address)
	ctx.Export(OpDefaultNodePoolId, createdCluster.NodePool.Id())
	ctx.Export(OpClusterSubnet, createdCluster.ClusterSubnet)
	ctx.Export(OpServiceSubnet, createdCluster.ServiceSubnet)

	return createdCluster, nil
}
