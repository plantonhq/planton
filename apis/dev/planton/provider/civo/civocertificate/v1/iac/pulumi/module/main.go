package module

import (
	civocertificatev1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civocertificate/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in other Planton modules.
// NOTE: As of 2025, the Civo Pulumi/Terraform provider does not expose certificate resources.
// This module provides validation and structure but cannot provision actual certificates.
// See: https://registry.terraform.io/providers/civo/civo/latest/docs
func Resources(
	ctx *pulumi.Context,
	stackInput *civocertificatev1.CivoCertificateStackInput,
) error {
	// 1. Prepare locals (metadata, labels, configuration)
	locals := initializeLocals(ctx, stackInput)

	// 2. Attempt certificate provisioning (currently logs warning about provider limitation)
	if err := certificate(ctx, locals); err != nil {
		return err
	}

	return nil
}
