package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	gcpgkefleetfeaturev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkefleetfeature/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/gkehub"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// featureApis maps each feature to the API it needs beside the Fleet API,
// from Google's per-feature setup guides; features absent here need only
// the Fleet API. The Terraform module carries the same table.
var featureApis = map[string]string{
	"configmanagement":             "anthosconfigmanagement.googleapis.com",
	"policycontroller":             "anthospolicycontroller.googleapis.com",
	"servicemesh":                  "mesh.googleapis.com",
	"multiclusteringress":          "multiclusteringress.googleapis.com",
	"multiclusterservicediscovery": "multiclusterservicediscovery.googleapis.com",
}

// feature turns the fleet feature on with its fleet-wide settings and,
// folded into it, one per-cluster entry per spec.membership_configs[]
// item. Creating a feature that is already on adopts it; destroy turns it
// off (rbacrolebindingactuation only empties its allowlist, which the
// provider does itself).
func feature(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpGkeFleetFeature.Spec
	resourceName := locals.GcpGkeFleetFeature.Metadata.Name
	project := spec.GetProjectId().GetValue()

	var projectArg pulumi.StringPtrInput
	if project != "" {
		projectArg = pulumi.StringPtr(project)
	}
	var deletionPolicy pulumi.StringPtrInput
	if spec.DeletionPolicy != "" {
		deletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}
	// Fleet features live in "global" unless the manifest says otherwise
	// -- identical to the Terraform module.
	location := spec.Location
	if location == "" {
		location = "global"
	}

	// The Fleet API, and the feature's own API when it has one.
	// DisableOnDestroy stays false: turning one feature off must never
	// disable an API the rest of the fleet uses.
	apis := []pulumi.Resource{}
	services := []string{"gkehub.googleapis.com"}
	if api, ok := featureApis[spec.Feature]; ok {
		services = append(services, api)
	}
	for _, service := range services {
		apiArgs := &projects.ServiceArgs{
			Project:                  projectArg,
			Service:                  pulumi.String(service),
			DisableDependentServices: pulumi.BoolPtr(true),
			DisableOnDestroy:         pulumi.BoolPtr(false),
		}
		createdApi, err := projects.NewService(ctx, "gcpflft-"+service, apiArgs, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to enable %s", service)
		}
		apis = append(apis, createdApi)
	}

	args := &gkehub.FeatureArgs{
		Project:        projectArg,
		Name:           pulumi.StringPtr(spec.Feature),
		Location:       pulumi.String(location),
		Labels:         pulumi.ToStringMap(locals.GcpLabels),
		DeletionPolicy: deletionPolicy,
	}
	if featureSpec := featureSpecArgs(spec); featureSpec != nil {
		args.Spec = featureSpec
	}
	if defaults := spec.FleetDefaultMemberConfig; defaults != nil {
		args.FleetDefaultMemberConfig = fleetDefaultMemberConfigArgs(defaults)
	}

	createdFeature, err := gkehub.NewFeature(ctx, resourceName, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn(apis))
	if err != nil {
		return errors.Wrap(err, "failed to create fleet feature")
	}

	// Per-cluster settings, one entry of the feature's membershipSpecs map
	// each. The entry lives in this feature's project, which Google
	// requires the membership to share; the membership's location and ID
	// are parsed from its full name.
	for _, config := range spec.MembershipConfigs {
		membershipName := config.GetMembership().GetValue()
		membershipLocation, membershipId, err := parseMembershipName(membershipName)
		if err != nil {
			return errors.Wrap(err, "membership_configs")
		}
		membershipArgs := &gkehub.FeatureMembershipArgs{
			Project:            projectArg,
			Location:           pulumi.String(location),
			Feature:            createdFeature.Name,
			Membership:         pulumi.String(membershipId),
			MembershipLocation: pulumi.StringPtr(membershipLocation),
			DeletionPolicy:     deletionPolicy,
		}
		if config.Configmanagement != nil {
			membershipArgs.Configmanagement = membershipConfigManagementArgs(config.Configmanagement)
		}
		if config.Mesh != nil {
			membershipArgs.Mesh = &gkehub.FeatureMembershipMeshArgs{Management: pulumi.StringPtr(config.Mesh.Management)}
		}
		if config.Policycontroller != nil {
			membershipArgs.Policycontroller = membershipPolicyControllerArgs(config.Policycontroller)
		}
		if _, err := gkehub.NewFeatureMembership(ctx,
			fmt.Sprintf("%s-membership-%s-%s", resourceName, membershipLocation, membershipId),
			membershipArgs, pulumi.Provider(gcpProvider)); err != nil {
			return errors.Wrapf(err, "failed to configure membership %s", membershipName)
		}
	}

	ctx.Export(OpName, createdFeature.ID())
	return nil
}

