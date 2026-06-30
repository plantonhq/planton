package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewaykapsulepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewaykapsulepool/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout
// the module. The StringValueOrRef for cluster_id is resolved to a plain
// string here -- at IaC runtime, valueFrom references have already been
// resolved by the platform middleware.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayKapsulePool    *scalewaykapsulepoolv1.ScalewayKapsulePool

	// ClusterId is resolved from the required StringValueOrRef field.
	ClusterId string

	// ScalewayTags is the merged tag slice containing:
	//   1. Standard Planton tags (planton-ai_resource, etc.)
	//   2. Kubernetes label tags (noprefix={key}={value})
	//   3. Kubernetes taint tags (taint=noprefix={key}={value}:{Effect})
	ScalewayTags []string
}

// initializeLocals copies stack-input fields into the Locals struct, resolves
// the StringValueOrRef cluster_id, and builds the merged tag slice.
//
// Tags are formatted as flat strings because Scaleway tags are strings (not
// key-value maps). Three categories of tags are merged:
//
//  1. Standard Planton tags: "planton-ai_resource=true", "planton-ai_name=...", etc.
//  2. Kubernetes labels: "noprefix={key}={value}" -- The Scaleway CCM syncs
//     these to K8s node labels as {key}={value}.
//  3. Kubernetes taints: "taint=noprefix={key}={value}:{Effect}" -- The CCM
//     syncs these to K8s node taints as {key}={value}:{Effect}.
//
// We use the "noprefix=" variant so users get exactly the label/taint keys
// they specified, without the default k8s.scaleway.com/ prefix.
func initializeLocals(_ *pulumi.Context, stackInput *scalewaykapsulepoolv1.ScalewayKapsulePoolStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayKapsulePool = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Resolve required Cluster ID from StringValueOrRef.
	if stackInput.Target.Spec.ClusterId != nil {
		locals.ClusterId = stackInput.Target.Spec.ClusterId.GetValue()
	}

	// ── 1. Standard Planton tags ──────────────────────────────────────────
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayKapsulePool.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayKapsulePool.String()),
	}

	if locals.ScalewayKapsulePool.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayKapsulePool.Metadata.Org))
	}

	if locals.ScalewayKapsulePool.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayKapsulePool.Metadata.Env))
	}

	if locals.ScalewayKapsulePool.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayKapsulePool.Metadata.Id))
	}

	// ── 2. Kubernetes label tags ──────────────────────────────────────────
	// Format: "noprefix={key}={value}"
	// CCM syncs to K8s node label: {key}={value}
	for key, value := range stackInput.Target.Spec.KubernetesLabels {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("noprefix=%s=%s", key, value))
	}

	// ── 3. Kubernetes taint tags ──────────────────────────────────────────
	// Format: "taint=noprefix={key}={value}:{Effect}"
	// CCM syncs to K8s node taint: {key}={value}:{Effect}
	for _, taint := range stackInput.Target.Spec.Taints {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("taint=noprefix=%s=%s:%s", taint.Key, taint.Value, taint.Effect))
	}

	return locals
}
