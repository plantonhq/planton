package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nodePool provisions the Kubernetes node pool and exports the stack
// outputs declared in outputs.proto.
func nodePool(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.KubernetesNodePool, error) {
	spec := locals.DigitalOceanKubernetesNodePool.Spec

	// Kubernetes node labels: user labels over the standard Planton labels
	// (identical map in both provisioners).
	labels := pulumi.StringMap{}
	for k, v := range locals.DigitalOceanLabels {
		labels[k] = pulumi.String(v)
	}
	for k, v := range spec.Labels {
		labels[k] = pulumi.String(v)
	}

	// User tags plus the standard Planton labels rendered as "key:value"
	// tags — the exact set the Terraform module applies.
	tagSet := map[string]bool{}
	var tags pulumi.StringArray
	for _, t := range spec.Tags {
		if !tagSet[t] {
			tagSet[t] = true
			tags = append(tags, pulumi.String(t))
		}
	}
	for k, v := range locals.DigitalOceanLabels {
		t := k + ":" + v
		if !tagSet[t] {
			tagSet[t] = true
			tags = append(tags, pulumi.String(t))
		}
	}

	var taints digitalocean.KubernetesNodePoolTaintArray
	for _, taint := range spec.Taints {
		taints = append(taints, &digitalocean.KubernetesNodePoolTaintArgs{
			Key: pulumi.String(taint.Key),
			// Kubernetes allows valueless taints; the provider requires the
			// value be sent, possibly empty.
			Value:  pulumi.String(taint.Value),
			Effect: pulumi.String(taint.Effect),
		})
	}

	nodePoolArgs := &digitalocean.KubernetesNodePoolArgs{
		// The FK resolves to the owning DOKS cluster's UUID.
		ClusterId: pulumi.String(spec.Cluster.GetValue()),
		Name:      pulumi.String(spec.NodePoolName),
		Size:      pulumi.String(spec.Size),
		// With auto_scale enabled this is the initial count; the provider
		// then suppresses diffs while the live count drifts.
		NodeCount: pulumi.IntPtr(int(spec.NodeCount)),
		Labels:    labels,
		Tags:      tags,
	}

	if len(taints) > 0 {
		nodePoolArgs.Taints = taints
	}

	// GPU partitioning is create-only and only meaningful on GPU sizes; unset
	// must arrive as null, never "" (the provider rejects it).
	if spec.GpuPartitionMode != "" {
		nodePoolArgs.GpuPartitionMode = pulumi.StringPtr(spec.GpuPartitionMode)
	}

	if spec.AutoScale {
		nodePoolArgs.AutoScale = pulumi.BoolPtr(true)
		nodePoolArgs.MinNodes = pulumi.IntPtr(int(spec.MinNodes))
		nodePoolArgs.MaxNodes = pulumi.IntPtr(int(spec.MaxNodes))
	}

	createdNodePool, err := digitalocean.NewKubernetesNodePool(
		ctx,
		"node_pool",
		nodePoolArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean kubernetes node pool")
	}

	ctx.Export(OpNodePoolId, createdNodePool.ID())
	ctx.Export(OpClusterId, createdNodePool.ClusterId)

	// The pool's nodes (Nodes[*].Id, Nodes[*].DropletId) are deliberately
	// not exported: DOKS replaces them by design (autoscaling, upgrades,
	// auto-repair), so an apply-time list is stale the next time the pool
	// changes shape. Droplet-scoped resources target the pool's tags instead
	// -- the same contract the Terraform module exports.

	return createdNodePool, nil
}
