package module

import (
	"github.com/pkg/errors"
	gcpvertexaitensorboardv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaitensorboard/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpvertexaitensorboardv1alpha1.GcpVertexAiTensorboardIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.GetInRegion(ctx, iacInput.ProviderConfig,
		iacInput.Target.GetSpec().GetProjectId().GetValue(),
		// Vertex AI's regional endpoint comes from the provider's region, not
		// the resource's location (GetInRegion).
		iacInput.Target.GetSpec().GetLocation())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := tensorboard(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create vertex ai tensorboard")
	}

	return nil
}
