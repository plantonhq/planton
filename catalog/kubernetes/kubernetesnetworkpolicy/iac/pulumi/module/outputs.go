package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesNetworkPolicyOutputs field names.
const (
	OutputNetworkPolicyName = "network_policy_name"
	OutputNamespace         = "namespace"
	OutputPolicyTypes       = "policy_types"
)

// exportOutputs exports the outputs from the created NetworkPolicy.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputNetworkPolicyName, pulumi.String(locals.Name))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))
	// Includes inferred directions when the spec omitted policy_types, so the
	// output always states the deployed truth.
	ctx.Export(OutputPolicyTypes, pulumi.String(policyTypesString(locals.PolicyTypes)))

	return nil
}
