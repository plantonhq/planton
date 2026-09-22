package module

import (
	"github.com/pkg/errors"
	gcpbillingbudgetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbillingbudget/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpbillingbudgetv1alpha1.GcpBillingBudgetStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := budget(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create billing budget")
	}
	return nil
}
