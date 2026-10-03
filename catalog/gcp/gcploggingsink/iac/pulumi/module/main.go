package module

import (
	"github.com/pkg/errors"
	gcploggingsinkv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcploggingsink/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcploggingsinkv1alpha1.GcpLoggingSinkIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := loggingSink(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create logging sink")
	}

	return nil
}
