package module

import (
	"github.com/pkg/errors"
	gcpidentityplatformconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpidentityplatformconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpidentityplatformconfigv1alpha1.GcpIdentityPlatformConfigIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// The Identity Toolkit API requires a quota project on user-credential
	// calls: without user_project_override, a deploy under plain ADC
	// (`gcloud auth application-default login`) fails at create with 403
	// "requires a quota project" (live-verified). The override attributes
	// quota to the resource's own project under every credential mode.
	gcpProvider, err := pulumigoogleprovider.GetWithUserProjectOverride(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := identityPlatformConfig(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create identity platform config")
	}

	return nil
}
