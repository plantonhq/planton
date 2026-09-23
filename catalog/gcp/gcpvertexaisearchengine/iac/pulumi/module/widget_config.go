package module

import (
	"github.com/pkg/errors"
	gcpvertexaisearchenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaisearchengine/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// widgetConfig configures the engine's embeddable search widget
// (`google_discovery_engine_widget_config`). Google creates a search
// engine's "default_search_widget_config" with the engine; the provider's
// create is a PATCH on it and a widget config cannot be deleted, so
// destroy removes it from management and leaves it as configured. Every
// optional lever is sent only when set. Returns an empty name when the
// spec declares no widget config.
func widgetConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	created *createdEngine) (pulumi.StringOutput, error) {
	spec := locals.GcpVertexAiSearchEngine.Spec
	cfg := spec.WidgetConfig
	if cfg == nil {
		return pulumi.String("").ToStringOutput(), nil
	}

	args := &discoveryengine.WidgetConfigArgs{
		EngineId:     created.EngineId,
		Location:     pulumi.String(spec.Location),
		CollectionId: pulumi.String(locals.CollectionId),
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if access := cfg.AccessSettings; access != nil {
		accessArgs := &discoveryengine.WidgetConfigAccessSettingsArgs{
			AllowPublicAccess: pulumi.BoolPtr(access.AllowPublicAccess),
			EnableWebApp:      pulumi.BoolPtr(access.EnableWebApp),
		}
		if len(access.AllowlistedDomains) > 0 {
			accessArgs.AllowlistedDomains = pulumi.ToStringArray(access.AllowlistedDomains)
		}
		if access.LanguageCode != "" {
			accessArgs.LanguageCode = pulumi.String(access.LanguageCode)
		}
		if access.WorkforceIdentityPoolProvider != "" {
			accessArgs.WorkforceIdentityPoolProvider = pulumi.String(access.WorkforceIdentityPoolProvider)
		}
		args.AccessSettings = accessArgs
	}
	if home := cfg.HomepageSetting; home != nil {
		shortcuts := discoveryengine.WidgetConfigHomepageSettingShortcutArray{}
		for _, shortcut := range home.Shortcuts {
			shortcutArgs := &discoveryengine.WidgetConfigHomepageSettingShortcutArgs{}
			if shortcut.Title != "" {
				shortcutArgs.Title = pulumi.String(shortcut.Title)
			}
			if shortcut.DestinationUri != "" {
				shortcutArgs.DestinationUri = pulumi.String(shortcut.DestinationUri)
			}
			if shortcut.IconUrl != "" {
				shortcutArgs.Icon = &discoveryengine.WidgetConfigHomepageSettingShortcutIconArgs{
					Url: pulumi.String(shortcut.IconUrl),
				}
			}
			shortcuts = append(shortcuts, shortcutArgs)
		}
		args.HomepageSetting = &discoveryengine.WidgetConfigHomepageSettingArgs{Shortcuts: shortcuts}
	}
	if branding := cfg.UiBranding; branding != nil {
		brandingArgs := &discoveryengine.WidgetConfigUiBrandingArgs{}
		if branding.LogoUrl != "" {
			brandingArgs.Logo = &discoveryengine.WidgetConfigUiBrandingLogoArgs{Url: pulumi.String(branding.LogoUrl)}
		}
		args.UiBranding = brandingArgs
	}
	if ui := cfg.UiSettings; ui != nil {
		args.UiSettings = buildUiSettings(ui)
	}

	createdWidgetConfig, err := discoveryengine.NewWidgetConfig(ctx,
		locals.GcpVertexAiSearchEngine.Metadata.Name+"-widget-config", args,
		pulumi.Provider(gcpProvider),
		pulumi.Parent(created.Resource))
	if err != nil {
		return pulumi.StringOutput{}, errors.Wrap(err, "failed to configure discovery engine widget config")
	}
	return createdWidgetConfig.Name, nil
}

// buildUiSettings maps the widget's UI behavior; bools are sent as
// declared, strings and ints only when set.
func buildUiSettings(ui *gcpvertexaisearchenginev1alpha1.GcpVertexAiSearchEngineWidgetUiSettings) *discoveryengine.WidgetConfigUiSettingsArgs {
	args := &discoveryengine.WidgetConfigUiSettingsArgs{
		DisableUserEventsCollection: pulumi.BoolPtr(ui.DisableUserEventsCollection),
		EnableAutocomplete:          pulumi.BoolPtr(ui.EnableAutocomplete),
		EnableCreateAgentButton:     pulumi.BoolPtr(ui.EnableCreateAgentButton),
		EnablePeopleSearch:          pulumi.BoolPtr(ui.EnablePeopleSearch),
		EnableQualityFeedback:       pulumi.BoolPtr(ui.EnableQualityFeedback),
		EnableSafeSearch:            pulumi.BoolPtr(ui.EnableSafeSearch),
		EnableSearchAsYouType:       pulumi.BoolPtr(ui.EnableSearchAsYouType),
		EnableVisualContentSummary:  pulumi.BoolPtr(ui.EnableVisualContentSummary),
	}
	if ui.InteractionType != "" {
		args.InteractionType = pulumi.String(ui.InteractionType)
	}
	if ui.ResultDescriptionType != "" {
		args.ResultDescriptionType = pulumi.String(ui.ResultDescriptionType)
	}
	if ui.DefaultSearchRequestOrderBy != "" {
		args.DefaultSearchRequestOrderBy = pulumi.String(ui.DefaultSearchRequestOrderBy)
	}
	if len(ui.DataStoreUiConfigs) > 0 {
		configs := discoveryengine.WidgetConfigUiSettingsDataStoreUiConfigArray{}
		for _, storeUi := range ui.DataStoreUiConfigs {
			storeArgs := &discoveryengine.WidgetConfigUiSettingsDataStoreUiConfigArgs{
				Name: pulumi.String(storeUi.Name.GetValue()),
			}
			if len(storeUi.FacetFields) > 0 {
				facets := discoveryengine.WidgetConfigUiSettingsDataStoreUiConfigFacetFieldArray{}
				for _, facet := range storeUi.FacetFields {
					facetArgs := &discoveryengine.WidgetConfigUiSettingsDataStoreUiConfigFacetFieldArgs{
						Field: pulumi.String(facet.Field),
					}
					if facet.DisplayName != "" {
						facetArgs.DisplayName = pulumi.String(facet.DisplayName)
					}
					facets = append(facets, facetArgs)
				}
				storeArgs.FacetFields = facets
			}
			if len(storeUi.FieldsUiComponentsMap) > 0 {
				components := discoveryengine.WidgetConfigUiSettingsDataStoreUiConfigFieldsUiComponentsMapArray{}
				for _, component := range storeUi.FieldsUiComponentsMap {
					componentArgs := &discoveryengine.WidgetConfigUiSettingsDataStoreUiConfigFieldsUiComponentsMapArgs{
						UiComponent: pulumi.String(component.UiComponent),
						Field:       pulumi.String(component.Field),
					}
					if len(component.DeviceVisibility) > 0 {
						componentArgs.DeviceVisibilities = pulumi.ToStringArray(component.DeviceVisibility)
					}
					if component.DisplayTemplate != "" {
						componentArgs.DisplayTemplate = pulumi.String(component.DisplayTemplate)
					}
					components = append(components, componentArgs)
				}
				storeArgs.FieldsUiComponentsMaps = components
			}
			configs = append(configs, storeArgs)
		}
		args.DataStoreUiConfigs = configs
	}
	if gen := ui.GenerativeAnswerConfig; gen != nil {
		genArgs := &discoveryengine.WidgetConfigUiSettingsGenerativeAnswerConfigArgs{
			DisableRelatedQuestions:     pulumi.BoolPtr(gen.DisableRelatedQuestions),
			IgnoreAdversarialQuery:      pulumi.BoolPtr(gen.IgnoreAdversarialQuery),
			IgnoreLowRelevantContent:    pulumi.BoolPtr(gen.IgnoreLowRelevantContent),
			IgnoreNonAnswerSeekingQuery: pulumi.BoolPtr(gen.IgnoreNonAnswerSeekingQuery),
		}
		if gen.ImageSource != "" {
			genArgs.ImageSource = pulumi.String(gen.ImageSource)
		}
		if gen.LanguageCode != "" {
			genArgs.LanguageCode = pulumi.String(gen.LanguageCode)
		}
		if gen.MaxRephraseSteps != nil {
			genArgs.MaxRephraseSteps = pulumi.Int(int(gen.GetMaxRephraseSteps()))
		}
		if gen.ModelPromptPreamble != "" {
			genArgs.ModelPromptPreamble = pulumi.String(gen.ModelPromptPreamble)
		}
		if gen.ModelVersion != "" {
			genArgs.ModelVersion = pulumi.String(gen.ModelVersion)
		}
		if gen.ResultCount != nil {
			genArgs.ResultCount = pulumi.Int(int(gen.GetResultCount()))
		}
		args.GenerativeAnswerConfig = genArgs
	}
	return args
}
