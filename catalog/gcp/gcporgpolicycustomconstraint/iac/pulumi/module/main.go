package module

import (
	"github.com/pkg/errors"
	gcporgpolicycustomconstraintv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcporgpolicycustomconstraint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcporgpolicycustomconstraintv1alpha1.GcpOrgPolicyCustomConstraintIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// No project of its own: the constraint belongs to an organization.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, "")
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := customConstraint(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create custom constraint")
	}

	return nil
}
