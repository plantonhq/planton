package module

import (
	"strconv"

	"github.com/pkg/errors"
	autoscalingv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/autoscaling/v2"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	policyv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/policy/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// hpa creates the HorizontalPodAutoscaler when enabled. Terraform twin:
// kubernetes_horizontal_pod_autoscaler_v2.flagd with count.
func hpa(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, deps []pulumi.ResourceOption) error {
	h := locals.Spec.GetHpa()
	if !h.GetEnabled() {
		return nil
	}
	minReplicas, maxReplicas, cpu := 1, 10, 80
	if h.MinReplicas != nil {
		minReplicas = int(h.GetMinReplicas())
	}
	if h.MaxReplicas != nil {
		maxReplicas = int(h.GetMaxReplicas())
	}
	if h.TargetCpuUtilizationPercent != nil {
		cpu = int(h.GetTargetCpuUtilizationPercent())
	}
	metrics := autoscalingv2.MetricSpecArray{
		autoscalingv2.MetricSpecArgs{Type: pulumi.String("Resource"), Resource: autoscalingv2.ResourceMetricSourceArgs{
			Name: pulumi.String("cpu"), Target: autoscalingv2.MetricTargetArgs{Type: pulumi.String("Utilization"), AverageUtilization: pulumi.Int(cpu)},
		}},
	}
	if h.TargetMemoryUtilizationPercent != nil {
		metrics = append(metrics, autoscalingv2.MetricSpecArgs{Type: pulumi.String("Resource"), Resource: autoscalingv2.ResourceMetricSourceArgs{
			Name: pulumi.String("memory"), Target: autoscalingv2.MetricTargetArgs{Type: pulumi.String("Utilization"), AverageUtilization: pulumi.Int(int(h.GetTargetMemoryUtilizationPercent()))},
		}})
	}
	_, err := autoscalingv2.NewHorizontalPodAutoscaler(ctx, locals.Name, &autoscalingv2.HorizontalPodAutoscalerArgs{
		Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
			Name: pulumi.String(locals.Name), Namespace: pulumi.String(locals.Namespace), Labels: pulumi.ToStringMap(locals.Labels),
		}),
		Spec: autoscalingv2.HorizontalPodAutoscalerSpecArgs{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReferenceArgs{
				ApiVersion: pulumi.String("apps/v1"), Kind: pulumi.String("Deployment"), Name: pulumi.String(locals.Name),
			},
			MinReplicas: pulumi.Int(minReplicas),
			MaxReplicas: pulumi.Int(maxReplicas),
			Metrics:     metrics,
		},
	}, append([]pulumi.ResourceOption{pulumi.Provider(provider)}, deps...)...)
	if err != nil {
		return errors.Wrap(err, "failed to create flagd autoscaler")
	}
	return nil
}

// pdb creates the PodDisruptionBudget when enabled (minAvailable 1 when no
// bound is given). Terraform twin: kubernetes_pod_disruption_budget_v1.flagd.
func pdb(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, deps []pulumi.ResourceOption) error {
	p := locals.Spec.GetPdb()
	if !p.GetEnabled() {
		return nil
	}
	spec := &policyv1.PodDisruptionBudgetSpecArgs{
		Selector: kubernetesmeta.LabelSelectorArgs{MatchLabels: pulumi.ToStringMap(locals.SelectorLabels)},
	}
	switch {
	case p.GetMaxUnavailable() != "":
		spec.MaxUnavailable = intOrStringInput(p.GetMaxUnavailable())
	case p.GetMinAvailable() != "":
		spec.MinAvailable = intOrStringInput(p.GetMinAvailable())
	default:
		spec.MinAvailable = pulumi.Int(1)
	}
	_, err := policyv1.NewPodDisruptionBudget(ctx, locals.Name, &policyv1.PodDisruptionBudgetArgs{
		Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
			Name: pulumi.String(locals.Name), Namespace: pulumi.String(locals.Namespace), Labels: pulumi.ToStringMap(locals.Labels),
		}),
		Spec: spec,
	}, append([]pulumi.ResourceOption{pulumi.Provider(provider)}, deps...)...)
	if err != nil {
		return errors.Wrap(err, "failed to create flagd disruption budget")
	}
	return nil
}

// intOrStringInput passes an integer bound as a number and a percentage as
// a string (the Kubernetes IntOrString convention).
func intOrStringInput(s string) pulumi.Input {
	if n, err := strconv.Atoi(s); err == nil {
		return pulumi.Int(n)
	}
	return pulumi.String(s)
}
