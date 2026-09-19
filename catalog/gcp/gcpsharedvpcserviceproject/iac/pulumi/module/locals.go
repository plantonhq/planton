package module

import (
	gcpsharedvpcserviceprojectv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsharedvpcserviceproject/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target. Both projects arrive as resolved references and the attachment
// carries no name or labels of its own, so nothing is derived.
type Locals struct {
	GcpSharedVpcServiceProject *gcpsharedvpcserviceprojectv1alpha1.GcpSharedVpcServiceProject
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpsharedvpcserviceprojectv1alpha1.GcpSharedVpcServiceProjectStackInput) *Locals {
	return &Locals{
		GcpSharedVpcServiceProject: stackInput.Target,
	}
}
