package module

import (
	gcpfirestoreindexv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirestoreindex/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds resolved values used across the Pulumi module.
// Firestore indexes do not support GCP labels — skip label merge.
type Locals struct {
	GcpFirestoreIndex *gcpfirestoreindexv1alpha1.GcpFirestoreIndex
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpfirestoreindexv1alpha1.GcpFirestoreIndexIacInput) *Locals {
	return &Locals{
		GcpFirestoreIndex: iacInput.Target,
	}
}
