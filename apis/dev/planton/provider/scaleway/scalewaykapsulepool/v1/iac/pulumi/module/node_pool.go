package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/kubernetes"
)

// nodePool provisions the Kubernetes node pool and exports its outputs.
//
// All Kubernetes labels and taints have already been encoded as CCM-formatted
// tags in locals.ScalewayTags by initializeLocals(). This function passes
// the merged tag array directly to the pool resource.
func nodePool(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) error {
	spec := locals.ScalewayKapsulePool.Spec
	name := locals.ScalewayKapsulePool.Metadata.Name

	// ── Build pool arguments ──────────────────────────────────────────────

	poolArgs := &kubernetes.PoolArgs{
		ClusterId: pulumi.String(locals.ClusterId),
		Name:      pulumi.String(name),
		NodeType:  pulumi.String(spec.NodeType),
		Size:      pulumi.Int(int(spec.Size)),
		Tags:      pulumi.ToStringArray(locals.ScalewayTags),
		Region:    pulumi.String(spec.Region),
	}

	// Autoscaling configuration.
	if spec.AutoScale {
		poolArgs.Autoscaling = pulumi.Bool(true)
		if spec.MinSize > 0 {
			poolArgs.MinSize = pulumi.Int(int(spec.MinSize))
		}
		if spec.MaxSize > 0 {
			poolArgs.MaxSize = pulumi.Int(int(spec.MaxSize))
		}
	}

	// Autohealing.
	if spec.Autohealing {
		poolArgs.Autohealing = pulumi.Bool(true)
	}

	// Container runtime.
	if spec.ContainerRuntime != "" {
		poolArgs.ContainerRuntime = pulumi.String(spec.ContainerRuntime)
	}

	// Root volume configuration.
	if spec.RootVolumeType != "" {
		poolArgs.RootVolumeType = pulumi.String(spec.RootVolumeType)
	}
	if spec.RootVolumeSizeInGb > 0 {
		poolArgs.RootVolumeSizeInGb = pulumi.Int(int(spec.RootVolumeSizeInGb))
	}

	// Disable public IPs on nodes.
	if spec.PublicIpDisabled {
		poolArgs.PublicIpDisabled = pulumi.Bool(true)
	}

	// Zone (optional, for zone-specific placement within region).
	if spec.Zone != "" {
		poolArgs.Zone = pulumi.String(spec.Zone)
	}

	// Placement group (optional, for anti-affinity scheduling).
	if spec.PlacementGroupId != "" {
		poolArgs.PlacementGroupId = pulumi.String(spec.PlacementGroupId)
	}

	// Kubelet arguments (optional, power-user escape hatch).
	if len(spec.KubeletArgs) > 0 {
		kubeletArgs := pulumi.StringMap{}
		for k, v := range spec.KubeletArgs {
			kubeletArgs[k] = pulumi.String(v)
		}
		poolArgs.KubeletArgs = kubeletArgs
	}

	// Upgrade policy (optional).
	if spec.UpgradePolicy != nil {
		poolArgs.UpgradePolicy = &kubernetes.PoolUpgradePolicyArgs{
			MaxSurge:       pulumi.Int(int(spec.UpgradePolicy.MaxSurge)),
			MaxUnavailable: pulumi.Int(int(spec.UpgradePolicy.MaxUnavailable)),
		}
	}

	// Wait for pool to be ready before marking complete.
	poolArgs.WaitForPoolReady = pulumi.Bool(true)

	// ── Create the node pool ──────────────────────────────────────────────

	createdPool, err := kubernetes.NewPool(
		ctx,
		"pool",
		poolArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create kapsule node pool")
	}

	// ── Export stack outputs ──────────────────────────────────────────────

	ctx.Export(OpPoolId, createdPool.ID())
	ctx.Export(OpPoolVersion, createdPool.Version)
	ctx.Export(OpCurrentSize, createdPool.CurrentSize)

	return nil
}
