package module

import (
	"github.com/pkg/errors"
	gcpcloudrunjobv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrunjob/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpCloudRunJob kind.
func Resources(ctx *pulumi.Context, iacInput *gcpcloudrunjobv1alpha1.GcpCloudRunJobIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	createdJob, err := job(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloud-run job")
	}

	ctx.Export(OpJobName, createdJob.Name)
	ctx.Export(OpLocation, createdJob.Location)
	ctx.Export(OpUid, createdJob.Uid)
	ctx.Export(OpLatestCreatedExecution, createdJob.LatestCreatedExecutions.ApplyT(func(v []cloudrunv2.JobLatestCreatedExecution) string {
		if len(v) == 0 || v[0].Name == nil {
			return ""
		}
		return *v[0].Name
	}))

	return nil
}
