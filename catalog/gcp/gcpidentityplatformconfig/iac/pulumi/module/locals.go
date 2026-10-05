package module

import (
	gcpidentityplatformconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpidentityplatformconfig/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention. The config
// is a project singleton with no name of its own and none of its
// resources carry labels, so the only local is the resolved target.
type Locals struct {
	GcpIdentityPlatformConfig *gcpidentityplatformconfigv1alpha1.GcpIdentityPlatformConfig
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpidentityplatformconfigv1alpha1.GcpIdentityPlatformConfigIacInput) *Locals {
	return &Locals{
		GcpIdentityPlatformConfig: iacInput.Target,
	}
}
