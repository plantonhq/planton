package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// searchEngine builds the SEARCH arm (`google_discovery_engine_search_engine`).
//
// Send posture (parity with the Terraform module): search_engine_config is
// ALWAYS sent because Google requires the block (empty when the spec
// leaves it out, so Google's Standard tier applies); app_type, features,
// kms_key_name, and knowledge_graph_config are sent only when set --
// features and the knowledge graph are Optional+Computed on Google's side
// and must not be defaulted by the module.
func searchEngine(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	dependsOn []pulumi.Resource) (*createdEngine, error) {
	spec := locals.GcpVertexAiSearchEngine.Spec

	searchConfig := &discoveryengine.SearchEngineSearchEngineConfigArgs{}
	if cfg := spec.SearchEngineConfig; cfg != nil {
		if cfg.SearchTier != "" {
			searchConfig.SearchTier = pulumi.String(cfg.SearchTier)
		}
		if len(cfg.SearchAddOns) > 0 {
			searchConfig.SearchAddOns = pulumi.ToStringArray(cfg.SearchAddOns)
		}
		if cfg.RequiredSubscriptionTier != "" {
			searchConfig.RequiredSubscriptionTier = pulumi.String(cfg.RequiredSubscriptionTier)
		}
	}

	args := &discoveryengine.SearchEngineArgs{
		EngineId:           pulumi.String(locals.EngineId),
		DisplayName:        pulumi.String(locals.DisplayName),
		Location:           pulumi.String(spec.Location),
		CollectionId:       pulumi.String(locals.CollectionId),
		DataStoreIds:       dataStoreIds(spec),
		SearchEngineConfig: searchConfig,
		DisableAnalytics:   pulumi.BoolPtr(spec.DisableAnalytics),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.IndustryVertical != "" {
		args.IndustryVertical = pulumi.String(spec.IndustryVertical)
	}
	if spec.CommonConfig != nil && spec.CommonConfig.CompanyName != "" {
		args.CommonConfig = &discoveryengine.SearchEngineCommonConfigArgs{
			CompanyName: pulumi.String(spec.CommonConfig.CompanyName),
		}
	}
	if spec.AppType != "" {
		args.AppType = pulumi.String(spec.AppType)
	}
	if len(spec.Features) > 0 {
		args.Features = pulumi.ToStringMap(spec.Features)
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.KmsKeyName = pulumi.String(spec.KmsKeyName.GetValue())
	}
	if kg := spec.KnowledgeGraphConfig; kg != nil {
		kgArgs := &discoveryengine.SearchEngineKnowledgeGraphConfigArgs{}
		if kg.EnableCloudKnowledgeGraph != nil {
			kgArgs.EnableCloudKnowledgeGraph = pulumi.BoolPtr(kg.GetEnableCloudKnowledgeGraph())
		}
		if kg.EnablePrivateKnowledgeGraph != nil {
			kgArgs.EnablePrivateKnowledgeGraph = pulumi.BoolPtr(kg.GetEnablePrivateKnowledgeGraph())
		}
		if len(kg.CloudKnowledgeGraphTypes) > 0 {
			kgArgs.CloudKnowledgeGraphTypes = pulumi.ToStringArray(kg.CloudKnowledgeGraphTypes)
		}
		if fc := kg.FeatureConfig; fc != nil {
			kgArgs.FeatureConfig = &discoveryengine.SearchEngineKnowledgeGraphConfigFeatureConfigArgs{
				DisablePrivateKgAutoComplete:       pulumi.BoolPtr(fc.DisablePrivateKgAutoComplete),
				DisablePrivateKgEnrichment:         pulumi.BoolPtr(fc.DisablePrivateKgEnrichment),
				DisablePrivateKgQueryUiChips:       pulumi.BoolPtr(fc.DisablePrivateKgQueryUiChips),
				DisablePrivateKgQueryUnderstanding: pulumi.BoolPtr(fc.DisablePrivateKgQueryUnderstanding),
			}
		}
		args.KnowledgeGraphConfig = kgArgs
	}

	// Client-side destroy stance, fanned to the controls and assistants:
	// PREVENT fails destroys, ABANDON removes from management without
	// deleting. Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := discoveryengine.NewSearchEngine(ctx,
		locals.GcpVertexAiSearchEngine.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn(dependsOn))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create discovery engine search engine")
	}
	return &createdEngine{
		Resource:        created,
		Name:            created.Name,
		EngineId:        created.EngineId,
		DialogflowAgent: pulumi.String("").ToStringOutput(),
	}, nil
}
