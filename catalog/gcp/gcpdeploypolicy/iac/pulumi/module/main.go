package module

import (
	"github.com/pkg/errors"
	gcpdeploypolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdeploypolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpdeploypolicyv1alpha1.GcpDeployPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := deployPolicy(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create the deploy policy")
	}

	return nil
}
