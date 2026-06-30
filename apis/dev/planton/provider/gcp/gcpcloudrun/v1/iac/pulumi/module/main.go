package module

import (
	"github.com/pkg/errors"
	gcpcloudrunv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudrun/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpCloudRun component.
func Resources(ctx *pulumi.Context, stackInput *gcpcloudrunv1.GcpCloudRunStackInput) error {
	locals := initializeLocals(stackInput)

	// Set up the GCP provider from the supplied credential.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// Create the Cloud Run service.
	createdService, err := service(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloud-run service")
	}

	ctx.Export(OpUrl, createdService.Uri)
	ctx.Export(OpServiceName, createdService.Name)
	ctx.Export(OpRevision, createdService.LatestReadyRevision)

	if locals.GcpCloudRun.Spec.Dns != nil && locals.GcpCloudRun.Spec.Dns.Enabled {
		if err := customDns(ctx, locals, createdService, gcpProvider); err != nil {
			return errors.Wrap(err, "failed to create custom dns mapping resources")
		}
	}

	return nil
}
