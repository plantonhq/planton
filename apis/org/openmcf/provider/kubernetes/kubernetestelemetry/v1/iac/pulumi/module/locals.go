package module

import (
	kubernetestelemetryv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetestelemetry/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesTelemetry *kubernetestelemetryv1.KubernetesTelemetry
	// TelemetryName is the Kubernetes resource name of the Telemetry resource. It
	// equals the OpenMCF resource metadata.name.
	TelemetryName string
	// Namespace is the resolved namespace the Telemetry resource is created in.
	Namespace string
	Labels    map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetestelemetryv1.KubernetesTelemetryStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	telemetryName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware resolves
	// valueFrom references to literal strings before the IaC module runs, so
	// GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "telemetry",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "openmcf",
		"app.kubernetes.io/component":  "telemetry",
	}

	return &Locals{
		KubernetesTelemetry: target,
		TelemetryName:       telemetryName,
		Namespace:           namespace,
		Labels:              labels,
	}
}
