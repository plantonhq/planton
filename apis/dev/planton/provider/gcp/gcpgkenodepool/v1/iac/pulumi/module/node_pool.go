package module

import (
	"github.com/pkg/errors"
	gcpgkenodepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpgkenodepool/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/container"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nodePool creates a single GKE node‑pool for the parent cluster.
// All inputs come straight from locals.GcpGkeNodePool.Spec.
func nodePool(ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider) error {

	spec := locals.GcpGkeNodePool.Spec

	// ----- Node‑count vs. autoscaling ------------------------------------

	var autoscaling *container.NodePoolAutoscalingArgs
	var nodeCount pulumi.IntPtrInput

	if spec.NodePoolSize != nil {
		switch x := spec.NodePoolSize.(type) {
		case *gcpgkenodepoolv1.GcpGkeNodePoolSpec_NodeCount:
			// Fixed size
			nodeCount = pulumi.IntPtr(int(x.NodeCount))
		case *gcpgkenodepoolv1.GcpGkeNodePoolSpec_Autoscaling:
			autoscaling = &container.NodePoolAutoscalingArgs{
				MinNodeCount: pulumi.Int(int(x.Autoscaling.MinNodes)),
				MaxNodeCount: pulumi.Int(int(x.Autoscaling.MaxNodes)),
				LocationPolicy: pulumi.StringPtr(
					x.Autoscaling.GetLocationPolicy()),
			}
		}
	}

	// Defaults if neither branch set anything (should not happen given proto validation).
	if nodeCount == nil && autoscaling == nil {
		nodeCount = pulumi.IntPtr(1)
	}

	// ----- Management ----------------------------------------------------

	var management *container.NodePoolManagementArgs
	if spec.Management != nil {
		management = &container.NodePoolManagementArgs{
			AutoUpgrade: pulumi.Bool(!spec.Management.DisableAutoUpgrade),
			AutoRepair:  pulumi.Bool(!spec.Management.DisableAutoRepair),
		}
	} else {
		management = &container.NodePoolManagementArgs{
			AutoUpgrade: pulumi.Bool(true),
			AutoRepair:  pulumi.Bool(true),
		}
	}

	// ----- Labels --------------------------------------------------------

	mergedLabels := map[string]string{}
	for k, v := range locals.GcpLabels {
		mergedLabels[k] = v
	}
	for k, v := range spec.NodeLabels {
		mergedLabels[k] = v
	}

	// ----- Node‑config ---------------------------------------------------

	nodeConfig := &container.NodePoolNodeConfigArgs{
		MachineType: pulumi.String(spec.GetMachineType()),
		Preemptible: pulumi.Bool(spec.Spot),
		Labels:      pulumi.ToStringMap(mergedLabels),
		Tags: pulumi.StringArray{
			pulumi.String(locals.NetworkTag),
		},
		Metadata: pulumi.StringMap{
			"disable-legacy-endpoints": pulumi.String("true"),
		},
		OauthScopes: pulumi.StringArray{
			pulumi.String("https://www.googleapis.com/auth/monitoring"),
			pulumi.String("https://www.googleapis.com/auth/logging.write"),
			pulumi.String("https://www.googleapis.com/auth/devstorage.read_only"),
		},
		ImageType: pulumi.StringPtr(spec.GetImageType()),
	}
	if spec.DiskSizeGb > 0 {
		nodeConfig.DiskSizeGb = pulumi.IntPtr(int(spec.DiskSizeGb))
	}
	if spec.DiskType != nil && *spec.DiskType != "" {
		nodeConfig.DiskType = pulumi.StringPtr(*spec.DiskType)
	}
	if spec.ServiceAccount != "" {
		nodeConfig.ServiceAccount = pulumi.StringPtr(spec.ServiceAccount)
	}

	// ----- Create the node‑pool -----------------------------------------

	createdNodePool, err := container.NewNodePool(ctx,
		locals.GcpGkeNodePool.Spec.NodePoolName,
		&container.NodePoolArgs{
			Cluster:     pulumi.String(locals.ClusterName),
			Location:    pulumi.String(locals.ClusterLocation),
			Project:     pulumi.String(locals.GcpGkeNodePool.Spec.ClusterProjectId.GetValue()),
			NodeConfig:  nodeConfig,
			NodeCount:   nodeCount,
			Autoscaling: autoscaling,
			Management:  management,
			UpgradeSettings: container.NodePoolUpgradeSettingsPtrInput(
				&container.NodePoolUpgradeSettingsArgs{
					MaxSurge:       pulumi.Int(2),
					MaxUnavailable: pulumi.Int(1),
				}),
		},
		pulumi.Provider(gcpProvider),
		pulumi.IgnoreChanges([]string{"nodeCount"}),
		pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return errors.Wrap(err, "failed to create node pool")
	}

	// ----- Export outputs ------------------------------------------------

	ctx.Export(OpNodePoolName, createdNodePool.Name)
	ctx.Export(OpNodePoolId, createdNodePool.ID())
	ctx.Export(OpInstanceGroupUrls, createdNodePool.InstanceGroupUrls)
	ctx.Export(OpCurrentNodeCount, createdNodePool.NodeCount)

	// min/max come from either fixed size or autoscaling.
	if nodeCount != nil {
		ctx.Export(OpMinNodes, nodeCount)
		ctx.Export(OpMaxNodes, nodeCount)
	} else if autoscaling != nil {
		ctx.Export(OpMinNodes, autoscaling.MinNodeCount)
		ctx.Export(OpMaxNodes, autoscaling.MaxNodeCount)
	}

	return nil
}
