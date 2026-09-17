package module

import (
	"github.com/pkg/errors"
	gcpapikeyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpapikey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpapikeyv1alpha1.GcpApiKeyStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// The API Keys API attributes quota to the caller's project on
	// user-credential calls: under plain ADC (`gcloud auth
	// application-default login`) a create without user_project_override
	// fails with 403 "requires a quota project" -- the same behavior the
	// Identity Toolkit API shows, and the same fix. The override attributes
	// quota to the key's own project under every credential mode.
	gcpProvider, err := pulumigoogleprovider.GetWithUserProjectOverride(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := apiKey(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create api key")
	}

	return nil
}
