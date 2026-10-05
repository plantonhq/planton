package module

import (
	"github.com/pkg/errors"
	gcpcloudcomposeruserworkloadssecretv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudcomposeruserworkloadssecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpcloudcomposeruserworkloadssecretv1alpha1.GcpCloudComposerUserWorkloadsSecretIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := userWorkloadsSecret(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create composer user workloads secret")
	}

	return nil
}
