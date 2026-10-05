package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpspannerdatabasev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpspannerdatabase/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds resolved values used across the Pulumi module.
// Note: Spanner databases do not support GCP labels. Labels are
// managed at the instance level only (see GcpSpannerInstance).
type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpSpannerDatabase *gcpspannerdatabasev1alpha1.GcpSpannerDatabase
	DatabaseName       string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpspannerdatabasev1alpha1.GcpSpannerDatabaseIacInput) *Locals {
	locals := &Locals{}
	locals.GcpSpannerDatabase = iacInput.Target

	locals.DatabaseName = locals.GcpSpannerDatabase.Spec.DatabaseName
	if locals.DatabaseName == "" {
		locals.DatabaseName = locals.GcpSpannerDatabase.Metadata.Name
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
