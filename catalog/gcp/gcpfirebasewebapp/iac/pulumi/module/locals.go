package module

import (
	gcpfirebasewebappv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirebasewebapp/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention. A Firebase
// web app is identified only by its display name and none of its resources
// carry labels, so the only local is the resolved target.
type Locals struct {
	GcpFirebaseWebApp *gcpfirebasewebappv1alpha1.GcpFirebaseWebApp
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpfirebasewebappv1alpha1.GcpFirebaseWebAppStackInput) *Locals {
	return &Locals{
		GcpFirebaseWebApp: stackInput.Target,
	}
}
