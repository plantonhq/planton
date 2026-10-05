package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpfirestoredatabasev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirestoredatabase/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds resolved values used across the Pulumi module.
// Note: Firestore databases do not support GCP labels. Labels are
// not available for this resource type.
type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpFirestoreDatabase *gcpfirestoredatabasev1alpha1.GcpFirestoreDatabase
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpfirestoredatabasev1alpha1.GcpFirestoreDatabaseIacInput) *Locals {
	locals := &Locals{}
	locals.GcpFirestoreDatabase = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
