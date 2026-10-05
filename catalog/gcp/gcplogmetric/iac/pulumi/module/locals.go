package module

import (
	gcplogmetricv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcplogmetric/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs. The metric resource
// carries no user-labels argument (metric_descriptor.labels is the metric's
// LABEL SCHEMA, not metadata), so there is no platform-label merge here.
type Locals struct {
	GcpLogMetric *gcplogmetricv1alpha1.GcpLogMetric

	// The metric name defaults to metadata.name when the spec leaves
	// metric_name empty — the same naming basis every kind uses.
	MetricName string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcplogmetricv1alpha1.GcpLogMetricIacInput) *Locals {
	target := iacInput.Target

	metricName := target.Spec.MetricName
	if metricName == "" {
		metricName = target.Metadata.Name
	}

	return &Locals{
		GcpLogMetric: target,
		MetricName:   metricName,
	}
}
