package module

import (
	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	gcpfirestoredatabasev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpfirestoredatabase/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds resolved values used across the Pulumi module.
// Note: Firestore databases do not support GCP labels. Labels are
// not available for this resource type.
type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpFirestoreDatabase *gcpfirestoredatabasev1.GcpFirestoreDatabase
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpfirestoredatabasev1.GcpFirestoreDatabaseStackInput) *Locals {
	locals := &Locals{}
	locals.GcpFirestoreDatabase = stackInput.Target
	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
