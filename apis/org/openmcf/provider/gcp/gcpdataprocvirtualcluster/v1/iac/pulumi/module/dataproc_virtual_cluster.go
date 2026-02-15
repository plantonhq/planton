package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/dataproc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dataprocVirtualCluster provisions a Dataproc on GKE virtual cluster.
func dataprocVirtualCluster(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDataprocVirtualCluster.Spec

	// Determine cluster name: explicit spec field, else metadata name.
	clusterName := spec.ClusterName
	if clusterName == "" && locals.GcpDataprocVirtualCluster.Metadata != nil {
		clusterName = locals.GcpDataprocVirtualCluster.Metadata.Name
	}

	// -- Build node pool targets ------------------------------------------

	nodePoolTargets := dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetArray{}
	for _, npt := range spec.NodePoolTargets {
		target := dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetArgs{
			NodePool: pulumi.String(npt.NodePool.GetValue()),
			Roles:    pulumi.ToStringArray(npt.Roles),
		}

		if npt.NodePoolConfig != nil {
			npc := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetNodePoolConfigArgs{}

			if len(npt.NodePoolConfig.Locations) > 0 {
				npc.Locations = pulumi.ToStringArray(npt.NodePoolConfig.Locations)
			}

			if npt.NodePoolConfig.Autoscaling != nil {
				npc.Autoscaling = &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetNodePoolConfigAutoscalingArgs{
					MinNodeCount: pulumi.IntPtr(int(npt.NodePoolConfig.Autoscaling.MinNodeCount)),
					MaxNodeCount: pulumi.IntPtr(int(npt.NodePoolConfig.Autoscaling.MaxNodeCount)),
				}
			}

			cfg := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetNodePoolConfigConfigArgs{}
			hasConfig := false

			if npt.NodePoolConfig.MachineType != "" {
				cfg.MachineType = pulumi.StringPtr(npt.NodePoolConfig.MachineType)
				hasConfig = true
			}
			if npt.NodePoolConfig.LocalSsdCount > 0 {
				cfg.LocalSsdCount = pulumi.IntPtr(int(npt.NodePoolConfig.LocalSsdCount))
				hasConfig = true
			}
			if npt.NodePoolConfig.MinCpuPlatform != "" {
				cfg.MinCpuPlatform = pulumi.StringPtr(npt.NodePoolConfig.MinCpuPlatform)
				hasConfig = true
			}
			if npt.NodePoolConfig.Preemptible {
				cfg.Preemptible = pulumi.BoolPtr(true)
				hasConfig = true
			}
			if npt.NodePoolConfig.Spot {
				cfg.Spot = pulumi.BoolPtr(true)
				hasConfig = true
			}

			if hasConfig {
				npc.Config = cfg
			}

			target.NodePoolConfig = npc
		}

		nodePoolTargets = append(nodePoolTargets, target)
	}

	// -- Build Kubernetes software config ---------------------------------

	softwareCfg := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigKubernetesSoftwareConfigArgs{
		ComponentVersion: pulumi.ToStringMap(spec.SoftwareConfig.ComponentVersion),
	}
	if len(spec.SoftwareConfig.Properties) > 0 {
		softwareCfg.Properties = pulumi.ToStringMap(spec.SoftwareConfig.Properties)
	}

	// -- Build GKE cluster config -----------------------------------------

	gkeClusterCfg := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigArgs{
		GkeClusterTarget: pulumi.StringPtr(spec.GkeClusterTarget.GetValue()),
		NodePoolTargets:  nodePoolTargets,
	}

	// -- Build Kubernetes cluster config ----------------------------------

	k8sClusterCfg := &dataproc.ClusterVirtualClusterConfigKubernetesClusterConfigArgs{
		GkeClusterConfig:        gkeClusterCfg,
		KubernetesSoftwareConfig: softwareCfg,
	}
	if spec.KubernetesNamespace != nil && spec.KubernetesNamespace.GetValue() != "" {
		k8sClusterCfg.KubernetesNamespace = pulumi.StringPtr(spec.KubernetesNamespace.GetValue())
	}

	// -- Build virtual cluster config -------------------------------------

	virtualCfg := &dataproc.ClusterVirtualClusterConfigArgs{
		KubernetesClusterConfig: k8sClusterCfg,
	}
	if spec.StagingBucket != nil && spec.StagingBucket.GetValue() != "" {
		virtualCfg.StagingBucket = pulumi.StringPtr(spec.StagingBucket.GetValue())
	}

	// -- Auxiliary services ------------------------------------------------

	if spec.AuxiliaryServicesConfig != nil {
		auxCfg := &dataproc.ClusterVirtualClusterConfigAuxiliaryServicesConfigArgs{}
		hasAux := false

		if spec.AuxiliaryServicesConfig.MetastoreService != "" {
			auxCfg.MetastoreConfig = &dataproc.ClusterVirtualClusterConfigAuxiliaryServicesConfigMetastoreConfigArgs{
				DataprocMetastoreService: pulumi.StringPtr(spec.AuxiliaryServicesConfig.MetastoreService),
			}
			hasAux = true
		}
		if spec.AuxiliaryServicesConfig.SparkHistoryServerCluster != "" {
			auxCfg.SparkHistoryServerConfig = &dataproc.ClusterVirtualClusterConfigAuxiliaryServicesConfigSparkHistoryServerConfigArgs{
				DataprocCluster: pulumi.StringPtr(spec.AuxiliaryServicesConfig.SparkHistoryServerCluster),
			}
			hasAux = true
		}

		if hasAux {
			virtualCfg.AuxiliaryServicesConfig = auxCfg
		}
	}

	// -- Create the Dataproc cluster with virtual config ------------------

	args := &dataproc.ClusterArgs{
		Name:                 pulumi.String(clusterName),
		Region:               pulumi.String(spec.Region),
		Project:              pulumi.String(spec.ProjectId.GetValue()),
		Labels:               pulumi.ToStringMap(locals.GcpLabels),
		VirtualClusterConfig: virtualCfg,
	}

	createdCluster, err := dataproc.NewCluster(ctx, "dataproc-virtual-cluster", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create dataproc virtual cluster")
	}

	// -- Outputs ----------------------------------------------------------

	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpClusterName, createdCluster.Name)
	// cluster_uuid is not directly exposed by the Pulumi/Terraform provider.
	// Downstream consumers should use cluster_id for resource references.
	ctx.Export(OpClusterUuid, pulumi.String(""))

	return nil
}
