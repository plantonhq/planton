package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ragEngineConfig sets the tier of RAG Engine's managed vector database
// for one project and location. Google owns the singleton
// (projects/{p}/locations/{l}/ragEngineConfig); the provider creates and
// deletes it by PATCH, so an apply over an already-configured location
// changes the tier in place and a destroy sets the location to
// UNPROVISIONED -- which deletes the managed database's data. ABANDON is
// the teardown for anyone who keeps corpora.
//
// The spec's tier enum is the honest form of Google's three exclusive
// empty blocks (basic / scaled / unprovisioned): exactly one is emitted,
// identical to the Terraform module's dynamic blocks.
func ragEngineConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiRagEngineConfig.Spec

	// Enable the Vertex AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down this block must
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
		"gcpragcf-aiplatform.googleapis.com", aiplatformApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	// Exactly one tier block, chosen by the spec's enum.
	dbConfig := &vertex.AiRagEngineConfigRagManagedDbConfigArgs{}
	switch spec.Tier {
	case "BASIC":
		dbConfig.Basic = &vertex.AiRagEngineConfigRagManagedDbConfigBasicArgs{}
	case "SCALED":
		dbConfig.Scaled = &vertex.AiRagEngineConfigRagManagedDbConfigScaledArgs{}
	case "UNPROVISIONED":
		dbConfig.Unprovisioned = &vertex.AiRagEngineConfigRagManagedDbConfigUnprovisionedArgs{}
	}

	args := &vertex.AiRagEngineConfigArgs{
		Region:             pulumi.String(spec.Location),
		RagManagedDbConfig: dbConfig,
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// Engine-side destroy stance: DELETE unprovisions the location (data
	// loss), PREVENT fails destroys, ABANDON leaves the tier as it is.
	// Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdConfig, err := vertex.NewAiRagEngineConfig(ctx,
		locals.GcpVertexAiRagEngineConfig.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAiplatformApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create rag engine config")
	}

	ctx.Export(OpName, createdConfig.Name)
	ctx.Export(OpLocation, createdConfig.Region)
	return nil
}