// parseMembershipName returns the location and ID of a membership's full
// name, "projects/{p}/locations/{l}/memberships/{id}" -- the Terraform
// module's split() twin.
func parseMembershipName(name string) (location, membershipId string, err error) {
	parts := strings.Split(name, "/")
	if len(parts) != 6 || parts[0] != "projects" || parts[2] != "locations" || parts[4] != "memberships" {
		return "", "", errors.Errorf("membership %q is not projects/{project}/locations/{location}/memberships/{id}", name)
	}
	return parts[3], parts[5], nil
}

// optionalString sends a string only when it is set.
func optionalString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.StringPtr(value)
}

// optionalBool sends a presence-tracked bool only when it is set.
func optionalBool(value *bool) pulumi.BoolPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.BoolPtr(*value)
}

// optionalInt sends a presence-tracked integer only when it is set.
func optionalInt(value *int64) pulumi.IntPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.IntPtr(int(*value))
}

// syncWaitSecs sends the sync period as the decimal string the provider
// takes, only when it is set.
func syncWaitSecs(value int64) pulumi.StringPtrInput {
	if value <= 0 {
		return nil
	}
	return pulumi.StringPtr(fmt.Sprintf("%d", value))
}

// optionalStrings sends a list only when it has entries.
func optionalStrings(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}

// featureSpecArgs maps the lifted feature settings blocks onto the
// provider's spec block; nil when none is declared.
func featureSpecArgs(spec *gcpgkefleetfeaturev1alpha1.GcpGkeFleetFeatureSpec) *gkehub.FeatureSpecArgs {
	if spec.Multiclusteringress == nil && spec.Fleetobservability == nil && spec.Clusterupgrade == nil &&
		spec.Rbacrolebindingactuation == nil && spec.Workloadidentity == nil {
		return nil
	}
	args := &gkehub.FeatureSpecArgs{}

	if mci := spec.Multiclusteringress; mci != nil {
		args.Multiclusteringress = &gkehub.FeatureSpecMulticlusteringressArgs{
			ConfigMembership: pulumi.String(mci.GetConfigMembership().GetValue()),
		}
	}

	if observability := spec.Fleetobservability; observability != nil {
		observabilityArgs := &gkehub.FeatureSpecFleetobservabilityArgs{}
		if logging := observability.LoggingConfig; logging != nil {
			loggingArgs := &gkehub.FeatureSpecFleetobservabilityLoggingConfigArgs{}
			if logging.DefaultConfig != nil {
				loggingArgs.DefaultConfig = &gkehub.FeatureSpecFleetobservabilityLoggingConfigDefaultConfigArgs{
					Mode: optionalString(logging.DefaultConfig.Mode),
				}
			}
			if logging.FleetScopeLogsConfig != nil {
				loggingArgs.FleetScopeLogsConfig = &gkehub.FeatureSpecFleetobservabilityLoggingConfigFleetScopeLogsConfigArgs{
					Mode: optionalString(logging.FleetScopeLogsConfig.Mode),
				}
			}
			observabilityArgs.LoggingConfig = loggingArgs
		}
		args.Fleetobservability = observabilityArgs
	}

	if upgrade := spec.Clusterupgrade; upgrade != nil {
		upstreamFleets := []string{}
		for _, fleet := range upgrade.UpstreamFleets {
			upstreamFleets = append(upstreamFleets, fleet.GetValue())
		}
		upgradeArgs := &gkehub.FeatureSpecClusterupgradeArgs{
			UpstreamFleets: pulumi.ToStringArray(upstreamFleets),
		}
		// Optional+Computed: sent only when declared, so Google's default
		// never shows as a diff.
		if upgrade.PostConditions != nil {
			upgradeArgs.PostConditions = &gkehub.FeatureSpecClusterupgradePostConditionsArgs{
				Soaking: pulumi.String(upgrade.PostConditions.Soaking),
			}
		}
		if len(upgrade.GkeUpgradeOverrides) > 0 {
			overrides := gkehub.FeatureSpecClusterupgradeGkeUpgradeOverrideArray{}
			for _, override := range upgrade.GkeUpgradeOverrides {
				overrides = append(overrides, &gkehub.FeatureSpecClusterupgradeGkeUpgradeOverrideArgs{
					Upgrade: &gkehub.FeatureSpecClusterupgradeGkeUpgradeOverrideUpgradeArgs{
						Name:    pulumi.String(override.Upgrade.Name),
						Version: pulumi.String(override.Upgrade.Version),
					},
					PostConditions: &gkehub.FeatureSpecClusterupgradeGkeUpgradeOverridePostConditionsArgs{
						Soaking: pulumi.String(override.PostConditions.Soaking),
					},
				})
			}
			upgradeArgs.GkeUpgradeOverrides = overrides
		}
		args.Clusterupgrade = upgradeArgs
	}

	if actuation := spec.Rbacrolebindingactuation; actuation != nil {
		args.Rbacrolebindingactuation = &gkehub.FeatureSpecRbacrolebindingactuationArgs{
			AllowedCustomRoles: pulumi.ToStringArray(actuation.AllowedCustomRoles),
		}
	}

	if identity := spec.Workloadidentity; identity != nil {
		args.Workloadidentity = &gkehub.FeatureSpecWorkloadidentityArgs{
			ScopeTenancyPool: optionalString(identity.GetScopeTenancyPool().GetValue()),
		}
	}

	return args
}

