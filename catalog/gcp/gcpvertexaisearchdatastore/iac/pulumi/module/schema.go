package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// schema creates the store's one custom schema
// (`google_discovery_engine_schema`) when the spec declares one. Google
// keeps exactly one schema per store, so the spec requires the default
// schema skipped first (a CEL rule); every field is immutable, so a change
// replaces the schema. Returns nil when the spec declares none.
func schema(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	createdDataStore *discoveryengine.DataStore) (*discoveryengine.Schema, error) {
	spec := locals.GcpVertexAiSearchDataStore.Spec
	if spec.Schema == nil {
		return nil, nil
	}

	args := &discoveryengine.SchemaArgs{
		DataStoreId: createdDataStore.DataStoreId,
		Location:    pulumi.String(spec.Location),
		SchemaId:    pulumi.String(spec.Schema.SchemaId),
		JsonSchema:  pulumi.String(spec.Schema.JsonSchema),
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdSchema, err := discoveryengine.NewSchema(ctx,
		locals.GcpVertexAiSearchDataStore.Metadata.Name+"-schema", args,
		pulumi.Provider(gcpProvider),
		pulumi.Parent(createdDataStore))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create discovery engine schema")
	}
	return createdSchema, nil
}
