package module

import (
	"github.com/pkg/errors"
	gcpmodelarmortemplatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmodelarmortemplate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpmodelarmortemplatev1alpha1.GcpModelArmorTemplateStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := template(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create model armor template")
	}

	return nil
}