// fleetDefaultMemberConfigArgs maps the fleet-wide member defaults.
func fleetDefaultMemberConfigArgs(defaults *gcpgkefleetfeaturev1alpha1.GcpGkeFleetFeatureFleetDefaultMemberConfig) *gkehub.FeatureFleetDefaultMemberConfigArgs {
	args := &gkehub.FeatureFleetDefaultMemberConfigArgs{}

	if cm := defaults.Configmanagement; cm != nil {
		cmArgs := &gkehub.FeatureFleetDefaultMemberConfigConfigmanagementArgs{
			Management: optionalString(cm.Management),
			Version:    optionalString(cm.Version),
		}
		if sync := cm.ConfigSync; sync != nil {
			syncArgs := &gkehub.FeatureFleetDefaultMemberConfigConfigmanagementConfigSyncArgs{
				Enabled:                       optionalBool(sync.Enabled),
				MetricsGcpServiceAccountEmail: optionalString(sync.GetMetricsGcpServiceAccountEmail().GetValue()),
				PreventDrift:                  optionalBool(sync.PreventDrift),
				SourceFormat:                  optionalString(sync.SourceFormat),
			}
			if git := sync.Git; git != nil {
				syncArgs.Git = &gkehub.FeatureFleetDefaultMemberConfigConfigmanagementConfigSyncGitArgs{
					SecretType:             pulumi.String(git.SecretType),
					GcpServiceAccountEmail: optionalString(git.GetGcpServiceAccountEmail().GetValue()),
					HttpsProxy:             optionalString(git.HttpsProxy),
					PolicyDir:              optionalString(git.PolicyDir),
					SyncBranch:             optionalString(git.SyncBranch),
					SyncRepo:               pulumi.StringPtr(git.SyncRepo),
					SyncRev:                optionalString(git.SyncRev),
					SyncWaitSecs:           syncWaitSecs(git.SyncWaitSecs),
				}
			}
			if oci := sync.Oci; oci != nil {
				syncArgs.Oci = &gkehub.FeatureFleetDefaultMemberConfigConfigmanagementConfigSyncOciArgs{
					SecretType:             pulumi.String(oci.SecretType),
					GcpServiceAccountEmail: optionalString(oci.GetGcpServiceAccountEmail().GetValue()),
					PolicyDir:              optionalString(oci.PolicyDir),
					SyncRepo:               pulumi.StringPtr(oci.SyncRepo),
					SyncWaitSecs:           syncWaitSecs(oci.SyncWaitSecs),
				}
			}
			cmArgs.ConfigSync = syncArgs
		}
		args.Configmanagement = cmArgs
	}

	if mesh := defaults.Mesh; mesh != nil {
		args.Mesh = &gkehub.FeatureFleetDefaultMemberConfigMeshArgs{Management: pulumi.String(mesh.Management)}
	}

	if pc := defaults.Policycontroller; pc != nil {
		hub := pc.PolicyControllerHubConfig
		hubArgs := &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigArgs{
			InstallSpec:              pulumi.String(hub.InstallSpec),
			AuditIntervalSeconds:     optionalInt(hub.AuditIntervalSeconds),
			ConstraintViolationLimit: optionalInt(hub.ConstraintViolationLimit),
			ExemptableNamespaces:     optionalStrings(hub.ExemptableNamespaces),
			LogDeniesEnabled:         pulumi.BoolPtr(hub.LogDeniesEnabled),
			MutationEnabled:          pulumi.BoolPtr(hub.MutationEnabled),
			ReferentialRulesEnabled:  pulumi.BoolPtr(hub.ReferentialRulesEnabled),
		}
		if hub.Monitoring != nil {
			hubArgs.Monitoring = &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigMonitoringArgs{
				Backends: pulumi.ToStringArray(hub.Monitoring.Backends),
			}
		}
		if len(hub.DeploymentConfigs) > 0 {
			deployments := gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigDeploymentConfigArray{}
			for _, deployment := range hub.DeploymentConfigs {
				deploymentArgs := &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigDeploymentConfigArgs{
					Component:    pulumi.String(deployment.Component),
					ReplicaCount: optionalInt(deployment.ReplicaCount),
					PodAffinity:  optionalString(deployment.PodAffinity),
				}
				if resources := deployment.ContainerResources; resources != nil {
					resourcesArgs := &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigDeploymentConfigContainerResourcesArgs{}
					if resources.Limits != nil {
						resourcesArgs.Limits = &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigDeploymentConfigContainerResourcesLimitsArgs{
							Cpu: optionalString(resources.Limits.Cpu), Memory: optionalString(resources.Limits.Memory),
						}
					}
					if resources.Requests != nil {
						resourcesArgs.Requests = &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigDeploymentConfigContainerResourcesRequestsArgs{
							Cpu: optionalString(resources.Requests.Cpu), Memory: optionalString(resources.Requests.Memory),
						}
					}
					deploymentArgs.ContainerResources = resourcesArgs
				}
				if len(deployment.PodTolerations) > 0 {
					tolerations := gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigDeploymentConfigPodTolerationArray{}
					for _, toleration := range deployment.PodTolerations {
						tolerations = append(tolerations, &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigDeploymentConfigPodTolerationArgs{
							Effect: optionalString(toleration.Effect), Key: optionalString(toleration.Key),
							Operator: optionalString(toleration.Operator), Value: optionalString(toleration.Value),
						})
					}
					deploymentArgs.PodTolerations = tolerations
				}
				deployments = append(deployments, deploymentArgs)
			}
			hubArgs.DeploymentConfigs = deployments
		}
		if content := hub.PolicyContent; content != nil {
			contentArgs := &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigPolicyContentArgs{}
			if len(content.Bundles) > 0 {
				bundles := gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigPolicyContentBundleArray{}
				for _, bundle := range content.Bundles {
					bundles = append(bundles, &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigPolicyContentBundleArgs{
						Bundle:             pulumi.String(bundle.Bundle),
						ExemptedNamespaces: optionalStrings(bundle.ExemptedNamespaces),
					})
				}
				contentArgs.Bundles = bundles
			}
			if content.TemplateLibrary != nil {
				contentArgs.TemplateLibrary = &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerPolicyControllerHubConfigPolicyContentTemplateLibraryArgs{
					Installation: optionalString(content.TemplateLibrary.Installation),
				}
			}
			hubArgs.PolicyContent = contentArgs
		}
		args.Policycontroller = &gkehub.FeatureFleetDefaultMemberConfigPolicycontrollerArgs{
			Version:                   optionalString(pc.Version),
			PolicyControllerHubConfig: hubArgs,
		}
	}

	return args
}

