package module

import (
	"github.com/pkg/errors"
	gcpcomputeimagev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcomputeimage/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpcomputeimagev1alpha1.GcpComputeImageIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := image(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create the compute image")
	}

	return nil
}
