package module

import (
	"github.com/pkg/errors"
	gcpspannerbackupschedulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpspannerbackupschedule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpspannerbackupschedulev1alpha1.GcpSpannerBackupScheduleIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := spannerBackupSchedule(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create spanner backup schedule")
	}

	return nil
}
