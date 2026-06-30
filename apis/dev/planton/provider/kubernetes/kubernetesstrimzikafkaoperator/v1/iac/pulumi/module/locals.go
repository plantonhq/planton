package module

import (
	kubernetesstrimzikafkaoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesstrimzikafkaoperator/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// locals computes derived configuration values from the stack input
type locals struct {
	namespace    string
	labels       pulumi.StringMap
	operatorName string
	chartVersion string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	HelmReleaseName string
}

// newLocals creates computed values from stack input
func newLocals(stackInput *kubernetesstrimzikafkaoperatorv1.KubernetesStrimziKafkaOperatorStackInput) *locals {
	operatorName := "strimzi-kafka-operator"
	if stackInput.Target != nil && stackInput.Target.Metadata != nil && stackInput.Target.Metadata.Name != "" {
		operatorName = stackInput.Target.Metadata.Name
	}

	labels := pulumi.StringMap{
		"app.kubernetes.io/name":       pulumi.String("strimzi-kafka-operator"),
		"app.kubernetes.io/managed-by": pulumi.String("planton"),
		"planton.ai/resource-kind":     pulumi.String("kubernetes-strimzi-kafka-operator"),
	}

	if stackInput.Target != nil && stackInput.Target.Metadata != nil {
		if stackInput.Target.Metadata.Name != "" {
			labels["planton.ai/resource-id"] = pulumi.String(stackInput.Target.Metadata.Name)
		}
		if stackInput.Target.Metadata.Org != "" {
			labels["planton.ai/organization"] = pulumi.String(stackInput.Target.Metadata.Org)
		}
		if stackInput.Target.Metadata.Env != "" {
			labels["planton.ai/environment"] = pulumi.String(stackInput.Target.Metadata.Env)
		}
	}

	// get namespace from spec
	namespace := vars.Namespace
	if stackInput.Target != nil && stackInput.Target.Spec != nil && stackInput.Target.Spec.Namespace != nil {
		namespace = stackInput.Target.Spec.Namespace.GetValue()
	}

	return &locals{
		namespace:    namespace,
		labels:       labels,
		operatorName: operatorName,
		chartVersion: vars.HelmChartVersion,

		// Computed resource names to avoid conflicts when multiple instances share a namespace
		// Format: {metadata.name} - users can prefix with component type if needed (e.g., "strimzi-prod")
		HelmReleaseName: operatorName,
	}
}
