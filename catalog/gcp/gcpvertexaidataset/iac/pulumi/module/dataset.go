package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dataset creates the Vertex AI managed dataset. Google assigns its
// numeric id at creation; location, metadata_schema_uri, and the
// encryption key are immutable, the display name and labels update in
// place.
func dataset(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiDataset.Spec

	// Enable the Vertex AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one dataset must
	// never disable the API for everything else in the project.
	aiplatformApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		aiplatformApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdAiplatformApi, err := projects.NewService(ctx,
		"gcpvds-aiplatform.googleapis.com", aiplatformApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	args := &vertex.AiDatasetArgs{
		// The provider names the axis `region`; the spec keeps the Vertex
		// family's single word, `location`.
		Region:            pulumi.String(spec.Location),
		DisplayName:       pulumi.String(locals.DisplayName),
		MetadataSchemaUri: pulumi.String(spec.MetadataSchemaUri),
		Labels:            pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &vertex.AiDatasetEncryptionSpecArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdDataset, err := vertex.NewAiDataset(ctx,
		locals.GcpVertexAiDataset.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAiplatformApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create vertex ai dataset")
	}

	ctx.Export(OpName, createdDataset.Name)
	ctx.Export(OpDatasetId, createdDataset.Name.ApplyT(lastPathSegment).(pulumi.StringOutput))
	ctx.Export(OpLocation, createdDataset.Region)
	return nil
}

// lastPathSegment returns the numeric id at the end of a resource name --
// the same split the Terraform module's outputs.tf performs.
func lastPathSegment(name string) string {
	return name[strings.LastIndex(name, "/")+1:]
}
