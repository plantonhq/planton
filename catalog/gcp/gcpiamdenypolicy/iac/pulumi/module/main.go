package module

import (
	"github.com/pkg/errors"
	gcpiamdenypolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpiamdenypolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpiamdenypolicyv1alpha1.GcpIamDenyPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := denyPolicy(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create deny policy")
	}

	return nil
}
