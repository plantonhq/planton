package module

import (
	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	gcpspannerdatabasev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpspannerdatabase/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds resolved values used across the Pulumi module.
// Note: Spanner databases do not support GCP labels. Labels are
// managed at the instance level only (see GcpSpannerInstance).
type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpSpannerDatabase  *gcpspannerdatabasev1.GcpSpannerDatabase
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpspannerdatabasev1.GcpSpannerDatabaseStackInput) *Locals {
	locals := &Locals{}
	locals.GcpSpannerDatabase = stackInput.Target
	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
