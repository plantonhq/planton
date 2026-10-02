package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// recommendationEngine builds the RECOMMENDATION arm
// (`google_discovery_engine_recommendation_engine`). The resource has no
// collection_id (it always lives in default_collection) and no
// kms_key_name; the spec walls both. The media config's levers are sent
// only when set so Google's per-type defaults apply.
func recommendationEngine(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	dependsOn []pulumi.Resource) (*createdEngine, error) {
	spec := locals.GcpVertexAiSearchEngine.Spec

	args := &discoveryengine.RecommendationEngineArgs{
		EngineId:     pulumi.String(locals.EngineId),
		DisplayName:  pulumi.String(locals.DisplayName),
		Location:     pulumi.String(spec.Location),
		DataStoreIds: dataStoreIds(spec),
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.IndustryVertical != "" {
		args.IndustryVertical = pulumi.String(spec.IndustryVertical)
	}
	if spec.CommonConfig != nil && spec.CommonConfig.CompanyName != "" {
		args.CommonConfig = &discoveryengine.RecommendationEngineCommonConfigArgs{
			CompanyName: pulumi.String(spec.CommonConfig.CompanyName),
		}
	}
	if media := spec.MediaRecommendationEngineConfig; media != nil {
		mediaArgs := &discoveryengine.RecommendationEngineMediaRecommendationEngineConfigArgs{}
		if media.Type != "" {
			mediaArgs.Type = pulumi.String(media.Type)
		}
		if media.OptimizationObjective != "" {
			mediaArgs.OptimizationObjective = pulumi.String(media.OptimizationObjective)
		}
		if media.TrainingState != "" {
			mediaArgs.TrainingState = pulumi.String(media.TrainingState)
		}
		if oc := media.OptimizationObjectiveConfig; oc != nil {
			ocArgs := &discoveryengine.RecommendationEngineMediaRecommendationEngineConfigOptimizationObjectiveConfigArgs{
				TargetField: pulumi.String(oc.TargetField),
			}
			if oc.TargetFieldValueFloat != nil {
				ocArgs.TargetFieldValueFloat = pulumi.Float64(float64(oc.GetTargetFieldValueFloat()))
			}
			mediaArgs.OptimizationObjectiveConfig = ocArgs
		}
		if fc := media.EngineFeaturesConfig; fc != nil {
			fcArgs := &discoveryengine.RecommendationEngineMediaRecommendationEngineConfigEngineFeaturesConfigArgs{}
			if mp := fc.MostPopularConfig; mp != nil {
				mpArgs := &discoveryengine.RecommendationEngineMediaRecommendationEngineConfigEngineFeaturesConfigMostPopularConfigArgs{}
				if mp.TimeWindowDays != nil {
					mpArgs.TimeWindowDays = pulumi.Int(int(mp.GetTimeWindowDays()))
				}
				fcArgs.MostPopularConfig = mpArgs
			}
			if rfy := fc.RecommendedForYouConfig; rfy != nil {
				rfyArgs := &discoveryengine.RecommendationEngineMediaRecommendationEngineConfigEngineFeaturesConfigRecommendedForYouConfigArgs{}
				if rfy.ContextEventType != "" {
					rfyArgs.ContextEventType = pulumi.String(rfy.ContextEventType)
				}
				fcArgs.RecommendedForYouConfig = rfyArgs
			}
			mediaArgs.EngineFeaturesConfig = fcArgs
		}
		args.MediaRecommendationEngineConfig = mediaArgs
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := discoveryengine.NewRecommendationEngine(ctx,
		locals.GcpVertexAiSearchEngine.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn(dependsOn))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create discovery engine recommendation engine")
	}
	return &createdEngine{
		Resource:        created,
		Name:            created.Name,
		EngineId:        created.EngineId,
		DialogflowAgent: pulumi.String("").ToStringOutput(),
	}, nil
}