// membershipConfigManagementArgs maps one cluster's Config Sync settings.
func membershipConfigManagementArgs(cm *gcpgkefleetfeaturev1alpha1.GcpGkeFleetFeatureMembershipConfigManagement) *gkehub.FeatureMembershipConfigmanagementArgs {
	args := &gkehub.FeatureMembershipConfigmanagementArgs{
		Management: optionalString(cm.Management),
		Version:    optionalString(cm.Version),
	}
	sync := cm.ConfigSync
	if sync == nil {
		return args
	}
	syncArgs := &gkehub.FeatureMembershipConfigmanagementConfigSyncArgs{
		Enabled:                       optionalBool(sync.Enabled),
		MetricsGcpServiceAccountEmail: optionalString(sync.GetMetricsGcpServiceAccountEmail().GetValue()),
		PreventDrift:                  optionalBool(sync.PreventDrift),
		SourceFormat:                  optionalString(sync.SourceFormat),
		StopSyncing:                   optionalBool(sync.StopSyncing),
	}
	if git := sync.Git; git != nil {
		syncArgs.Git = &gkehub.FeatureMembershipConfigmanagementConfigSyncGitArgs{
			SecretType:             pulumi.StringPtr(git.SecretType),
			GcpServiceAccountEmail: optionalString(git.GetGcpServiceAccountEmail().GetValue()),
			HttpsProxy:             optionalString(git.HttpsProxy),
			PolicyDir:              optionalString(git.PolicyDir),
			SyncBranch:             optionalString(git.SyncBranch),
			SyncRepo:               pulumi.StringPtr(git.SyncRepo),
			SyncRev:                optionalString(git.SyncRev),
			SyncWaitSecs:           syncWaitSecs(git.SyncWaitSecs),
		}
	}
	if oci := sync.Oci; oci != nil {
		syncArgs.Oci = &gkehub.FeatureMembershipConfigmanagementConfigSyncOciArgs{
			SecretType:             pulumi.StringPtr(oci.SecretType),
			GcpServiceAccountEmail: optionalString(oci.GetGcpServiceAccountEmail().GetValue()),
			PolicyDir:              optionalString(oci.PolicyDir),
			SyncRepo:               pulumi.StringPtr(oci.SyncRepo),
			SyncWaitSecs:           syncWaitSecs(oci.SyncWaitSecs),
		}
	}
	if len(sync.DeploymentOverrides) > 0 {
		overrides := gkehub.FeatureMembershipConfigmanagementConfigSyncDeploymentOverrideArray{}
		for _, override := range sync.DeploymentOverrides {
			containers := gkehub.FeatureMembershipConfigmanagementConfigSyncDeploymentOverrideContainerArray{}
			for _, container := range override.Containers {
				containers = append(containers, &gkehub.FeatureMembershipConfigmanagementConfigSyncDeploymentOverrideContainerArgs{
					ContainerName: optionalString(container.ContainerName),
					CpuLimit:      optionalString(container.CpuLimit),
					CpuRequest:    optionalString(container.CpuRequest),
					MemoryLimit:   optionalString(container.MemoryLimit),
					MemoryRequest: optionalString(container.MemoryRequest),
				})
			}
			overrideArgs := &gkehub.FeatureMembershipConfigmanagementConfigSyncDeploymentOverrideArgs{
				DeploymentName:      optionalString(override.DeploymentName),
				DeploymentNamespace: optionalString(override.DeploymentNamespace),
			}
			if len(containers) > 0 {
				overrideArgs.Containers = containers
			}
			overrides = append(overrides, overrideArgs)
		}
		syncArgs.DeploymentOverrides = overrides
	}
	args.ConfigSync = syncArgs
	return args
}

