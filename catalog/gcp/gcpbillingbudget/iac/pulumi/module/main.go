package module

import (
	"github.com/pkg/errors"
	gcpbillingbudgetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbillingbudget/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpbillingbudgetv1alpha1.GcpBillingBudgetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// No project of its own: a budget belongs to a billing account. A person's
	// own gcloud sign-in has no project to bill the Budget API's quota to, so
	// spec.quota_project_id names one; unset, the deploying identity's own
	// project carries the quota and the provider sets no override.
	var gcpProvider *gcp.Provider
	var err error
	if locals.QuotaProjectID != "" {
		gcpProvider, err = pulumigoogleprovider.GetWithQuotaProject(ctx, iacInput.ProviderConfig, "", locals.QuotaProjectID)
	} else {
		gcpProvider, err = pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, "")
	}
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := budget(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create billing budget")
	}
	return nil
}
