package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesHorizontalPodAutoscalerOutputs field names.
const (
	OutputHorizontalPodAutoscalerName = "horizontal_pod_autoscaler_name"
	OutputNamespace                   = "namespace"
	OutputScaleTarget                 = "scale_target"
	OutputMinReplicas                 = "min_replicas"
	OutputMaxReplicas                 = "max_replicas"
)

// exportOutputs exports the outputs from the created HorizontalPodAutoscaler.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputHorizontalPodAutoscalerName, pulumi.String(locals.Name))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OutputScaleTarget, pulumi.String(locals.scaleTargetString()))
	ctx.Export(OutputMinReplicas, pulumi.Int(int(locals.MinReplicas)))
	ctx.Export(OutputMaxReplicas, pulumi.Int(int(locals.Spec.GetMaxReplicas())))

	return nil
}
