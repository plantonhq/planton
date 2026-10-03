package module

import (
	"fmt"
	"strconv"

	kuberneteshorizontalpodautoscalerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteshorizontalpodautoscaler/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across the module.
type Locals struct {
	Context     *pulumi.Context
	Spec        *kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerSpec
	Namespace   string
	Name        string
	Labels      map[string]string
	Annotations map[string]string

	// The resolved scale target, with the spec's defaults (apps/v1 Deployment)
	// applied module-side so both engines submit identical objects.
	TargetApiVersion string
	TargetKind       string
	TargetName       string

	// The replica floor with the Kubernetes default (1) applied — sent
	// explicitly by both engines.
	MinReplicas int32
}

// initializeLocals extracts and transforms spec fields into module-local values.
func initializeLocals(ctx *pulumi.Context, iacInput *kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to what
	// the Terraform module stamps for the same manifest. User labels merge in
	// afterwards and cannot override the identity keys.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesHorizontalPodAutoscaler.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}
	for k, v := range spec.GetLabels() {
		if _, isIdentityKey := labels[k]; !isIdentityKey {
			labels[k] = v
		}
	}

	annotations := make(map[string]string)
	for k, v := range spec.GetAnnotations() {
		annotations[k] = v
	}

	// namespace is a StringValueOrRef foreign key. References are resolved to
	// literal strings before the module runs, so GetValue() returns the final
	// namespace name. When omitted entirely, fall back to the cluster's
	// "default" namespace — the same behavior as kubectl without a namespace flag.
	namespace := spec.GetNamespace().GetValue()
	if namespace == "" {
		namespace = "default"
	}

	scaleTarget := spec.GetScaleTarget()
	targetApiVersion := scaleTarget.GetApiVersion()
	if targetApiVersion == "" {
		targetApiVersion = "apps/v1"
	}
	targetKind := scaleTarget.GetKind()
	if targetKind == "" {
		targetKind = "Deployment"
	}

	minReplicas := int32(1)
	if spec.MinReplicas != nil {
		minReplicas = spec.GetMinReplicas()
	}

	return &Locals{
		Context:          ctx,
		Spec:             spec,
		Namespace:        namespace,
		Name:             spec.GetName(),
		Labels:           labels,
		Annotations:      annotations,
		TargetApiVersion: targetApiVersion,
		TargetKind:       targetKind,
		TargetName:       scaleTarget.GetName().GetValue(),
		MinReplicas:      minReplicas,
	}
}

// scaleTargetString renders the "Kind/name" outputs handle.
func (l *Locals) scaleTargetString() string {
	return fmt.Sprintf("%s/%s", l.TargetKind, l.TargetName)
}

// metricTargetTypeApiString maps the target-type enum to the API string.
func metricTargetTypeApiString(t kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerMetricTargetType) string {
	switch t {
	case kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerMetricTargetType_utilization:
		return "Utilization"
	case kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerMetricTargetType_raw_value:
		return "Value"
	default:
		return "AverageValue"
	}
}

// selectPolicyApiString maps the select-policy enum to the API string,
// applying the Kubernetes default (Max) when unset.
func selectPolicyApiString(rules *kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerScalingRules) string {
	switch rules.GetSelectPolicy() {
	case kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerScalingRules_min_change:
		return "Min"
	case kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerScalingRules_disabled:
		return "Disabled"
	default:
		return "Max"
	}
}

// scalingPolicyTypeApiString maps the policy-type enum to the API string.
func scalingPolicyTypeApiString(t kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerScalingPolicy_KubernetesHorizontalPodAutoscalerScalingPolicyType) string {
	if t == kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerScalingPolicy_percent {
		return "Percent"
	}
	return "Pods"
}
