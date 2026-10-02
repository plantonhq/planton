package module

import (
	"github.com/pkg/errors"
	gcpmodelarmortemplatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmodelarmortemplate/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/modelarmor"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// template creates the Model Armor template. location and template_id are
// immutable; the filters, metadata, and labels update in place. Every
// optional string is sent only when set -- the Terraform module's posture.
func template(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpModelArmorTemplate.Spec

	// Enable the Model Armor API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one template must
	// never disable the API for everything else in the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("modelarmor.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpmarm-modelarmor.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable modelarmor.googleapis.com api")
	}

	templateId := spec.TemplateId
	if templateId == "" {
		templateId = locals.GcpModelArmorTemplate.Metadata.Name
	}

	args := &modelarmor.TemplateArgs{
		Location:     pulumi.String(spec.Location),
		TemplateId:   pulumi.String(templateId),
		Labels:       pulumi.ToStringMap(locals.GcpLabels),
		FilterConfig: buildFilterConfig(spec.FilterConfig),
	}
	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.TemplateMetadata != nil {
		args.TemplateMetadata = buildTemplateMetadata(spec.TemplateMetadata)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdTemplate, err := modelarmor.NewTemplate(ctx,
		locals.GcpModelArmorTemplate.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create model armor template")
	}

	ctx.Export(OpName, createdTemplate.Name)
	ctx.Export(OpTemplateId, createdTemplate.TemplateId)
	ctx.Export(OpLocation, createdTemplate.Location)
	return nil
}

func buildFilterConfig(fc *gcpmodelarmortemplatev1alpha1.GcpModelArmorTemplateFilterConfig) *modelarmor.TemplateFilterConfigArgs {
	args := &modelarmor.TemplateFilterConfigArgs{}
	if s := fc.MaliciousUriFilterSettings; s != nil {
		settings := &modelarmor.TemplateFilterConfigMaliciousUriFilterSettingsArgs{}
		if s.FilterEnforcement != "" {
			settings.FilterEnforcement = pulumi.String(s.FilterEnforcement)
		}
		args.MaliciousUriFilterSettings = settings
	}
	if s := fc.PiAndJailbreakFilterSettings; s != nil {
		settings := &modelarmor.TemplateFilterConfigPiAndJailbreakFilterSettingsArgs{}
		if s.FilterEnforcement != "" {
			settings.FilterEnforcement = pulumi.String(s.FilterEnforcement)
		}
		if s.ConfidenceLevel != "" {
			settings.ConfidenceLevel = pulumi.String(s.ConfidenceLevel)
		}
		args.PiAndJailbreakFilterSettings = settings
	}
	if s := fc.RaiSettings; s != nil {
		filters := modelarmor.TemplateFilterConfigRaiSettingsRaiFilterArray{}
		for _, f := range s.RaiFilters {
			filter := &modelarmor.TemplateFilterConfigRaiSettingsRaiFilterArgs{
				FilterType: pulumi.String(f.FilterType),
			}
			if f.ConfidenceLevel != "" {
				filter.ConfidenceLevel = pulumi.String(f.ConfidenceLevel)
			}
			filters = append(filters, filter)
		}
		args.RaiSettings = &modelarmor.TemplateFilterConfigRaiSettingsArgs{RaiFilters: filters}
	}
	if s := fc.SdpSettings; s != nil {
		settings := &modelarmor.TemplateFilterConfigSdpSettingsArgs{}
		if b := s.BasicConfig; b != nil {
			basic := &modelarmor.TemplateFilterConfigSdpSettingsBasicConfigArgs{}
			if b.FilterEnforcement != "" {
				basic.FilterEnforcement = pulumi.String(b.FilterEnforcement)
			}
			settings.BasicConfig = basic
		}
		if a := s.AdvancedConfig; a != nil {
			advanced := &modelarmor.TemplateFilterConfigSdpSettingsAdvancedConfigArgs{}
			if a.InspectTemplate != "" {
				advanced.InspectTemplate = pulumi.String(a.InspectTemplate)
			}
			if a.DeidentifyTemplate != "" {
				advanced.DeidentifyTemplate = pulumi.String(a.DeidentifyTemplate)
			}
			settings.AdvancedConfig = advanced
		}
		args.SdpSettings = settings
	}
	return args
}

func buildTemplateMetadata(tm *gcpmodelarmortemplatev1alpha1.GcpModelArmorTemplateMetadata) *modelarmor.TemplateTemplateMetadataArgs {
	args := &modelarmor.TemplateTemplateMetadataArgs{
		LogTemplateOperations:           pulumi.Bool(tm.LogTemplateOperations),
		LogSanitizeOperations:           pulumi.Bool(tm.LogSanitizeOperations),
		IgnorePartialInvocationFailures: pulumi.Bool(tm.IgnorePartialInvocationFailures),
	}
	if tm.CustomPromptSafetyErrorCode != 0 {
		args.CustomPromptSafetyErrorCode = pulumi.Int(int(tm.CustomPromptSafetyErrorCode))
	}
	if tm.CustomPromptSafetyErrorMessage != "" {
		args.CustomPromptSafetyErrorMessage = pulumi.String(tm.CustomPromptSafetyErrorMessage)
	}
	if tm.CustomLlmResponseSafetyErrorCode != 0 {
		args.CustomLlmResponseSafetyErrorCode = pulumi.Int(int(tm.CustomLlmResponseSafetyErrorCode))
	}
	if tm.CustomLlmResponseSafetyErrorMessage != "" {
		args.CustomLlmResponseSafetyErrorMessage = pulumi.String(tm.CustomLlmResponseSafetyErrorMessage)
	}
	if tm.EnforcementType != "" {
		args.EnforcementType = pulumi.String(tm.EnforcementType)
	}
	// The spec lifts Google's one-leaf multi_language_detection wrapper to a
	// bool; the block is sent only when it is true.
	if tm.EnableMultiLanguageDetection {
		args.MultiLanguageDetection = &modelarmor.TemplateTemplateMetadataMultiLanguageDetectionArgs{
			EnableMultiLanguageDetection: pulumi.Bool(true),
		}
	}
	if s := tm.FilterVersionSelector; s != nil {
		selector := &modelarmor.TemplateTemplateMetadataFilterVersionSelectorArgs{}
		if s.Alias != "" {
			selector.Alias = pulumi.String(s.Alias)
		}
		if s.Version != "" {
			selector.Version = pulumi.String(s.Version)
		}
		args.FilterVersionSelector = selector
	}
	return args
}