// membershipPolicyControllerArgs maps one cluster's Policy Controller
// settings, under the membership resource's names (component_name,
// pod_tolerations, bundle_name).
func membershipPolicyControllerArgs(pc *gcpgkefleetfeaturev1alpha1.GcpGkeFleetFeaturePolicyController) *gkehub.FeatureMembershipPolicycontrollerArgs {
	hub := pc.PolicyControllerHubConfig
	hubArgs := &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigArgs{
		InstallSpec:              pulumi.StringPtr(hub.InstallSpec),
		AuditIntervalSeconds:     optionalInt(hub.AuditIntervalSeconds),
		ConstraintViolationLimit: optionalInt(hub.ConstraintViolationLimit),
		ExemptableNamespaces:     optionalStrings(hub.ExemptableNamespaces),
		LogDeniesEnabled:         pulumi.BoolPtr(hub.LogDeniesEnabled),
		MutationEnabled:          pulumi.BoolPtr(hub.MutationEnabled),
		ReferentialRulesEnabled:  pulumi.BoolPtr(hub.ReferentialRulesEnabled),
	}
	if hub.Monitoring != nil {
		hubArgs.Monitoring = &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigMonitoringArgs{
			Backends: pulumi.ToStringArray(hub.Monitoring.Backends),
		}
	}
	if len(hub.DeploymentConfigs) > 0 {
		deployments := gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigDeploymentConfigArray{}
		for _, deployment := range hub.DeploymentConfigs {
			deploymentArgs := &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigDeploymentConfigArgs{
				ComponentName: pulumi.String(deployment.Component),
				ReplicaCount:  optionalInt(deployment.ReplicaCount),
				PodAffinity:   optionalString(deployment.PodAffinity),
			}
			if resources := deployment.ContainerResources; resources != nil {
				resourcesArgs := &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigDeploymentConfigContainerResourcesArgs{}
				if resources.Limits != nil {
					resourcesArgs.Limits = &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigDeploymentConfigContainerResourcesLimitsArgs{
						Cpu: optionalString(resources.Limits.Cpu), Memory: optionalString(resources.Limits.Memory),
					}
				}
				if resources.Requests != nil {
					resourcesArgs.Requests = &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigDeploymentConfigContainerResourcesRequestsArgs{
						Cpu: optionalString(resources.Requests.Cpu), Memory: optionalString(resources.Requests.Memory),
					}
				}
				deploymentArgs.ContainerResources = resourcesArgs
			}
			if len(deployment.PodTolerations) > 0 {
				tolerations := gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigDeploymentConfigPodTolerationArray{}
				for _, toleration := range deployment.PodTolerations {
					tolerations = append(tolerations, &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigDeploymentConfigPodTolerationArgs{
						Effect: optionalString(toleration.Effect), Key: optionalString(toleration.Key),
						Operator: optionalString(toleration.Operator), Value: optionalString(toleration.Value),
					})
				}
				deploymentArgs.PodTolerations = tolerations
			}
			deployments = append(deployments, deploymentArgs)
		}
		hubArgs.DeploymentConfigs = deployments
	}
	if content := hub.PolicyContent; content != nil {
		contentArgs := &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigPolicyContentArgs{}
		if len(content.Bundles) > 0 {
			bundles := gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigPolicyContentBundleArray{}
			for _, bundle := range content.Bundles {
				bundles = append(bundles, &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigPolicyContentBundleArgs{
					BundleName:         pulumi.String(bundle.Bundle),
					ExemptedNamespaces: optionalStrings(bundle.ExemptedNamespaces),
				})
			}
			contentArgs.Bundles = bundles
		}
		if content.TemplateLibrary != nil {
			contentArgs.TemplateLibrary = &gkehub.FeatureMembershipPolicycontrollerPolicyControllerHubConfigPolicyContentTemplateLibraryArgs{
				Installation: optionalString(content.TemplateLibrary.Installation),
			}
		}
		hubArgs.PolicyContent = contentArgs
	}
	return &gkehub.FeatureMembershipPolicycontrollerArgs{
		Version:                   optionalString(pc.Version),
		PolicyControllerHubConfig: hubArgs,
	}
}
