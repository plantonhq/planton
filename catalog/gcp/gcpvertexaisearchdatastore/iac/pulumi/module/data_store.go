package module

import (
	"github.com/pkg/errors"
	gcpvertexaisearchdatastorev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaisearchdatastore/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dataStore enables the Discovery Engine API and creates the data store
// (`google_discovery_engine_data_store`). Everything but display_name and
// kms_key_name is immutable on Google's side, so a change replaces the
// store (its documents included).
//
// Send posture (parity with the Terraform module): optional strings and
// blocks are sent only when set so Google's defaults stay in charge
// (content_config defaults to NO_CONTENT, solution_types to search, the
// parser to digital). The two virtual inputs -- create_advanced_site_search
// and skip_default_schema_creation -- are sent as declared; Google never
// reads them back.
func dataStore(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (*discoveryengine.DataStore, error) {
	spec := locals.GcpVertexAiSearchDataStore.Spec

	// Enable the Discovery Engine API first so a fresh project works on the
	// first deploy. DisableOnDestroy stays false: tearing down one store
	// must never disable the API for everything else in the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("discoveryengine.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpvsds-discoveryengine.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable discoveryengine.googleapis.com api")
	}

	args := &discoveryengine.DataStoreArgs{
		DataStoreId:               pulumi.String(locals.DataStoreId),
		Location:                  pulumi.String(spec.Location),
		DisplayName:               pulumi.String(locals.DisplayName),
		IndustryVertical:          pulumi.String(spec.IndustryVertical),
		AclEnabled:                pulumi.BoolPtr(spec.AclEnabled),
		CreateAdvancedSiteSearch:  pulumi.BoolPtr(spec.CreateAdvancedSiteSearch),
		SkipDefaultSchemaCreation: pulumi.BoolPtr(spec.SkipDefaultSchemaCreation),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.ContentConfig != "" {
		args.ContentConfig = pulumi.String(spec.ContentConfig)
	}
	if len(spec.SolutionTypes) > 0 {
		args.SolutionTypes = pulumi.ToStringArray(spec.SolutionTypes)
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.KmsKeyName = pulumi.String(spec.KmsKeyName.GetValue())
	}
	if spec.AdvancedSiteSearchConfig != nil {
		args.AdvancedSiteSearchConfig = &discoveryengine.DataStoreAdvancedSiteSearchConfigArgs{
			DisableInitialIndex:     pulumi.BoolPtr(spec.AdvancedSiteSearchConfig.DisableInitialIndex),
			DisableAutomaticRefresh: pulumi.BoolPtr(spec.AdvancedSiteSearchConfig.DisableAutomaticRefresh),
		}
	}
	if spec.DocumentProcessingConfig != nil {
		args.DocumentProcessingConfig = buildDocumentProcessingConfig(spec.DocumentProcessingConfig)
	}

	// Client-side destroy stance, fanned to the schema, target sites, and
	// sitemaps: PREVENT fails destroys, ABANDON removes from management
	// without deleting. Sent only when set so the provider default stays
	// in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdDataStore, err := discoveryengine.NewDataStore(ctx,
		locals.GcpVertexAiSearchDataStore.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create discovery engine data store")
	}
	return createdDataStore, nil
}

// buildDocumentProcessingConfig maps the parsing and chunking tree. The
// spec's digital_parsing bool becomes Google's empty digital_parsing_config
// block (the marker-block idiom); the layout and OCR arms carry their
// settings. The default parser and each per-file-type override share the
// spec's parsing-config message but are distinct SDK types, hence the two
// mapper pairs below.
func buildDocumentProcessingConfig(cfg *gcpvertexaisearchdatastorev1alpha1.GcpVertexAiSearchDataStoreDocumentProcessingConfig) *discoveryengine.DataStoreDocumentProcessingConfigArgs {
	args := &discoveryengine.DataStoreDocumentProcessingConfigArgs{}

	if cfg.ChunkingConfig != nil {
		layout := &discoveryengine.DataStoreDocumentProcessingConfigChunkingConfigLayoutBasedChunkingConfigArgs{
			IncludeAncestorHeadings: pulumi.BoolPtr(cfg.ChunkingConfig.IncludeAncestorHeadings),
		}
		if cfg.ChunkingConfig.ChunkSize != nil {
			layout.ChunkSize = pulumi.Int(int(cfg.ChunkingConfig.GetChunkSize()))
		}
		args.ChunkingConfig = &discoveryengine.DataStoreDocumentProcessingConfigChunkingConfigArgs{
			LayoutBasedChunkingConfig: layout,
		}
	}

	if p := cfg.DefaultParsingConfig; p != nil {
		def := &discoveryengine.DataStoreDocumentProcessingConfigDefaultParsingConfigArgs{}
		if p.DigitalParsing {
			def.DigitalParsingConfig = &discoveryengine.DataStoreDocumentProcessingConfigDefaultParsingConfigDigitalParsingConfigArgs{}
		}
		if l := p.LayoutParsingConfig; l != nil {
			def.LayoutParsingConfig = &discoveryengine.DataStoreDocumentProcessingConfigDefaultParsingConfigLayoutParsingConfigArgs{
				EnableGetProcessedDocument: pulumi.BoolPtr(l.EnableGetProcessedDocument),
				EnableImageAnnotation:      pulumi.BoolPtr(l.EnableImageAnnotation),
				EnableLlmLayoutParsing:     pulumi.BoolPtr(l.EnableLlmLayoutParsing),
				EnableTableAnnotation:      pulumi.BoolPtr(l.EnableTableAnnotation),
				ExcludeHtmlClasses:         stringArrayOrNil(l.ExcludeHtmlClasses),
				ExcludeHtmlElements:        stringArrayOrNil(l.ExcludeHtmlElements),
				ExcludeHtmlIds:             stringArrayOrNil(l.ExcludeHtmlIds),
				StructuredContentTypes:     stringArrayOrNil(l.StructuredContentTypes),
			}
		}
		if o := p.OcrParsingConfig; o != nil {
			def.OcrParsingConfig = &discoveryengine.DataStoreDocumentProcessingConfigDefaultParsingConfigOcrParsingConfigArgs{
				UseNativeText: pulumi.BoolPtr(o.UseNativeText),
			}
		}
		args.DefaultParsingConfig = def
	}

	overrides := discoveryengine.DataStoreDocumentProcessingConfigParsingConfigOverrideArray{}
	for _, override := range cfg.ParsingConfigOverrides {
		o := &discoveryengine.DataStoreDocumentProcessingConfigParsingConfigOverrideArgs{
			FileType: pulumi.String(override.FileType),
		}
		if p := override.ParsingConfig; p != nil {
			if p.DigitalParsing {
				o.DigitalParsingConfig = &discoveryengine.DataStoreDocumentProcessingConfigParsingConfigOverrideDigitalParsingConfigArgs{}
			}
			if l := p.LayoutParsingConfig; l != nil {
				o.LayoutParsingConfig = &discoveryengine.DataStoreDocumentProcessingConfigParsingConfigOverrideLayoutParsingConfigArgs{
					EnableGetProcessedDocument: pulumi.BoolPtr(l.EnableGetProcessedDocument),
					EnableImageAnnotation:      pulumi.BoolPtr(l.EnableImageAnnotation),
					EnableLlmLayoutParsing:     pulumi.BoolPtr(l.EnableLlmLayoutParsing),
					EnableTableAnnotation:      pulumi.BoolPtr(l.EnableTableAnnotation),
					ExcludeHtmlClasses:         stringArrayOrNil(l.ExcludeHtmlClasses),
					ExcludeHtmlElements:        stringArrayOrNil(l.ExcludeHtmlElements),
					ExcludeHtmlIds:             stringArrayOrNil(l.ExcludeHtmlIds),
					StructuredContentTypes:     stringArrayOrNil(l.StructuredContentTypes),
				}
			}
			if ocr := p.OcrParsingConfig; ocr != nil {
				o.OcrParsingConfig = &discoveryengine.DataStoreDocumentProcessingConfigParsingConfigOverrideOcrParsingConfigArgs{
					UseNativeText: pulumi.BoolPtr(ocr.UseNativeText),
				}
			}
		}
		overrides = append(overrides, o)
	}
	if len(overrides) > 0 {
		args.ParsingConfigOverrides = overrides
	}
	return args
}

// stringArrayOrNil sends a list only when it has members, so an empty
// spec list never renders as an empty argument the provider would diff.
func stringArrayOrNil(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}
