package module

import (
	"github.com/pkg/errors"
	gcpcolabruntimev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcolabruntime/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpcolabruntimev1alpha1.GcpColabRuntimeIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := runtime(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create colab runtime")
	}

	return nil
}
