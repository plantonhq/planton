package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdataprocautoscalingpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdataprocautoscalingpolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig            *gcpprovider.GcpProviderConfig
	GcpDataprocAutoscalingPolicy *gcpdataprocautoscalingpolicyv1alpha1.GcpDataprocAutoscalingPolicy
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpdataprocautoscalingpolicyv1alpha1.GcpDataprocAutoscalingPolicyIacInput) *Locals {
	locals := &Locals{}
	locals.GcpDataprocAutoscalingPolicy = iacInput.Target

	// The autoscaling-policy resource has no labels surface in the
	// Dataproc API — no platform attribution labels are stamped,
	// identically on both engines.

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
