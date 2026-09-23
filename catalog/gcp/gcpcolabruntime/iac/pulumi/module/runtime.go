package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/colab"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// runtime creates the Colab Enterprise runtime, assigned to runtime_user
// from the template. desired_state and auto_upgrade are client-side
// controls the provider enforces on every apply; everything else is fixed
// at creation.
func runtime(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpColabRuntime.Spec
	metadata := locals.GcpColabRuntime.Metadata

	// Enable the Vertex AI API (Colab Enterprise's API) first so a fresh
	// project works on the first deploy. DisableOnDestroy stays false.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpcolr-aiplatform.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	// The runtime id and display name default to metadata.name -- identical
	// to the Terraform module. The id is always sent: the provider never
	// reads it back.
	runtimeId := spec.RuntimeId
	if runtimeId == "" {
		runtimeId = metadata.Name
	}
	displayName := spec.DisplayName
	if displayName == "" {
		displayName = metadata.Name
	}

	args := &colab.RuntimeArgs{
		Location:    pulumi.String(spec.Location),
		Name:        pulumi.String(runtimeId),
		DisplayName: pulumi.String(displayName),
		RuntimeUser: pulumi.String(spec.RuntimeUser),
		NotebookRuntimeTemplateRef: &colab.RuntimeNotebookRuntimeTemplateRefArgs{
			NotebookRuntimeTemplate: pulumi.String(spec.RuntimeTemplate.GetValue()),
		},
		AutoUpgrade: pulumi.Bool(spec.AutoUpgrade),
	}
	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.DesiredState != "" {
		args.DesiredState = pulumi.String(spec.DesiredState)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdRuntime, err := colab.NewRuntime(ctx, metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create colab runtime")
	}

	ctx.Export(OpName, createdRuntime.ID())
	ctx.Export(OpRuntimeId, createdRuntime.Name)
	ctx.Export(OpLocation, createdRuntime.Location)
	return nil
}
