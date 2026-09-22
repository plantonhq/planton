package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpvectorsearchcollectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvectorsearchcollection/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vectorsearch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// collection creates the Vector Search collection and, folded into it, one
// index per spec.indexes[] entry. Indexes belong to exactly one collection
// and nothing else in the catalog references one, so they ride the
// collection's lifecycle here rather than as blocks of their own.
//
// Send posture (parity with the Terraform module): Optional+Computed
// levers on an index -- distance_metric, dense_scann, dedicated
// infrastructure and its replica bounds -- are sent only when set, so
// Google's defaults stay in charge and a manifest that never mentions them
// re-plans clean on either engine.
func collection(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVectorSearchCollection.Spec

	// Enable the Vector Search API first so a fresh project works on the
	// first deploy. DisableOnDestroy stays false: tearing down one
	// collection must never disable the API for everything else in the
	// project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("vectorsearch.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpvsc-vectorsearch.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable vectorsearch.googleapis.com api")
	}

	args := &vectorsearch.CollectionArgs{
		CollectionId: pulumi.String(locals.CollectionId),
		Location:     pulumi.String(spec.Location),
		Labels:       pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.String(spec.DisplayName)
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.DataSchema != "" {
		args.DataSchema = pulumi.String(spec.DataSchema)
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &vectorsearch.CollectionEncryptionSpecArgs{
			CryptoKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}
	if len(spec.VectorSchemas) > 0 {
		args.VectorSchemas = buildVectorSchemas(spec.VectorSchemas)
	}

	// Engine-side destroy stance, fanned to every index below: PREVENT
	// fails destroys, ABANDON removes from management without deleting.
	// Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdCollection, err := vectorsearch.NewCollection(ctx,
		locals.GcpVectorSearchCollection.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create vector search collection")
	}

	// The folded indexes, one resource each, keyed by index_id. Every
	// setting but labels is immutable on Google's side, so a change here
	// replaces the index (rebuilt from the collection's data).
	indexNames := pulumi.StringArray{}
	for _, index := range spec.Indexes {
		indexArgs := &vectorsearch.IndexArgs{
			CollectionId: createdCollection.CollectionId,
			Location:     pulumi.String(spec.Location),
			IndexId:      pulumi.String(index.IndexId),
			IndexField:   pulumi.String(index.IndexField),
			Labels:       pulumi.ToStringMap(mergeLabels(locals.GcpLabels, index.Labels)),
		}
		if spec.ProjectId.GetValue() != "" {
			indexArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if index.DisplayName != "" {
			indexArgs.DisplayName = pulumi.String(index.DisplayName)
		}
		if index.Description != "" {
			indexArgs.Description = pulumi.String(index.Description)
		}
		if index.DistanceMetric != "" {
			indexArgs.DistanceMetric = pulumi.String(index.DistanceMetric)
		}
		if index.FeatureNormType != "" {
			indexArgs.DenseScann = &vectorsearch.IndexDenseScannArgs{
				FeatureNormType: pulumi.String(index.FeatureNormType),
			}
		}
		if len(index.FilterFields) > 0 {
			indexArgs.FilterFields = pulumi.ToStringArray(index.FilterFields)
		}
		if len(index.StoreFields) > 0 {
			indexArgs.StoreFields = pulumi.ToStringArray(index.StoreFields)
		}
		if index.DedicatedInfrastructure != nil {
			indexArgs.DedicatedInfrastructure = buildDedicatedInfrastructure(index.DedicatedInfrastructure)
		}
		if spec.DeletionPolicy != "" {
			indexArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}

		createdIndex, err := vectorsearch.NewIndex(ctx,
			fmt.Sprintf("%s-%s", locals.GcpVectorSearchCollection.Metadata.Name, index.IndexId), indexArgs,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdCollection))
		if err != nil {
			return errors.Wrapf(err, "failed to create vector search index %s", index.IndexId)
		}
		indexNames = append(indexNames, createdIndex.Name)
	}

	ctx.Export(OpName, createdCollection.Name)
	ctx.Export(OpCollectionId, createdCollection.CollectionId)
	ctx.Export(OpLocation, createdCollection.Location)
	ctx.Export(OpIndexNames, indexNames.ToStringArrayOutput())
	ctx.Export(OpIndexCount, pulumi.Int(len(spec.Indexes)))
	return nil
}

// buildVectorSchemas maps the spec's vector fields: a dense field carries
// its dimensions and optional Vertex embedding config; the sparse flag
// becomes Google's empty sparse_vector block (the spec's bool is the
// honest form of a block with no settings).
func buildVectorSchemas(fields []*gcpvectorsearchcollectionv1alpha1.GcpVectorSearchCollectionVectorSchema) vectorsearch.CollectionVectorSchemaArray {
	schemas := vectorsearch.CollectionVectorSchemaArray{}
	for _, field := range fields {
		schemaArgs := &vectorsearch.CollectionVectorSchemaArgs{
			FieldName: pulumi.String(field.FieldName),
		}
		if field.DenseVector != nil {
			dense := &vectorsearch.CollectionVectorSchemaDenseVectorArgs{}
			if field.DenseVector.Dimensions != nil {
				dense.Dimensions = pulumi.Int(int(field.DenseVector.GetDimensions()))
			}
			if cfg := field.DenseVector.VertexEmbeddingConfig; cfg != nil {
				dense.VertexEmbeddingConfig = &vectorsearch.CollectionVectorSchemaDenseVectorVertexEmbeddingConfigArgs{
					ModelId:      pulumi.String(cfg.ModelId),
					TaskType:     pulumi.String(cfg.TaskType),
					TextTemplate: pulumi.String(cfg.TextTemplate),
				}
			}
			schemaArgs.DenseVector = dense
		}
		if field.SparseVector {
			schemaArgs.SparseVector = &vectorsearch.CollectionVectorSchemaSparseVectorArgs{}
		}
		schemas = append(schemas, schemaArgs)
	}
	return schemas
}

// buildDedicatedInfrastructure sends the mode and replica bounds only when
// set; Google defaults PERFORMANCE_OPTIMIZED and two replicas otherwise.
func buildDedicatedInfrastructure(infra *gcpvectorsearchcollectionv1alpha1.GcpVectorSearchCollectionDedicatedInfrastructure) *vectorsearch.IndexDedicatedInfrastructureArgs {
	args := &vectorsearch.IndexDedicatedInfrastructureArgs{}
	if infra.Mode != "" {
		args.Mode = pulumi.String(infra.Mode)
	}
	if infra.AutoscalingSpec != nil {
		autoscaling := &vectorsearch.IndexDedicatedInfrastructureAutoscalingSpecArgs{}
		if infra.AutoscalingSpec.MinReplicaCount != nil {
			autoscaling.MinReplicaCount = pulumi.Int(int(infra.AutoscalingSpec.GetMinReplicaCount()))
		}
		if infra.AutoscalingSpec.MaxReplicaCount != nil {
			autoscaling.MaxReplicaCount = pulumi.Int(int(infra.AutoscalingSpec.GetMaxReplicaCount()))
		}
		args.AutoscalingSpec = autoscaling
	}
	return args
}

// mergeLabels lays an index's own labels under the platform attribution
// set so the attribution keys can never be clobbered -- the same order
// the Terraform module uses.
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
