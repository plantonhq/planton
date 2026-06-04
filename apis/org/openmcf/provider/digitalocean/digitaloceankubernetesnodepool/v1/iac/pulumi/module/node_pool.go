package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nodePool provisions the Kubernetes node‑pool and exports its IDs.
func nodePool(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.KubernetesNodePool, error) {

	// 1. Build labels - merge metadata labels with spec labels
	labels := pulumi.StringMap{}
	// First add metadata-derived labels
	for k, v := range locals.DigitalOceanLabels {
		labels[k] = pulumi.String(v)
	}
	// Then add user-specified labels from spec (can override metadata labels)
	for k, v := range locals.DigitalOceanKubernetesNodePool.Spec.Labels {
		labels[k] = pulumi.String(v)
	}

	// 2. Build taints array
	var taints digitalocean.KubernetesNodePoolTaintArray
	for _, taint := range locals.DigitalOceanKubernetesNodePool.Spec.Taints {
		taints = append(taints, &digitalocean.KubernetesNodePoolTaintArgs{
			Key:    pulumi.String(taint.Key),
			Value:  pulumi.String(taint.Value),
			Effect: pulumi.String(taint.Effect),
		})
	}

	// 3. Build tags array
	tags := pulumi.StringArray{}
	for _, t := range locals.DigitalOceanKubernetesNodePool.Spec.Tags {
		tags = append(tags, pulumi.String(t))
	}

	// 4. Build node pool arguments
	nodePoolArgs := &digitalocean.KubernetesNodePoolArgs{
		ClusterId: pulumi.String(locals.DigitalOceanKubernetesNodePool.Spec.Cluster.GetValue()),
		Name:      pulumi.String(locals.DigitalOceanKubernetesNodePool.Spec.NodePoolName),
		Size:      pulumi.String(locals.DigitalOceanKubernetesNodePool.Spec.Size),
		Labels:    labels,
		Tags:      tags,
		NodeCount: pulumi.IntPtr(int(locals.DigitalOceanKubernetesNodePool.Spec.NodeCount)),
	}

	// Add taints if specified
	if len(taints) > 0 {
		nodePoolArgs.Taints = taints
	}

	// 5. Configure autoscaling if enabled
	if locals.DigitalOceanKubernetesNodePool.Spec.AutoScale {
		nodePoolArgs.AutoScale = pulumi.BoolPtr(true)
		nodePoolArgs.MinNodes = pulumi.IntPtr(int(locals.DigitalOceanKubernetesNodePool.Spec.MinNodes))
		nodePoolArgs.MaxNodes = pulumi.IntPtr(int(locals.DigitalOceanKubernetesNodePool.Spec.MaxNodes))
	}

	// 6. Create the node‑pool.
	createdNodePool, err := digitalocean.NewKubernetesNodePool(
		ctx,
		"node_pool",
		nodePoolArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean kubernetes node pool")
	}

	// 7. Export stack outputs.
	ctx.Export(OpNodePoolId, createdNodePool.ID())

	// node_ids: the Droplet IDs backing the pool's nodes. The provider populates
	// KubernetesNodePool.Nodes once the pool's Droplets exist; project each
	// node's droplet_id into a string slice for the repeated proto output.
	nodeDropletIds := createdNodePool.Nodes.ApplyT(
		func(nodes []digitalocean.KubernetesNodePoolNode) []string {
			ids := make([]string, 0, len(nodes))
			for _, node := range nodes {
				if node.DropletId != nil {
					ids = append(ids, *node.DropletId)
				}
			}
			return ids
		},
	).(pulumi.StringArrayOutput)
	ctx.Export(OpNodeIds, nodeDropletIds)

	return createdNodePool, nil
}
