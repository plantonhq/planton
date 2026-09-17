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

	// Pulumi SDK v4.53.0 gaps: these spec fields are modeled and the
	// Terraform module wires them, but the SDK has no matching inputs on
	// KubernetesCluster (verified against KubernetesClusterArgs and
	// KubernetesClusterNodePoolArgs at that pin). Fail loudly on a
	// meaningful set (proto zero values pass) rather than silently dropping
	// configuration. Every guard names the pin it was verified against and
	// is re-checked on every SDK bump: guards for surfaces the SDK has since
	// grown are deleted and the arm wired, never left in place.
	if spec.WorkerSubnetUuid != "" {
		return nil, errors.New("PARITY-EXCEPTION: spec.worker_subnet_uuid is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no worker_subnet_uuid field on KubernetesCluster. Re-evaluate when the SDK exposes worker_subnet_uuid.")
	}
	if spec.IsolatedWorkers {
		return nil, errors.New("PARITY-EXCEPTION: spec.isolated_workers is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no isolated_workers field on KubernetesCluster. Re-evaluate when the SDK exposes isolated_workers.")
	}
	if spec.Sso != nil {
		return nil, errors.New("PARITY-EXCEPTION: spec.sso is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no sso block on KubernetesCluster. Re-evaluate when the SDK exposes sso.")
	}
	if spec.P2POciRegistryPlugin != nil {
		return nil, errors.New("PARITY-EXCEPTION: spec.p2p_oci_registry_plugin is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no p2p_oci_registry_plugin block on KubernetesCluster. Re-evaluate when the SDK exposes p2p_oci_registry_plugin.")
	}
	if spec.AmdGpuDraDriver != nil {
		return nil, errors.New("PARITY-EXCEPTION: spec.amd_gpu_dra_driver is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no amd_gpu_dra_driver block on KubernetesCluster. Re-evaluate when the SDK exposes amd_gpu_dra_driver.")
	}
	if spec.NvidiaGpuDevicePlugin != nil {
		return nil, errors.New("PARITY-EXCEPTION: spec.nvidia_gpu_device_plugin is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no nvidia_gpu_device_plugin block on KubernetesCluster. Re-evaluate when the SDK exposes nvidia_gpu_device_plugin.")
	}
	if spec.NvidiaGpuDraDriver != nil {
		return nil, errors.New("PARITY-EXCEPTION: spec.nvidia_gpu_dra_driver is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no nvidia_gpu_dra_driver block on KubernetesCluster. Re-evaluate when the SDK exposes nvidia_gpu_dra_driver.")
	}
	if spec.RdmaSharedDevicePlugin != nil {
		return nil, errors.New("PARITY-EXCEPTION: spec.rdma_shared_device_plugin is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no rdma_shared_device_plugin block on KubernetesCluster. Re-evaluate when the SDK exposes rdma_shared_device_plugin.")
	}
	if spec.CorednsAutoscaler != nil {
		return nil, errors.New("PARITY-EXCEPTION: spec.coredns_autoscaler is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no coredns_autoscaler block on KubernetesCluster. Re-evaluate when the SDK exposes coredns_autoscaler.")
	}
	if spec.DefaultNodePool.GpuPartitionMode != "" {
		return nil, errors.New("PARITY-EXCEPTION: spec.default_node_pool.gpu_partition_mode is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.53.0 has no gpu_partition_mode field on the cluster's node pool. Re-evaluate when the SDK exposes gpu_partition_mode.")
	}

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
		NodeCount: pulumi.IntPtr(int(spec.DefaultNodePool.NodeCount)),
		AutoScale: pulumi.BoolPtr(spec.DefaultNodePool.AutoScale),
		Labels:    poolLabels,
	}
	// Autoscaler bounds only travel with autoscaling on -- matching the
	// Terraform module, which nulls them otherwise.
	if spec.DefaultNodePool.AutoScale {
		poolArgs.MinNodes = pulumi.IntPtr(int(spec.DefaultNodePool.MinNodes))
		poolArgs.MaxNodes = pulumi.IntPtr(int(spec.DefaultNodePool.MaxNodes))
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
	// dynamic blocks. The SDK carries three of the nine addon blocks (routing
	// agent, AMD GPU device plugin, AMD GPU device metrics exporter); the
	// other six fail loudly above.
	if spec.RoutingAgent != nil {
		clusterArgs.RoutingAgent = &digitalocean.KubernetesClusterRoutingAgentArgs{
			Enabled: pulumi.Bool(spec.RoutingAgent.GetEnabled()),
		}
	}
	if spec.AmdGpuDevicePlugin != nil {
		clusterArgs.AmdGpuDevicePlugin = &digitalocean.KubernetesClusterAmdGpuDevicePluginArgs{
			Enabled: pulumi.Bool(spec.AmdGpuDevicePlugin.GetEnabled()),
		}
	}
	if spec.AmdGpuDeviceMetricsExporterPlugin != nil {
		clusterArgs.AmdGpuDeviceMetricsExporterPlugin = &digitalocean.KubernetesClusterAmdGpuDeviceMetricsExporterPluginArgs{
			Enabled: pulumi.Bool(spec.AmdGpuDeviceMetricsExporterPlugin.GetEnabled()),
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
