package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output keys for stack outputs
const (
	OutputNamespace              = "namespace"
	OutputNamespaceID            = "namespace_id"
	OutputResourceQuotasApplied  = "resource_quotas_applied"
	OutputLimitRangesApplied     = "limit_ranges_applied"
	OutputNetworkPoliciesApplied = "network_policies_applied"
	OutputServiceMeshEnabled     = "service_mesh_enabled"
	OutputServiceMeshType        = "service_mesh_type"
	OutputPodSecurityStandard    = "pod_security_standard"
	OutputLabelsJSON             = "labels_json"
	OutputAnnotationsJSON        = "annotations_json"
)

// exportOutputs exports all stack outputs
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	// Export namespace name
	ctx.Export(OutputNamespace, pulumi.String(locals.NamespaceName))

	// Export namespace ID (same as name for Kubernetes namespaces)
	ctx.Export(OutputNamespaceID, pulumi.String(locals.NamespaceName))

	// Export resource quotas applied flag
	ctx.Export(OutputResourceQuotasApplied, pulumi.Bool(locals.ResourceQuota.Enabled))

	// Export limit ranges applied flag
	ctx.Export(OutputLimitRangesApplied, pulumi.Bool(locals.LimitRange.Enabled))

	// Export network policies applied flag
	networkPoliciesApplied := locals.NetworkPolicy.IsolateIngress || locals.NetworkPolicy.RestrictEgress
	ctx.Export(OutputNetworkPoliciesApplied, pulumi.Bool(networkPoliciesApplied))

	// Export service mesh enabled flag
	ctx.Export(OutputServiceMeshEnabled, pulumi.Bool(locals.ServiceMesh.Enabled))

	// Export service mesh type
	ctx.Export(OutputServiceMeshType, pulumi.String(locals.ServiceMesh.MeshType))

	// Export pod security standard
	ctx.Export(OutputPodSecurityStandard, pulumi.String(locals.PodSecurityStandard))

	// Export labels as JSON
	labelsJSON := labelsToJSON(locals.Labels)
	ctx.Export(OutputLabelsJSON, pulumi.String(labelsJSON))

	// Export annotations as JSON
	annotationsJSON := annotationsToJSON(locals.Annotations)
	ctx.Export(OutputAnnotationsJSON, pulumi.String(annotationsJSON))

	return nil
}
