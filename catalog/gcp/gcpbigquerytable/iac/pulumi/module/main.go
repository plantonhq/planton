package module

import (
	"github.com/pkg/errors"
	gcpbigquerytablev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigquerytable/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpbigquerytablev1alpha1.GcpBigQueryTableIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := table(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create bigquery table")
	}

	return nil
}
