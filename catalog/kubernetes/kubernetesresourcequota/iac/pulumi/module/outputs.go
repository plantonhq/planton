package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesResourceQuotaOutputs field names.
const (
	OutputResourceQuotaName = "resource_quota_name"
	OutputNamespace         = "namespace"
	OutputLimitRangeName    = "limit_range_name"
)

// exportOutputs exports the outputs from the created governance pair.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputResourceQuotaName, pulumi.String(locals.Name))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))
	// Empty when no limit_defaults were configured (no LimitRange exists).
	ctx.Export(OutputLimitRangeName, pulumi.String(locals.LimitRangeName))

	return nil
}
