package module

import (
	gcpserviceaccountv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpserviceaccount/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the resource structure for the GcpServiceAccount component
// as well as any auxiliary fields we might need in the Pulumi module.
type Locals struct {
	GcpServiceAccount *gcpserviceaccountv1.GcpServiceAccount
}

// initializeLocals creates and returns a Locals struct.
func initializeLocals(ctx *pulumi.Context, stackInput *gcpserviceaccountv1.GcpServiceAccountStackInput) *Locals {
	locals := &Locals{
		GcpServiceAccount: stackInput.Target,
	}
	return locals
}
