package module

import (
	"github.com/pkg/errors"
	gcporgpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcporgpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcporgpolicyv1alpha1.GcpOrgPolicyIacInput) error {
	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return err
	}

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := orgPolicy(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create organization policy")
	}

	return nil
}
