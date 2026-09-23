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

// featureGroup creates the feature group -- the registry entry pointing
// Feature Store at a BigQuery table or view -- and, folded into it, one
// feature per spec.features[] entry. Features belong to exactly one group
// and nothing else in the catalog references one as a resource, so they
// ride the group's lifecycle here.
//
// Send posture (parity with the Terraform module): version_column_name is
// Optional+Computed and sent only when set, so Google's default (the
// column named like the feature) stays in charge.
func featureGroup(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiFeatureGroup.Spec
	resourceName := locals.GcpVertexAiFeatureGroup.Metadata.Name

	// Enable the Vertex AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one feature group
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
		"gcpvfg-aiplatform.googleapis.com", aiplatformApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	args := &vertex.AiFeatureGroupArgs{
		// The provider names the axis `region`; the spec keeps the Vertex
		// family's single word, `location`.
		Region: pulumi.String(spec.Location),
		Name:   pulumi.String(spec.FeatureGroupId),
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	// The spec lifts the one-leaf big_query_source wrapper; Google's
	// nesting is restored here. A GcpBigQueryTable reference resolves to
	// project.dataset.table, so the bq:// prefix Google stores is added
	// when missing.
	if spec.BigQuery != nil {
		bigQuery := &vertex.AiFeatureGroupBigQueryArgs{
			BigQuerySource: &vertex.AiFeatureGroupBigQueryBigQuerySourceArgs{
				InputUri: pulumi.String(bigQueryUri(spec.BigQuery.InputUri.GetValue())),
			},
		}
		if len(spec.BigQuery.EntityIdColumns) > 0 {
			bigQuery.EntityIdColumns = pulumi.ToStringArray(spec.BigQuery.EntityIdColumns)
		}
		args.BigQuery = bigQuery
	}

	// Engine-side destroy stance, fanned to every feature below: PREVENT
	// fails destroys, ABANDON removes from management without deleting.
	// Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdGroup, err := vertex.NewAiFeatureGroup(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAiplatformApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create vertex ai feature group")
	}

	featureNames := pulumi.StringArray{}
	for _, feature := range spec.Features {
		featureArgs := &vertex.AiFeatureGroupFeatureArgs{
			Region:       pulumi.String(spec.Location),
			FeatureGroup: createdGroup.Name,
			Name:         pulumi.String(feature.FeatureId),
			Labels:       pulumi.ToStringMap(mergeLabels(locals.GcpLabels, feature.Labels)),
		}
		if spec.ProjectId.GetValue() != "" {
			featureArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if feature.Description != "" {
			featureArgs.Description = pulumi.String(feature.Description)
		}
		if feature.VersionColumnName != "" {
			featureArgs.VersionColumnName = pulumi.String(feature.VersionColumnName)
		}
		if spec.DeletionPolicy != "" {
			featureArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}

		createdFeature, err := vertex.NewAiFeatureGroupFeature(ctx,
			fmt.Sprintf("%s-%s", resourceName, feature.FeatureId), featureArgs,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdGroup))
		if err != nil {
			return errors.Wrapf(err, "failed to create feature %s", feature.FeatureId)
		}
		featureNames = append(featureNames, createdFeature.ID().ToStringOutput())
	}

	ctx.Export(OpName, createdGroup.ID().ToStringOutput())
	ctx.Export(OpFeatureGroupId, createdGroup.Name)
	ctx.Export(OpLocation, createdGroup.Region)
	ctx.Export(OpFeatureNames, featureNames.ToStringArrayOutput())
	return nil
}

// bigQueryUri adds the bq:// prefix Google stores when the value is a bare
// project.dataset.table -- the same rule as the Terraform module.
func bigQueryUri(value string) string {
	if strings.HasPrefix(value, "bq://") {
		return value
	}
	return "bq://" + value
}

// mergeLabels lays a feature's own labels under the platform attribution
// set so the attribution keys can never be clobbered -- the same order the
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
