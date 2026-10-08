package module

import (
	"github.com/pkg/errors"
	gcpcloudbuildrepositoryv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudbuildrepository/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpcloudbuildrepositoryv1alpha1.GcpCloudBuildRepositoryIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// The repository lives in its connection's project (see parseConnectionName).
	project, _, err := parseConnectionName(iacInput.Target.GetSpec().GetParentConnection().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to read the connection's project")
	}

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, project)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := repository(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create the repository link")
	}

	return nil
}
