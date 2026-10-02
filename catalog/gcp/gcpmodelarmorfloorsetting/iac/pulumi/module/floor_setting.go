package module

import (
	"strings"

	"github.com/pkg/errors"
	gcpmodelarmorfloorsettingv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmodelarmorfloorsetting/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/modelarmor"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// floorSetting applies the Model Armor floor setting. Google keeps exactly
// one per parent: create and update are the same PATCH, and destroy only
// stops managing it -- the last applied floor stays in force.
func floorSetting(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpModelArmorFloorSetting.Spec

	// Scope selection. Exactly one arm renders Google's parent; an empty
	// scope means the provider's default project, read from the provider's
	// configuration -- the Terraform module's google_client_config twin.
	scope := spec.Scope
	var floorProject, parent string
	switch {
	case scope.GetProjectId().GetValue() != "":
		floorProject = strings.TrimPrefix(scope.GetProjectId().GetValue(), "projects/")
	case scope.GetFolderId().GetValue() != "":
		parent = prefixed("folders/", scope.GetFolderId().GetValue())
	case scope.GetOrganizationId() != "":
		parent = "organizations/" + scope.GetOrganizationId()
	default:
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to resolve the provider's default project for the floor scope")
		}
		if clientConfig.Project == "" {
			return errors.New("the floor names no scope and the provider has no default project -- set spec.scope or configure a project")
		}
		floorProject = clientConfig.Project
	}

	var dependsOn []pulumi.Resource
	if floorProject != "" {
		parent = "projects/" + floorProject
		// A project floor needs the Model Armor API on that project; a
		// folder or organization floor has no project of its own.
		createdApi, err := projects.NewService(ctx, "gcpmafs-modelarmor.googleapis.com", &projects.ServiceArgs{
			Project:                  pulumi.String(floorProject),
			Service:                  pulumi.String("modelarmor.googleapis.com"),
			DisableDependentServices: pulumi.BoolPtr(true),
			DisableOnDestroy:         pulumi.BoolPtr(false),
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to enable modelarmor.googleapis.com api")
		}
		dependsOn = append(dependsOn, createdApi)
	}

	location := spec.Location
	if location == "" {
		location = "global"
	}

	args := &modelarmor.FloorsettingArgs{
		Parent:                        pulumi.String(parent),
		Location:                      pulumi.String(location),
		EnableFloorSettingEnforcement: pulumi.Bool(spec.EnableFloorSettingEnforcement),
		FilterConfig:                  buildFilterConfig(spec.FilterConfig),
	}
	if len(spec.IntegratedServices) > 0 {
		args.IntegratedServices = pulumi.ToStringArray(spec.IntegratedServices)
	}
	// Google wants exactly one of inspect_only / inspect_and_block; the spec
	// carries the choice as enforcement_type and only the chosen flag is
	// sent.
	if s := spec.AiPlatformFloorSetting; s != nil {
		setting := &modelarmor.FloorsettingAiPlatformFloorSettingArgs{
			EnableCloudLogging: pulumi.Bool(s.EnableCloudLogging),
		}
		if s.EnforcementType == "INSPECT_ONLY" {
			setting.InspectOnly = pulumi.Bool(true)
		} else {
			setting.InspectAndBlock = pulumi.Bool(true)
		}
		args.AiPlatformFloorSetting = setting
	}
	if s := spec.GoogleMcpServerFloorSetting; s != nil {
		setting := &modelarmor.FloorsettingGoogleMcpServerFloorSettingArgs{
			EnableCloudLogging: pulumi.Bool(s.EnableCloudLogging),
		}
		if s.EnforcementType == "INSPECT_ONLY" {
			setting.InspectOnly = pulumi.Bool(true)
		} else {
			setting.InspectAndBlock = pulumi.Bool(true)
		}
		args.GoogleMcpServerFloorSetting = setting
	}
	// The spec lifts Google's two one-leaf wrappers to a bool; the blocks
	// are sent only when it is true.
	if spec.EnableMultiLanguageDetection {
		args.FloorSettingMetadata = &modelarmor.FloorsettingFloorSettingMetadataArgs{
			MultiLanguageDetection: &modelarmor.FloorsettingFloorSettingMetadataMultiLanguageDetectionArgs{
				EnableMultiLanguageDetection: pulumi.Bool(true),
			},
		}
	}

	createdFloor, err := modelarmor.NewFloorsetting(ctx,
		locals.GcpModelArmorFloorSetting.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn(dependsOn))
	if err != nil {
		return errors.Wrap(err, "failed to apply model armor floor setting")
	}

	ctx.Export(OpName, createdFloor.ID())
	ctx.Export(OpParent, createdFloor.Parent)
	return nil
}

func buildFilterConfig(fc *gcpmodelarmorfloorsettingv1alpha1.GcpModelArmorFloorSettingFilterConfig) *modelarmor.FloorsettingFilterConfigArgs {
	args := &modelarmor.FloorsettingFilterConfigArgs{}
	if s := fc.MaliciousUriFilterSettings; s != nil {
		settings := &modelarmor.FloorsettingFilterConfigMaliciousUriFilterSettingsArgs{}
		if s.FilterEnforcement != "" {
			settings.FilterEnforcement = pulumi.String(s.FilterEnforcement)
		}
		args.MaliciousUriFilterSettings = settings
	}
	if s := fc.PiAndJailbreakFilterSettings; s != nil {
		settings := &modelarmor.FloorsettingFilterConfigPiAndJailbreakFilterSettingsArgs{}
		if s.FilterEnforcement != "" {
			settings.FilterEnforcement = pulumi.String(s.FilterEnforcement)
		}
		if s.ConfidenceLevel != "" {
			settings.ConfidenceLevel = pulumi.String(s.ConfidenceLevel)
		}
		args.PiAndJailbreakFilterSettings = settings
	}
	if s := fc.RaiSettings; s != nil {
		filters := modelarmor.FloorsettingFilterConfigRaiSettingsRaiFilterArray{}
		for _, f := range s.RaiFilters {
			filter := &modelarmor.FloorsettingFilterConfigRaiSettingsRaiFilterArgs{
				FilterType: pulumi.String(f.FilterType),
			}
			if f.ConfidenceLevel != "" {
				filter.ConfidenceLevel = pulumi.String(f.ConfidenceLevel)
			}
			filters = append(filters, filter)
		}
		args.RaiSettings = &modelarmor.FloorsettingFilterConfigRaiSettingsArgs{RaiFilters: filters}
	}
	if s := fc.SdpSettings; s != nil {
		settings := &modelarmor.FloorsettingFilterConfigSdpSettingsArgs{}
		if b := s.BasicConfig; b != nil {
			basic := &modelarmor.FloorsettingFilterConfigSdpSettingsBasicConfigArgs{}
			if b.FilterEnforcement != "" {
				basic.FilterEnforcement = pulumi.String(b.FilterEnforcement)
			}
			settings.BasicConfig = basic
		}
		if a := s.AdvancedConfig; a != nil {
			advanced := &modelarmor.FloorsettingFilterConfigSdpSettingsAdvancedConfigArgs{}
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

// prefixed adds Google's collection prefix to a bare id and leaves an
// already-prefixed value alone.
func prefixed(prefix, value string) string {
	if strings.HasPrefix(value, prefix) {
		return value
	}
	return prefix + value
}
