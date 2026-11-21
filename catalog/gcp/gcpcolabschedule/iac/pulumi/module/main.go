package module

import (
	"github.com/pkg/errors"
	gcpcolabschedulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcolabschedule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpcolabschedulev1alpha1.GcpColabScheduleIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.GetInRegion(ctx, iacInput.ProviderConfig,
		iacInput.Target.GetSpec().GetProjectId().GetValue(),
		// The provider fills each run request's parent from its own region
		// before every create and update, and fails with none; the region
		// is the schedule's location (GetInRegion).
		iacInput.Target.GetSpec().GetLocation())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := schedule(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create colab schedule")
	}

	return nil
}
