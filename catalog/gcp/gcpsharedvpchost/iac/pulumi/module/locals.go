package module

import (
	gcpsharedvpchostv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsharedvpchost/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target. The host resource carries no name of its own and no labels (it
// is a flag on a project), so there is nothing to derive here beyond the
// project, which is resolved at create time because it may come from the
// provider's configuration rather than the spec.
type Locals struct {
	GcpSharedVpcHost *gcpsharedvpchostv1alpha1.GcpSharedVpcHost
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpsharedvpchostv1alpha1.GcpSharedVpcHostStackInput) *Locals {
	return &Locals{
		GcpSharedVpcHost: stackInput.Target,
	}
}
