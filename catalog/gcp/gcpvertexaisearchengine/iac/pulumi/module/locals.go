package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaisearchenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaisearchengine/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The three engine arms and the Discovery Engine solution type each one
// hard-codes (the provider sets solutionType per resource; controls must
// carry the same value, so the module derives it from the arm).
const (
	engineTypeSearch         = "SEARCH"
	engineTypeChat           = "CHAT"
	engineTypeRecommendation = "RECOMMENDATION"

	solutionTypeSearch         = "SOLUTION_TYPE_SEARCH"
	solutionTypeChat           = "SOLUTION_TYPE_CHAT"
	solutionTypeRecommendation = "SOLUTION_TYPE_RECOMMENDATION"

	// Google's collection every data store lives in unless a data
	// connector created its own.
	defaultCollectionId = "default_collection"
)

type Locals struct {
	GcpProviderConfig       *gcpprovider.GcpProviderConfig
	GcpVertexAiSearchEngine *gcpvertexaisearchenginev1alpha1.GcpVertexAiSearchEngine

	// EngineType is the resolved arm: spec.engine_type, or SEARCH when the
	// spec leaves it empty -- the same fallback the Terraform module applies
	// in locals.tf.
	EngineType string

	// SolutionType is the Discovery Engine solution the arm hard-codes,
	// carried onto every folded control.
	SolutionType string

	// EngineId is the engine's GCP id: spec.engine_id when set, otherwise
	// metadata.name.
	EngineId string

	// DisplayName is spec.display_name when set, otherwise metadata.name;
	// Google requires a display name, so the module always sends one.
	DisplayName string

	// CollectionId is spec.collection_id when set, otherwise Google's
	// default_collection. Every companion resource (controls, the serving
	// config, the widget config, assistants) is addressed under it.
	CollectionId string
}

// initializeLocals resolves the arm and the defaulted names. Discovery
// Engine resources carry no labels, so there is no attribution label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpvertexaisearchenginev1alpha1.GcpVertexAiSearchEngineStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiSearchEngine = stackInput.Target
	spec := locals.GcpVertexAiSearchEngine.Spec

	locals.EngineType = spec.EngineType
	if locals.EngineType == "" {
		locals.EngineType = engineTypeSearch
	}
	switch locals.EngineType {
	case engineTypeChat:
		locals.SolutionType = solutionTypeChat
	case engineTypeRecommendation:
		locals.SolutionType = solutionTypeRecommendation
	default:
		locals.SolutionType = solutionTypeSearch
	}

	locals.EngineId = spec.EngineId
	if locals.EngineId == "" {
		locals.EngineId = locals.GcpVertexAiSearchEngine.Metadata.Name
	}
	locals.DisplayName = spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = locals.GcpVertexAiSearchEngine.Metadata.Name
	}
	locals.CollectionId = spec.CollectionId.GetValue()
	if locals.CollectionId == "" {
		locals.CollectionId = defaultCollectionId
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}

// dataStoreIds flattens the spec's data store references to the ids the
// engine resources take (references resolve to literals before the module
// runs).
func dataStoreIds(spec *gcpvertexaisearchenginev1alpha1.GcpVertexAiSearchEngineSpec) pulumi.StringArray {
	ids := pulumi.StringArray{}
	for _, ref := range spec.DataStoreIds {
		ids = append(ids, pulumi.String(ref.GetValue()))
	}
	return ids
}
