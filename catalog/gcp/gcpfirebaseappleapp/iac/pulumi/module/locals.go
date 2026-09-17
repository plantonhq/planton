package module

import (
	gcpfirebaseappleappv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirebaseappleapp/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention. A Firebase
// app registration is identified by its bundle id and none of its
// resources carry labels, so the only local is the resolved target.
type Locals struct {
	GcpFirebaseAppleApp *gcpfirebaseappleappv1alpha1.GcpFirebaseAppleApp
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpfirebaseappleappv1alpha1.GcpFirebaseAppleAppStackInput) *Locals {
	return &Locals{
		GcpFirebaseAppleApp: stackInput.Target,
	}
}
