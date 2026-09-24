package module

import (
	gcpfirebaseprojectv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirebaseproject/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention. Firebase
// enablement is a project singleton with no name of its own and none of
// its resources carry labels, so the only local is the resolved target.
type Locals struct {
	GcpFirebaseProject *gcpfirebaseprojectv1alpha1.GcpFirebaseProject
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpfirebaseprojectv1alpha1.GcpFirebaseProjectStackInput) *Locals {
	return &Locals{
		GcpFirebaseProject: stackInput.Target,
	}
}
