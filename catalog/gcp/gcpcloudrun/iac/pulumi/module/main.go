package module

import (
	"github.com/pkg/errors"
	gcpcloudrunv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrun/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry-point for the GcpCloudRun kind.
func Resources(ctx *pulumi.Context, iacInput *gcpcloudrunv1alpha1.GcpCloudRunIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Set up the GCP provider from the supplied credential.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	// Create the Cloud Run service (and its public-invoker grant when the
	// spec asks for one).
	createdService, err := service(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloud-run service")
	}

	ctx.Export(OpUrl, createdService.Uri)
	ctx.Export(OpServiceName, createdService.Name)
	ctx.Export(OpRevision, createdService.LatestReadyRevision)
	ctx.Export(OpLocation, createdService.Location)
	// The project is part of the service's provider-side identity (the
	// projects/{project}/locations/{location}/services/{name} coordinate);
	// exporting it lets readers of the record locate the service on the
	// provider without re-deriving where the deploy resolved its project.
	ctx.Export(OpProjectId, createdService.Project)
	ctx.Export(OpUid, createdService.Uid)
	ctx.Export(OpUrls, createdService.Urls)

	return nil
}
