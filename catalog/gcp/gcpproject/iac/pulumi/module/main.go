package module

import (
	"github.com/pkg/errors"
	gcpprojectv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpprojectv1alpha1.GcpProjectIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	createdProject, err := project(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create GCP project")
	}

	if err := apis(ctx, locals, createdProject, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to enable apis for GCP project")
	}

	ctx.Export(OpProjectId, createdProject.ProjectId)
	ctx.Export(OpProjectNumber, createdProject.Number)
	ctx.Export(OpProjectName, createdProject.Name)

	return nil
}
