package module

import (
	"github.com/pkg/errors"
	gcpdeploycustomtargettypev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdeploycustomtargettype/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpdeploycustomtargettypev1alpha1.GcpDeployCustomTargetTypeIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := customTargetType(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create the custom target type")
	}

	return nil
}
