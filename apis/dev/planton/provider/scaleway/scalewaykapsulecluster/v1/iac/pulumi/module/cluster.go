package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/kubernetes"
)

// cluster provisions the complete KapsuleCluster composite:
//
//  1. The Kapsule cluster (managed Kubernetes control plane) with CNI,
//     Private Network, optional auto-upgrade, and optional autoscaler config.
//  2. The default node pool (with optional autoscaling, autohealing, and
//     upgrade policy).
//
// All outputs (cluster_id, kubeconfig, apiserver_url, CA cert, wildcard DNS,
// default pool ID) are exported for downstream resource references.
func cluster(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) error {
	spec := locals.ScalewayKapsuleCluster.Spec
	name := locals.ScalewayKapsuleCluster.Metadata.Name

	// ── 1. Build cluster arguments ────────────────────────────────────────

	clusterArgs := &kubernetes.ClusterArgs{
		Name:                      pulumi.String(name),
		Version:                   pulumi.String(spec.KubernetesVersion),
		Cni:                       pulumi.String(spec.Cni),
		PrivateNetworkId:          pulumi.String(locals.PrivateNetworkId),
		DeleteAdditionalResources: pulumi.Bool(spec.DeleteAdditionalResources),
		Tags:                      pulumi.ToStringArray(locals.ScalewayTags),
		Region:                    pulumi.String(spec.Region),
	}

	// Cluster type (optional, defaults to "kapsule" on the Scaleway side).
	if spec.Type != "" {
		clusterArgs.Type = pulumi.String(spec.Type)
	}

	// Description (optional).
	if spec.Description != "" {
		clusterArgs.Description = pulumi.String(spec.Description)
	}

	// Feature gates (optional).
	if len(spec.FeatureGates) > 0 {
		clusterArgs.FeatureGates = pulumi.ToStringArray(spec.FeatureGates)
	}

	// Admission plugins (optional).
	if len(spec.AdmissionPlugins) > 0 {
		clusterArgs.AdmissionPlugins = pulumi.ToStringArray(spec.AdmissionPlugins)
	}

	// Pod CIDR (optional, ForceNew).
	if spec.PodCidr != "" {
		clusterArgs.PodCidr = pulumi.String(spec.PodCidr)
	}

	// Service CIDR (optional, ForceNew).
	if spec.ServiceCidr != "" {
		clusterArgs.ServiceCidr = pulumi.String(spec.ServiceCidr)
	}

	// Auto-upgrade configuration (optional).
	if spec.AutoUpgrade != nil {
		clusterArgs.AutoUpgrade = &kubernetes.ClusterAutoUpgradeArgs{
			Enable:                     pulumi.Bool(spec.AutoUpgrade.Enable),
			MaintenanceWindowStartHour: pulumi.Int(int(spec.AutoUpgrade.MaintenanceWindowStartHour)),
			MaintenanceWindowDay:       pulumi.String(spec.AutoUpgrade.MaintenanceWindowDay),
		}
	}

	// Autoscaler configuration (optional).
	if spec.AutoscalerConfig != nil {
		autoscalerArgs := &kubernetes.ClusterAutoscalerConfigArgs{}

		if spec.AutoscalerConfig.DisableScaleDown {
			autoscalerArgs.DisableScaleDown = pulumi.Bool(true)
		}

		if spec.AutoscalerConfig.ScaleDownDelayAfterAdd != "" {
			autoscalerArgs.ScaleDownDelayAfterAdd = pulumi.String(spec.AutoscalerConfig.ScaleDownDelayAfterAdd)
		}

		if spec.AutoscalerConfig.ScaleDownUnneededTime != "" {
			autoscalerArgs.ScaleDownUnneededTime = pulumi.String(spec.AutoscalerConfig.ScaleDownUnneededTime)
		}

		if spec.AutoscalerConfig.Estimator != "" {
			autoscalerArgs.Estimator = pulumi.String(spec.AutoscalerConfig.Estimator)
		}

		if spec.AutoscalerConfig.Expander != "" {
			autoscalerArgs.Expander = pulumi.String(spec.AutoscalerConfig.Expander)
		}

		if spec.AutoscalerConfig.ScaleDownUtilizationThreshold != 0 {
			autoscalerArgs.ScaleDownUtilizationThreshold = pulumi.Float64(spec.AutoscalerConfig.ScaleDownUtilizationThreshold)
		}

		if spec.AutoscalerConfig.MaxGracefulTerminationSec != 0 {
			autoscalerArgs.MaxGracefulTerminationSec = pulumi.Int(int(spec.AutoscalerConfig.MaxGracefulTerminationSec))
		}

		if spec.AutoscalerConfig.IgnoreDaemonsetsUtilization {
			autoscalerArgs.IgnoreDaemonsetsUtilization = pulumi.Bool(true)
		}

		if spec.AutoscalerConfig.BalanceSimilarNodeGroups {
			autoscalerArgs.BalanceSimilarNodeGroups = pulumi.Bool(true)
		}

		if spec.AutoscalerConfig.ExpendablePodsPriorityCutoff != 0 {
			autoscalerArgs.ExpendablePodsPriorityCutoff = pulumi.Int(int(spec.AutoscalerConfig.ExpendablePodsPriorityCutoff))
		}

		clusterArgs.AutoscalerConfig = autoscalerArgs
	}

	// ── 2. Create the Kapsule cluster ─────────────────────────────────────

	createdCluster, err := kubernetes.NewCluster(
		ctx,
		"cluster",
		clusterArgs,
		pulumi.Provider(scalewayProvider),
		// Ignore version changes to prevent drift when auto-upgrade is
		// enabled. When Scaleway patches the cluster (e.g., 1.32.1 -> 1.32.3),
		// the IaC state should not try to revert it.
		pulumi.IgnoreChanges([]string{"version"}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create kapsule cluster")
	}

	// ── 3. Export cluster outputs ─────────────────────────────────────────

	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpApiserverUrl, createdCluster.ApiserverUrl)
	ctx.Export(OpWildcardDns, createdCluster.WildcardDns)

	// Extract kubeconfig and CA certificate from the first kubeconfig entry.
	ctx.Export(OpKubeconfig, createdCluster.Kubeconfigs.ApplyT(
		func(kubeconfigs []kubernetes.ClusterKubeconfig) string {
			if len(kubeconfigs) > 0 && kubeconfigs[0].ConfigFile != nil {
				return *kubeconfigs[0].ConfigFile
			}
			return ""
		},
	).(pulumi.StringOutput))

	ctx.Export(OpClusterCaCertificate, createdCluster.Kubeconfigs.ApplyT(
		func(kubeconfigs []kubernetes.ClusterKubeconfig) string {
			if len(kubeconfigs) > 0 && kubeconfigs[0].ClusterCaCertificate != nil {
				return *kubeconfigs[0].ClusterCaCertificate
			}
			return ""
		},
	).(pulumi.StringOutput))

	// ── 4. Create the default node pool ───────────────────────────────────

	pool := spec.DefaultNodePool

	poolName := pool.Name
	if poolName == "" {
		poolName = fmt.Sprintf("%s-default", name)
	}

	poolArgs := &kubernetes.PoolArgs{
		ClusterId: createdCluster.ID(),
		Name:      pulumi.String(poolName),
		NodeType:  pulumi.String(pool.NodeType),
		Size:      pulumi.Int(int(pool.Size)),
		Tags:      pulumi.ToStringArray(locals.ScalewayTags),
	}

	// Autoscaling configuration.
	if pool.AutoScale {
		poolArgs.Autoscaling = pulumi.Bool(true)
		if pool.MinSize > 0 {
			poolArgs.MinSize = pulumi.Int(int(pool.MinSize))
		}
		if pool.MaxSize > 0 {
			poolArgs.MaxSize = pulumi.Int(int(pool.MaxSize))
		}
	}

	// Autohealing.
	if pool.Autohealing {
		poolArgs.Autohealing = pulumi.Bool(true)
	}

	// Container runtime.
	if pool.ContainerRuntime != "" {
		poolArgs.ContainerRuntime = pulumi.String(pool.ContainerRuntime)
	}

	// Root volume configuration.
	if pool.RootVolumeType != "" {
		poolArgs.RootVolumeType = pulumi.String(pool.RootVolumeType)
	}
	if pool.RootVolumeSizeInGb > 0 {
		poolArgs.RootVolumeSizeInGb = pulumi.Int(int(pool.RootVolumeSizeInGb))
	}

	// Disable public IPs on nodes.
	if pool.PublicIpDisabled {
		poolArgs.PublicIpDisabled = pulumi.Bool(true)
	}

	// Upgrade policy.
	if pool.UpgradePolicy != nil {
		poolArgs.UpgradePolicy = &kubernetes.PoolUpgradePolicyArgs{
			MaxSurge:       pulumi.Int(int(pool.UpgradePolicy.MaxSurge)),
			MaxUnavailable: pulumi.Int(int(pool.UpgradePolicy.MaxUnavailable)),
		}
	}

	createdPool, err := kubernetes.NewPool(
		ctx,
		"default-pool",
		poolArgs,
		pulumi.Provider(scalewayProvider),
		// Wait for the cluster to be ready before creating the pool.
		pulumi.DependsOn([]pulumi.Resource{createdCluster}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create default node pool")
	}

	// ── 5. Export default pool output ─────────────────────────────────────

	ctx.Export(OpDefaultPoolId, createdPool.ID())

	return nil
}
