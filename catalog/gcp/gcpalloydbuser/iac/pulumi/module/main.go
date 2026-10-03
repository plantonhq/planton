package module

import (
	"github.com/pkg/errors"
	gcpalloydbuserv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpalloydbuser/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpalloydbuserv1alpha1.GcpAlloydbUserIacInput) error {
	locals := initializeLocals(iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := user(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create alloydb user")
	}

	return nil
}
