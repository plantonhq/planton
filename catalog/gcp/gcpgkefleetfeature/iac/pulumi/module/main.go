package module

import (
	"github.com/pkg/errors"
	gcpgkefleetfeaturev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkefleetfeature/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpgkefleetfeaturev1alpha1.GcpGkeFleetFeatureStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := feature(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create the fleet feature")
	}

	return nil
}
