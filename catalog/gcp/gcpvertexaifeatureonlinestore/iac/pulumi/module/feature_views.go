package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	gcpvertexaifeatureonlinestorev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaifeatureonlinestore/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// featureViews creates one feature view per spec.feature_views[] entry.
// Views belong to exactly one store and nothing else in the catalog
// references one, so they ride the store's lifecycle here. Each has
// exactly one source (the spec's CEL); the sync cron is Optional+Computed
// and sent only when set. Returns the views' full names in manifest order.
func featureViews(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, createdStore *vertex.AiFeatureOnlineStore) (pulumi.StringArray, error) {
	spec := locals.GcpVertexAiFeatureOnlineStore.Spec
	names := pulumi.StringArray{}

	for _, view := range spec.FeatureViews {
		args := &vertex.AiFeatureOnlineStoreFeatureviewArgs{
			Region:             pulumi.String(spec.Location),
			FeatureOnlineStore: createdStore.Name,
			Name:               pulumi.String(view.FeatureViewId),
			Labels:             pulumi.ToStringMap(mergeLabels(locals.GcpLabels, view.Labels)),
		}
		if spec.ProjectId.GetValue() != "" {
			args.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if source := view.BigQuerySource; source != nil {
			args.BigQuerySource = &vertex.AiFeatureOnlineStoreFeatureviewBigQuerySourceArgs{
				Uri:             pulumi.String(bigQueryUri(source.Uri.GetValue())),
				EntityIdColumns: pulumi.ToStringArray(source.EntityIdColumns),
			}
		}
		if source := view.FeatureRegistrySource; source != nil {
			args.FeatureRegistrySource = buildFeatureRegistrySource(source)
		}
		if sync := view.SyncConfig; sync != nil {
			syncArgs := &vertex.AiFeatureOnlineStoreFeatureviewSyncConfigArgs{}
			if sync.Cron != "" {
				syncArgs.Cron = pulumi.String(sync.Cron)
			}
			if sync.Continuous {
				syncArgs.Continuous = pulumi.Bool(true)
			}
			args.SyncConfig = syncArgs
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}

		createdView, err := vertex.NewAiFeatureOnlineStoreFeatureview(ctx,
			fmt.Sprintf("%s-%s", locals.GcpVertexAiFeatureOnlineStore.Metadata.Name, view.FeatureViewId), args,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdStore))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create feature view %s", view.FeatureViewId)
		}
		names = append(names, createdView.ID().ToStringOutput())
	}
	return names, nil
}

// buildFeatureRegistrySource maps the feature groups and features a view
// serves; project_number is sent only when set.
func buildFeatureRegistrySource(source *gcpvertexaifeatureonlinestorev1alpha1.GcpVertexAiFeatureOnlineStoreFeatureRegistrySource) *vertex.AiFeatureOnlineStoreFeatureviewFeatureRegistrySourceArgs {
	groups := vertex.AiFeatureOnlineStoreFeatureviewFeatureRegistrySourceFeatureGroupArray{}
	for _, group := range source.FeatureGroups {
		groups = append(groups, &vertex.AiFeatureOnlineStoreFeatureviewFeatureRegistrySourceFeatureGroupArgs{
			FeatureGroupId: pulumi.String(group.FeatureGroupId.GetValue()),
			FeatureIds:     pulumi.ToStringArray(group.FeatureIds),
		})
	}
	args := &vertex.AiFeatureOnlineStoreFeatureviewFeatureRegistrySourceArgs{FeatureGroups: groups}
	if source.ProjectNumber.GetValue() != "" {
		args.ProjectNumber = pulumi.String(source.ProjectNumber.GetValue())
	}
	return args
}

// bigQueryUri adds the bq:// prefix Google stores when the value is a bare
// project.dataset.table -- the same rule as the Terraform module.
func bigQueryUri(value string) string {
	if strings.HasPrefix(value, "bq://") {
		return value
	}
	return "bq://" + value
}

// mergeLabels lays a view's own labels under the platform attribution set
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
