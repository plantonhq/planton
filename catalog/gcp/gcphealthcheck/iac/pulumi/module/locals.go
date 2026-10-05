package module

import (
	gcphealthcheckv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphealthcheck/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpHealthCheck *gcphealthcheckv1alpha1.GcpHealthCheck

	// The cloud-side name defaults to metadata.name when the spec leaves
	// health_check_name empty — the same naming basis every kind uses.
	HealthCheckName string
}

func initializeLocals(ctx *pulumi.Context, iacInput *gcphealthcheckv1alpha1.GcpHealthCheckIacInput) *Locals {
	target := iacInput.Target

	healthCheckName := target.Spec.HealthCheckName
	if healthCheckName == "" {
		healthCheckName = target.Metadata.Name
	}

	return &Locals{
		GcpHealthCheck:  target,
		HealthCheckName: healthCheckName,
	}
}
