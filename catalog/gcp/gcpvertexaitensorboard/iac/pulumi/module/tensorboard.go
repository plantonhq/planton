package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tensorboard creates the TensorBoard instance and, folded into it, one
// experiment per spec.experiments[] entry and one run per
// experiments[].runs[] entry. Experiments and runs live and die with their
// TensorBoard and nothing else in the catalog references one, so they ride
// its lifecycle here rather than as blocks of their own.
func tensorboard(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiTensorboard.Spec
	resourceName := locals.GcpVertexAiTensorboard.Metadata.Name

	// Enable the Vertex AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one TensorBoard
	// must never disable the API for everything else in the project.
	aiplatformApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		aiplatformApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdAiplatformApi, err := projects.NewService(ctx,
		"gcpvtb-aiplatform.googleapis.com", aiplatformApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	args := &vertex.AiTensorboardArgs{
		// The provider names the axis `region`; the spec keeps the Vertex
		// family's single word, `location`.
		Region:      pulumi.String(spec.Location),
		DisplayName: pulumi.String(locals.DisplayName),
		Labels:      pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &vertex.AiTensorboardEncryptionSpecArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}

	// Engine-side destroy stance, fanned to every experiment and run below:
	// PREVENT fails destroys, ABANDON removes from management without
	// deleting. Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdTensorboard, err := vertex.NewAiTensorboard(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAiplatformApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create vertex ai tensorboard")
	}

	// Experiments and runs address their TensorBoard by the numeric id
	// Google assigned -- the last segment of its resource name.
	tensorboardId := createdTensorboard.Name.ApplyT(lastPathSegment).(pulumi.StringOutput)

	experimentNames := pulumi.StringArray{}
	runNames := pulumi.StringArray{}
	for _, experiment := range spec.Experiments {
		experimentArgs := &vertex.AiTensorboardExperimentArgs{
			Location:                pulumi.String(spec.Location),
			Tensorboard:             tensorboardId,
			TensorboardExperimentId: pulumi.String(experiment.ExperimentId),
			Labels:                  pulumi.ToStringMap(mergeLabels(locals.GcpLabels, experiment.Labels)),
		}
		if spec.ProjectId.GetValue() != "" {
			experimentArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if experiment.DisplayName != "" {
			experimentArgs.DisplayName = pulumi.String(experiment.DisplayName)
		}
		if experiment.Description != "" {
			experimentArgs.Description = pulumi.String(experiment.Description)
		}
		if experiment.Source != "" {
			experimentArgs.Source = pulumi.String(experiment.Source)
		}
		if spec.DeletionPolicy != "" {
			experimentArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}

		createdExperiment, err := vertex.NewAiTensorboardExperiment(ctx,
			fmt.Sprintf("%s-%s", resourceName, experiment.ExperimentId), experimentArgs,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdTensorboard))
		if err != nil {
			return errors.Wrapf(err, "failed to create tensorboard experiment %s", experiment.ExperimentId)
		}
		experimentNames = append(experimentNames, createdExperiment.ID().ToStringOutput())

		for _, run := range experiment.Runs {
			// Google requires a display name unique within the experiment;
			// it defaults to the run id (the Terraform module's rule).
			displayName := run.DisplayName
			if displayName == "" {
				displayName = run.RunId
			}
			runArgs := &vertex.AiTensorboardRunArgs{
				Location:         pulumi.String(spec.Location),
				Tensorboard:      tensorboardId,
				Experiment:       createdExperiment.TensorboardExperimentId,
				TensorboardRunId: pulumi.String(run.RunId),
				DisplayName:      pulumi.String(displayName),
				Labels:           pulumi.ToStringMap(mergeLabels(locals.GcpLabels, run.Labels)),
			}
			if spec.ProjectId.GetValue() != "" {
				runArgs.Project = pulumi.String(spec.ProjectId.GetValue())
			}
			if run.Description != "" {
				runArgs.Description = pulumi.String(run.Description)
			}
			if spec.DeletionPolicy != "" {
				runArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
			}

			createdRun, err := vertex.NewAiTensorboardRun(ctx,
				fmt.Sprintf("%s-%s-%s", resourceName, experiment.ExperimentId, run.RunId), runArgs,
				pulumi.Provider(gcpProvider),
				pulumi.Parent(createdExperiment))
			if err != nil {
				return errors.Wrapf(err, "failed to create tensorboard run %s/%s", experiment.ExperimentId, run.RunId)
			}
			runNames = append(runNames, createdRun.ID().ToStringOutput())
		}
	}

	ctx.Export(OpName, createdTensorboard.Name)
	ctx.Export(OpTensorboardId, tensorboardId)
	ctx.Export(OpLocation, createdTensorboard.Region)
	ctx.Export(OpBlobStoragePathPrefix, createdTensorboard.BlobStoragePathPrefix)
	ctx.Export(OpExperimentNames, experimentNames.ToStringArrayOutput())
	ctx.Export(OpRunNames, runNames.ToStringArrayOutput())
	return nil
}

// lastPathSegment returns the numeric id at the end of a resource name --
// the same split the Terraform module's locals.tf performs.
func lastPathSegment(name string) string {
	return name[strings.LastIndex(name, "/")+1:]
}

// mergeLabels lays a child's own labels under the platform attribution set
// so the attribution keys can never be clobbered -- the same order the
// Terraform module uses.
func mergeLabels(base map[string]string, own map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range own {
		merged[key] = value
	}
	for key, value := range base {
		merged[key] = value
	}
	return merged
}
