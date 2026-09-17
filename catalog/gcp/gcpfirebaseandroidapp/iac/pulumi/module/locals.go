package module

import (
	gcpfirebaseandroidappv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirebaseandroidapp/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention. A Firebase
// app registration is identified by its package name and none of its
// resources carry labels, so the only local is the resolved target.
type Locals struct {
	GcpFirebaseAndroidApp *gcpfirebaseandroidappv1alpha1.GcpFirebaseAndroidApp
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpfirebaseandroidappv1alpha1.GcpFirebaseAndroidAppStackInput) *Locals {
	return &Locals{
		GcpFirebaseAndroidApp: stackInput.Target,
	}
}
