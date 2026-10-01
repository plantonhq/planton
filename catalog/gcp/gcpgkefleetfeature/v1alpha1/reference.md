# GcpGkeFleetFeature

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpGkeFleetFeatureSpec turns on one GKE fleet feature and configures it
(`google_gke_hub_feature`), together with the per-cluster settings that
override the fleet-wide defaults (`google_gke_hub_feature_membership`).

One block is one feature, named by `feature`:

  "configmanagement"             -- Config Sync: clusters sync their
                                    Kubernetes configuration from Git or
                                    an OCI image (fleet_default_member_config
                                    .configmanagement, membership_configs)
  "policycontroller"             -- Policy Controller: admission policies
                                    and audits, with Google's policy
                                    bundles (.policycontroller)
  "servicemesh"                  -- Cloud Service Mesh (.mesh)
  "multiclusteringress"          -- one ingress across clusters, hosted on
                                    a config cluster (multiclusteringress)
  "multiclusterservicediscovery" -- services exported across clusters (no
                                    settings)
  "fleetobservability"           -- fleet logging (fleetobservability)
  "clusterupgrade"               -- upgrade sequencing behind an upstream
                                    fleet (clusterupgrade)
  "rbacrolebindingactuation"     -- custom ClusterRoles team scopes may
                                    grant (rbacrolebindingactuation)
  "workloadidentity"             -- fleet tenancy's workload identity pool
                                    (workloadidentity)

Each settings block is accepted only on its feature: Google's API holds
one spec per feature, and a block on the wrong feature would never take
effect.

fleet_default_member_config is the recommended way to configure Config
Sync, Policy Controller, and Service Mesh: Google applies it to every
cluster in the fleet, including clusters that join later.
membership_configs overrides it for named clusters. A per-cluster entry
has no API object of its own -- Google stores it inside this feature --
so it lives here, keyed by the cluster's fleet membership.

Important behavioral notes:

  - Creating a feature that is already on adopts it: the existing
    feature's settings are replaced by this block's.
  - Destroy turns the feature off and removes its per-cluster settings,
    except "rbacrolebindingactuation", which Google never deletes:
    destroy empties its allowlist instead.
  - Both modules enable the Fleet API and the feature's own API
    (anthosconfigmanagement, anthospolicycontroller, mesh,
    multiclusteringress, multiclusterservicediscovery) and never disable
    them on destroy.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleetFeature
metadata:
  name: configmanagement
spec:
  projectId:
    value: my-gcp-project
  feature: configmanagement
  fleetDefaultMemberConfig:
    configmanagement:
      management: MANAGEMENT_AUTOMATIC
      configSync:
        enabled: true
        sourceFormat: unstructured
        git:
          secretType: none
          syncRepo: https://github.com/example/platform-config
          syncBranch: main
          policyDir: clusters/all
  membershipConfigs:
    - membership:
        value: projects/my-gcp-project/locations/us-central1/memberships/orders-uc1
      configmanagement:
        configSync:
          git:
            secretType: none
            syncRepo: https://github.com/example/platform-config
            syncBranch: main
            policyDir: clusters/orders-uc1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpGkeFleet (`status.outputs.project_id`) |
| `spec.feature` | `string` | yes |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.multiclusteringress` | `GcpGkeFleetFeatureMultiClusterIngress` |  |  |  |
| `spec.multiclusteringress.configMembership` | `string \| valueFrom` | yes |  | GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`) |
| `spec.fleetobservability` | `GcpGkeFleetFeatureFleetObservability` |  |  |  |
| `spec.fleetobservability.loggingConfig` | `GcpGkeFleetFeatureFleetLoggingConfig` |  |  |  |
| `spec.fleetobservability.loggingConfig.defaultConfig` | `GcpGkeFleetFeatureLogRoutingConfig` |  |  |  |
| `spec.fleetobservability.loggingConfig.defaultConfig.mode` | `string` |  |  |  |
| `spec.fleetobservability.loggingConfig.fleetScopeLogsConfig` | `GcpGkeFleetFeatureLogRoutingConfig` |  |  |  |
| `spec.fleetobservability.loggingConfig.fleetScopeLogsConfig.mode` | `string` |  |  |  |
| `spec.clusterupgrade` | `GcpGkeFleetFeatureClusterUpgrade` |  |  |  |
| `spec.clusterupgrade.upstreamFleets` | `[]string \| valueFrom` | yes |  | GcpGkeFleet (`status.outputs.project_id`) |
| `spec.clusterupgrade.postConditions` | `GcpGkeFleetFeatureUpgradePostConditions` |  |  |  |
| `spec.clusterupgrade.postConditions.soaking` | `string` | yes |  |  |
| `spec.clusterupgrade.gkeUpgradeOverrides` | `[]GcpGkeFleetFeatureGkeUpgradeOverride` |  |  |  |
| `spec.clusterupgrade.gkeUpgradeOverrides[].upgrade` | `GcpGkeFleetFeatureGkeUpgrade` | yes |  |  |
| `spec.clusterupgrade.gkeUpgradeOverrides[].upgrade.name` | `string` | yes |  |  |
| `spec.clusterupgrade.gkeUpgradeOverrides[].upgrade.version` | `string` | yes |  |  |
| `spec.clusterupgrade.gkeUpgradeOverrides[].postConditions` | `GcpGkeFleetFeatureUpgradePostConditions` | yes |  |  |
| `spec.clusterupgrade.gkeUpgradeOverrides[].postConditions.soaking` | `string` | yes |  |  |
| `spec.rbacrolebindingactuation` | `GcpGkeFleetFeatureRbacRoleBindingActuation` |  |  |  |
| `spec.rbacrolebindingactuation.allowedCustomRoles` | `[]string` |  |  |  |
| `spec.workloadidentity` | `GcpGkeFleetFeatureWorkloadIdentity` |  |  |  |
| `spec.workloadidentity.scopeTenancyPool` | `string \| valueFrom` |  |  | GcpWorkloadIdentityPool (`status.outputs.name`) |
| `spec.fleetDefaultMemberConfig` | `GcpGkeFleetFeatureFleetDefaultMemberConfig` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement` | `GcpGkeFleetFeatureConfigManagement` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.management` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.version` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync` | `GcpGkeFleetFeatureConfigSync` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.enabled` | `bool` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git` | `GcpGkeFleetFeatureConfigSyncGit` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git.secretType` | `string` | yes |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git.gcpServiceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git.httpsProxy` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git.policyDir` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git.syncBranch` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git.syncRepo` | `string` | yes |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git.syncRev` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git.syncWaitSecs` | `int64` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.oci` | `GcpGkeFleetFeatureConfigSyncOci` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.secretType` | `string` | yes |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.gcpServiceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.policyDir` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.syncRepo` | `string` | yes |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.syncWaitSecs` | `int64` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.metricsGcpServiceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.preventDrift` | `bool` |  |  |  |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.sourceFormat` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.mesh` | `GcpGkeFleetFeatureMesh` |  |  |  |
| `spec.fleetDefaultMemberConfig.mesh.management` | `string` | yes |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller` | `GcpGkeFleetFeaturePolicyController` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.version` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig` | `GcpGkeFleetFeaturePolicyControllerHubConfig` | yes |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.installSpec` | `string` | yes |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.auditIntervalSeconds` | `int64` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.constraintViolationLimit` | `int64` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.exemptableNamespaces` | `[]string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.logDeniesEnabled` | `bool` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.mutationEnabled` | `bool` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.referentialRulesEnabled` | `bool` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.monitoring` | `GcpGkeFleetFeaturePolicyControllerMonitoring` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.monitoring.backends` | `[]string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs` | `[]GcpGkeFleetFeaturePolicyControllerDeploymentConfig` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].component` | `string` | yes |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].replicaCount` | `int64` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podAffinity` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources` | `GcpGkeFleetFeaturePolicyControllerContainerResources` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits` | `GcpGkeFleetFeaturePolicyControllerResourceList` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits.cpu` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits.memory` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests` | `GcpGkeFleetFeaturePolicyControllerResourceList` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests.cpu` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests.memory` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations` | `[]GcpGkeFleetFeaturePolicyControllerToleration` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].effect` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].key` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].operator` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].value` | `string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent` | `GcpGkeFleetFeaturePolicyControllerPolicyContent` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.bundles` | `[]GcpGkeFleetFeaturePolicyControllerBundle` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.bundles[].bundle` | `string` | yes |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.bundles[].exemptedNamespaces` | `[]string` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.templateLibrary` | `GcpGkeFleetFeaturePolicyControllerTemplateLibrary` |  |  |  |
| `spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.templateLibrary.installation` | `string` |  |  |  |
| `spec.membershipConfigs` | `[]GcpGkeFleetFeatureMembershipConfig` |  |  |  |
| `spec.membershipConfigs[].membership` | `string \| valueFrom` | yes |  | GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`) |
| `spec.membershipConfigs[].configmanagement` | `GcpGkeFleetFeatureMembershipConfigManagement` |  |  |  |
| `spec.membershipConfigs[].configmanagement.management` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.version` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync` | `GcpGkeFleetFeatureMembershipConfigSync` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.enabled` | `bool` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.git` | `GcpGkeFleetFeatureConfigSyncGit` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.git.secretType` | `string` | yes |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.git.gcpServiceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.membershipConfigs[].configmanagement.configSync.git.httpsProxy` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.git.policyDir` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.git.syncBranch` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.git.syncRepo` | `string` | yes |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.git.syncRev` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.git.syncWaitSecs` | `int64` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.oci` | `GcpGkeFleetFeatureConfigSyncOci` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.oci.secretType` | `string` | yes |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.oci.gcpServiceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.membershipConfigs[].configmanagement.configSync.oci.policyDir` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.oci.syncRepo` | `string` | yes |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.oci.syncWaitSecs` | `int64` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.metricsGcpServiceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.membershipConfigs[].configmanagement.configSync.preventDrift` | `bool` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.sourceFormat` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.stopSyncing` | `bool` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides` | `[]GcpGkeFleetFeatureConfigSyncDeploymentOverride` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].deploymentName` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].deploymentNamespace` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers` | `[]GcpGkeFleetFeatureConfigSyncContainerOverride` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].containerName` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].cpuLimit` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].cpuRequest` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].memoryLimit` | `string` |  |  |  |
| `spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].memoryRequest` | `string` |  |  |  |
| `spec.membershipConfigs[].mesh` | `GcpGkeFleetFeatureMesh` |  |  |  |
| `spec.membershipConfigs[].mesh.management` | `string` | yes |  |  |
| `spec.membershipConfigs[].policycontroller` | `GcpGkeFleetFeaturePolicyController` |  |  |  |
| `spec.membershipConfigs[].policycontroller.version` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig` | `GcpGkeFleetFeaturePolicyControllerHubConfig` | yes |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.installSpec` | `string` | yes |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.auditIntervalSeconds` | `int64` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.constraintViolationLimit` | `int64` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.exemptableNamespaces` | `[]string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.logDeniesEnabled` | `bool` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.mutationEnabled` | `bool` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.referentialRulesEnabled` | `bool` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.monitoring` | `GcpGkeFleetFeaturePolicyControllerMonitoring` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.monitoring.backends` | `[]string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs` | `[]GcpGkeFleetFeaturePolicyControllerDeploymentConfig` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].component` | `string` | yes |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].replicaCount` | `int64` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podAffinity` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources` | `GcpGkeFleetFeaturePolicyControllerContainerResources` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits` | `GcpGkeFleetFeaturePolicyControllerResourceList` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits.cpu` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits.memory` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests` | `GcpGkeFleetFeaturePolicyControllerResourceList` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests.cpu` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests.memory` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations` | `[]GcpGkeFleetFeaturePolicyControllerToleration` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].effect` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].key` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].operator` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].value` | `string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent` | `GcpGkeFleetFeaturePolicyControllerPolicyContent` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.bundles` | `[]GcpGkeFleetFeaturePolicyControllerBundle` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.bundles[].bundle` | `string` | yes |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.bundles[].exemptedNamespaces` | `[]string` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.templateLibrary` | `GcpGkeFleetFeaturePolicyControllerTemplateLibrary` |  |  |  |
| `spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.templateLibrary.installation` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The fleet host project: a literal project ID or a GcpGkeFleet
reference (the fleet this feature configures, which orders it after
the fleet in a chart). Empty means the provider's default project.
Immutable.

- references: GcpGkeFleet (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeFleet, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.feature

`string` · required

Which feature, by Google's feature name (see the list above): one of
the documented names, or another lowercase name Google adds. Immutable.

- rule: feature must be a lowercase Google fleet feature name, e.g. configmanagement
- rule: {"required":true}

### spec.location

`string`

Where the feature lives. Fleet features are "global" (the default).
Immutable.

- rule: location must be global or a region

### spec.labels

`map<string, string>`

Labels on the feature. The platform attribution labels are added on
top and win on key conflicts.

### spec.multiclusteringress

`GcpGkeFleetFeatureMultiClusterIngress`

Settings for "multiclusteringress".

### spec.multiclusteringress.configMembership

`string | valueFrom` · required

The config cluster's fleet membership, by full name: a
GcpGkeFleetMembership reference, a GcpGkeCluster reference (its
fleet_membership output), or the literal name. Required.

- references: GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`)
- rule: config_membership must be a full membership name: projects/{project}/locations/{location}/memberships/{id}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeFleetMembership, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.fleetobservability

`GcpGkeFleetFeatureFleetObservability`

Settings for "fleetobservability".

### spec.fleetobservability.loggingConfig

`GcpGkeFleetFeatureFleetLoggingConfig`

The routing configuration.

### spec.fleetobservability.loggingConfig.defaultConfig

`GcpGkeFleetFeatureLogRoutingConfig`

Routing for logs no other route covers.

### spec.fleetobservability.loggingConfig.defaultConfig.mode

`string`

How the logs reach the fleet host project:
  "COPY" -- copied there and kept in the cluster's project too
  "MOVE" -- moved there only
Empty leaves the route off.

- rule: mode must be COPY or MOVE

### spec.fleetobservability.loggingConfig.fleetScopeLogsConfig

`GcpGkeFleetFeatureLogRoutingConfig`

Routing for all logs of every fleet scope.

### spec.fleetobservability.loggingConfig.fleetScopeLogsConfig.mode

`string`

How the logs reach the fleet host project:
  "COPY" -- copied there and kept in the cluster's project too
  "MOVE" -- moved there only
Empty leaves the route off.

- rule: mode must be COPY or MOVE

### spec.clusterupgrade

`GcpGkeFleetFeatureClusterUpgrade`

Settings for "clusterupgrade".

### spec.clusterupgrade.upstreamFleets

`[]string | valueFrom` · required

The upstream fleet whose completed upgrades this fleet consumes: its
host project's ID or number, as a literal or a GcpGkeFleet reference.
Google accepts at most one today; the list leaves room for more.

- references: GcpGkeFleet (`status.outputs.project_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeFleet, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.clusterupgrade.postConditions

`GcpGkeFleetFeatureUpgradePostConditions`

When an upgrade counts as complete in this fleet. Empty takes Google's
default.

### spec.clusterupgrade.postConditions.soaking

`string` · required

How long to soak after the rollout finishes before marking it
complete, as seconds with an "s" suffix ("604800s" is seven days); at
most 30 days. Required.

- rule: soaking must be a duration in seconds with an s suffix, e.g. 604800s
- rule: {"required":true}

### spec.clusterupgrade.gkeUpgradeOverrides

`[]GcpGkeFleetFeatureGkeUpgradeOverride`

Different completion conditions for individual upgrades.

### spec.clusterupgrade.gkeUpgradeOverrides[].upgrade

`GcpGkeFleetFeatureGkeUpgrade` · required

Which upgrade. Required.

- rule: {"required":true}

### spec.clusterupgrade.gkeUpgradeOverrides[].upgrade.name

`string` · required

The upgrade's name, e.g. "k8s_control_plane" or "k8s_node". At most 99
characters. Required.

- rule: {"string":{"minLen":"1","maxLen":"99"}}

### spec.clusterupgrade.gkeUpgradeOverrides[].upgrade.version

`string` · required

The upgrade's version, e.g. "1.31.1-gke.1146000". At most 99
characters. Required.

- rule: {"string":{"minLen":"1","maxLen":"99"}}

### spec.clusterupgrade.gkeUpgradeOverrides[].postConditions

`GcpGkeFleetFeatureUpgradePostConditions` · required

Its completion conditions. Required.

- rule: {"required":true}

### spec.clusterupgrade.gkeUpgradeOverrides[].postConditions.soaking

`string` · required

How long to soak after the rollout finishes before marking it
complete, as seconds with an "s" suffix ("604800s" is seven days); at
most 30 days. Required.

- rule: soaking must be a duration in seconds with an s suffix, e.g. 604800s
- rule: {"required":true}

### spec.rbacrolebindingactuation

`GcpGkeFleetFeatureRbacRoleBindingActuation`

Settings for "rbacrolebindingactuation".

### spec.rbacrolebindingactuation.allowedCustomRoles

`[]string`

The Kubernetes ClusterRoles a scope role binding may grant through
custom_role. A role in use cannot leave the list until the bindings
that use it are gone.

- rule: {"repeated":{"unique":true}}

### spec.workloadidentity

`GcpGkeFleetFeatureWorkloadIdentity`

Settings for "workloadidentity".

### spec.workloadidentity.scopeTenancyPool

`string | valueFrom`

The workload identity pool, in trust-domain mode, that fleet tenancy
uses so identities stay the same across the clusters of a scope: a
GcpWorkloadIdentityPool reference (its name) or the literal pool name.

- references: GcpWorkloadIdentityPool (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpWorkloadIdentityPool, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.fleetDefaultMemberConfig

`GcpGkeFleetFeatureFleetDefaultMemberConfig`

The fleet-wide defaults for "configmanagement", "servicemesh", or
"policycontroller": applied to every cluster in the fleet, including
clusters that join later, unless membership_configs overrides one.

### spec.fleetDefaultMemberConfig.configmanagement

`GcpGkeFleetFeatureConfigManagement`

Config Sync for every cluster (feature "configmanagement").

### spec.fleetDefaultMemberConfig.configmanagement.management

`string`

Upgrades: "MANAGEMENT_AUTOMATIC" lets Google keep Config Sync current;
"MANAGEMENT_MANUAL" pins it to `version`. Empty takes Google's default.

- rule: management must be MANAGEMENT_AUTOMATIC or MANAGEMENT_MANUAL

### spec.fleetDefaultMemberConfig.configmanagement.version

`string`

The Config Sync version to install, e.g. "1.22.0"; meaningful with
MANAGEMENT_MANUAL. Empty takes the current release.

### spec.fleetDefaultMemberConfig.configmanagement.configSync

`GcpGkeFleetFeatureConfigSync`

What to sync and how.

- rule: config_sync syncs from git or oci, not both

### spec.fleetDefaultMemberConfig.configmanagement.configSync.enabled

`bool` · optional (explicit presence)

true installs Config Sync and applies these settings; false removes it
and ignores the rest. Unset installs it exactly when git or oci is set.

### spec.fleetDefaultMemberConfig.configmanagement.configSync.git

`GcpGkeFleetFeatureConfigSyncGit`

Sync from a Git repository.

### spec.fleetDefaultMemberConfig.configmanagement.configSync.git.secretType

`string` · required

How Config Sync authenticates to the repository:
  "none" (public repository), "ssh", "cookiefile", "token",
  "gcenode" (the node's service account), "gcpserviceaccount" (a
  Google service account through Workload Identity, see
  gcp_service_account_email), "githubapp".
Required; Google matches it case-sensitively.

- rule: secret_type must be one of: none, ssh, cookiefile, token, gcenode, gcpserviceaccount, githubapp
- rule: {"required":true}

### spec.fleetDefaultMemberConfig.configmanagement.configSync.git.gcpServiceAccountEmail

`string | valueFrom`

The Google service account Config Sync reads as when secret_type is
"gcpserviceaccount": a GcpServiceAccount reference or a literal email.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.fleetDefaultMemberConfig.configmanagement.configSync.git.httpsProxy

`string`

An HTTPS proxy for reaching the repository.

### spec.fleetDefaultMemberConfig.configmanagement.configSync.git.policyDir

`string`

The directory within the repository to sync. Empty is the root.

### spec.fleetDefaultMemberConfig.configmanagement.configSync.git.syncBranch

`string`

The branch to sync. Empty is Google's default ("master").

### spec.fleetDefaultMemberConfig.configmanagement.configSync.git.syncRepo

`string` · required

The repository URL. Required.

- rule: {"required":true}

### spec.fleetDefaultMemberConfig.configmanagement.configSync.git.syncRev

`string`

A tag or commit to check out instead of the branch head.

### spec.fleetDefaultMemberConfig.configmanagement.configSync.git.syncWaitSecs

`int64`

Seconds between syncs. 0 takes Google's default (15).

- rule: {"int64":{"gte":"0"}}

### spec.fleetDefaultMemberConfig.configmanagement.configSync.oci

`GcpGkeFleetFeatureConfigSyncOci`

Sync from an OCI image in Artifact Registry.

### spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.secretType

`string` · required

How Config Sync authenticates to the registry:
  "none", "gcenode" (the node's service account), "gcpserviceaccount"
  (a Google service account through Workload Identity, see
  gcp_service_account_email), "k8sserviceaccount".
Required; Google matches it case-sensitively.

- rule: secret_type must be one of: none, gcenode, gcpserviceaccount, k8sserviceaccount
- rule: {"required":true}

### spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.gcpServiceAccountEmail

`string | valueFrom`

The Google service account Config Sync pulls as when secret_type is
"gcpserviceaccount": a GcpServiceAccount reference or a literal email.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.policyDir

`string`

The directory within the image to sync. Empty is the image root.

### spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.syncRepo

`string` · required

The image repository, e.g.
"us-docker.pkg.dev/my-project/configs/platform". Required.

- rule: {"required":true}

### spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.syncWaitSecs

`int64`

Seconds between syncs. 0 takes Google's default (15).

- rule: {"int64":{"gte":"0"}}

### spec.fleetDefaultMemberConfig.configmanagement.configSync.metricsGcpServiceAccountEmail

`string | valueFrom`

The service account Config Sync exports metrics to Cloud Monitoring
as (it needs roles/monitoring.metricWriter): a GcpServiceAccount
reference or a literal email.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.fleetDefaultMemberConfig.configmanagement.configSync.preventDrift

`bool` · optional (explicit presence)

true turns on Config Sync's admission webhook, which refuses manual
changes to synced objects; false turns it off. Unset takes Google's
default.

### spec.fleetDefaultMemberConfig.configmanagement.configSync.sourceFormat

`string`

The repository's layout: "unstructured" (any layout; the recommended
choice) or "hierarchy" (the namespace-directory layout). Empty takes
Google's default.

- rule: source_format must be unstructured or hierarchy

### spec.fleetDefaultMemberConfig.mesh

`GcpGkeFleetFeatureMesh`

Cloud Service Mesh for every cluster (feature "servicemesh").

### spec.fleetDefaultMemberConfig.mesh.management

`string` · required

"MANAGEMENT_AUTOMATIC" lets Google provision and upgrade the managed
mesh; "MANAGEMENT_MANUAL" leaves mesh components to you. Required: it
is the one mesh setting Google offers here.

- rule: management must be MANAGEMENT_AUTOMATIC or MANAGEMENT_MANUAL
- rule: {"required":true}

### spec.fleetDefaultMemberConfig.policycontroller

`GcpGkeFleetFeaturePolicyController`

Policy Controller for every cluster (feature "policycontroller").

### spec.fleetDefaultMemberConfig.policycontroller.version

`string`

The Policy Controller version to install. Empty takes the current
release.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig

`GcpGkeFleetFeaturePolicyControllerHubConfig` · required

How Policy Controller is installed and what it enforces. Required.

- rule: {"required":true}
- rule: each Policy Controller component appears at most once in deployment_configs

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.installSpec

`string` · required

The installation state:
  "INSTALL_SPEC_ENABLED"       -- installed and enforcing
  "INSTALL_SPEC_SUSPENDED"     -- installed with its webhooks off
  "INSTALL_SPEC_NOT_INSTALLED" -- uninstalled
  "INSTALL_SPEC_DETACHED"      -- Google stops reconciling (break-glass)
Required.

- rule: install_spec must be INSTALL_SPEC_ENABLED, INSTALL_SPEC_SUSPENDED, INSTALL_SPEC_NOT_INSTALLED, or INSTALL_SPEC_DETACHED
- rule: {"required":true}

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.auditIntervalSeconds

`int64` · optional (explicit presence)

Seconds between audit scans; 0 turns auditing off. Unset takes
Google's default.

- rule: {"int64":{"gte":"0"}}

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.constraintViolationLimit

`int64` · optional (explicit presence)

How many violations a constraint stores. Unset takes Google's default
(20).

- rule: {"int64":{"gte":"0"}}

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.exemptableNamespaces

`[]string`

Namespaces Policy Controller never checks; they need not exist yet.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.logDeniesEnabled

`bool`

Log every denial and dry-run failure.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.mutationEnabled

`bool`

Allow mutation policies, which change objects at admission.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.referentialRulesEnabled

`bool`

Allow constraint templates that read objects other than the one being
checked (referential rules).

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.monitoring

`GcpGkeFleetFeaturePolicyControllerMonitoring`

Where Policy Controller exports metrics. Unset takes Google's default;
set with no backends to turn metrics export off.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.monitoring.backends

`[]string`

"PROMETHEUS" and/or "CLOUD_MONITORING". Empty turns export off.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["PROMETHEUS","CLOUD_MONITORING"]}}}}

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs

`[]GcpGkeFleetFeaturePolicyControllerDeploymentConfig`

Sizing and placement for Policy Controller's components.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].component

`string` · required

The component: "admission", "audit", or "mutation". Required.

- rule: component must be admission, audit, or mutation
- rule: {"required":true}

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].replicaCount

`int64` · optional (explicit presence)

Pod replicas. Unset takes Google's default.

- rule: {"int64":{"gte":"0"}}

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podAffinity

`string`

"ANTI_AFFINITY" spreads replicas across nodes; "NO_AFFINITY" does not.
Empty takes Google's default.

- rule: pod_affinity must be NO_AFFINITY or ANTI_AFFINITY

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources

`GcpGkeFleetFeaturePolicyControllerContainerResources`

Container resources.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits

`GcpGkeFleetFeaturePolicyControllerResourceList`

The most the container may use.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits.cpu

`string`

CPU.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits.memory

`string`

Memory.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests

`GcpGkeFleetFeaturePolicyControllerResourceList`

What the scheduler reserves for the container.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests.cpu

`string`

CPU.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests.memory

`string`

Memory.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations

`[]GcpGkeFleetFeaturePolicyControllerToleration`

Tolerations, so the component can run on tainted nodes.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].effect

`string`

The taint effect matched, e.g. "NoSchedule".

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].key

`string`

The taint key matched.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].operator

`string`

"Equal" or "Exists".

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].value

`string`

The taint value matched (with "Equal").

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent

`GcpGkeFleetFeaturePolicyControllerPolicyContent`

The policies installed.

- rule: each bundle appears at most once

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.bundles

`[]GcpGkeFleetFeaturePolicyControllerBundle`

Google's policy bundles to install, e.g. "cis-k8s-v1.5.1",
"pss-baseline-v2022", "policy-essentials-v2022".

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.bundles[].bundle

`string` · required

The bundle's name. Required.

- rule: {"required":true}

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.bundles[].exemptedNamespaces

`[]string`

Namespaces the bundle's constraints skip.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.templateLibrary

`GcpGkeFleetFeaturePolicyControllerTemplateLibrary`

The constraint template library.

### spec.fleetDefaultMemberConfig.policycontroller.policyControllerHubConfig.policyContent.templateLibrary.installation

`string`

"ALL" installs Google's whole template library; "NOT_INSTALLED"
installs none. Empty takes Google's default.

- rule: installation must be ALL or NOT_INSTALLED

### spec.membershipConfigs

`[]GcpGkeFleetFeatureMembershipConfig`

Per-cluster settings for "configmanagement", "servicemesh", or
"policycontroller", each overriding the fleet-wide default for one
cluster.

- rule: a membership_configs entry carries exactly one of configmanagement, mesh, or policycontroller

### spec.membershipConfigs[].membership

`string | valueFrom` · required

The cluster's fleet membership, by full name: a GcpGkeFleetMembership
reference, a GcpGkeCluster reference (its fleet_membership output), or
the literal name. The membership must be in this feature's fleet
project; both modules derive the membership ID and location from it.
Immutable.

- references: GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`)
- rule: membership must be a full membership name: projects/{project}/locations/{location}/memberships/{id}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeFleetMembership, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.membershipConfigs[].configmanagement

`GcpGkeFleetFeatureMembershipConfigManagement`

Config Sync for this cluster (feature "configmanagement").

### spec.membershipConfigs[].configmanagement.management

`string`

Upgrades: "MANAGEMENT_AUTOMATIC" or "MANAGEMENT_MANUAL" (pinned to
`version`). Empty takes Google's default.

- rule: management must be MANAGEMENT_AUTOMATIC or MANAGEMENT_MANUAL

### spec.membershipConfigs[].configmanagement.version

`string`

The Config Sync version to install on this cluster.

### spec.membershipConfigs[].configmanagement.configSync

`GcpGkeFleetFeatureMembershipConfigSync`

What this cluster syncs and how.

- rule: config_sync syncs from git or oci, not both

### spec.membershipConfigs[].configmanagement.configSync.enabled

`bool` · optional (explicit presence)

true installs Config Sync and applies these settings; false removes it
and ignores the rest. Unset installs it exactly when git or oci is set.

### spec.membershipConfigs[].configmanagement.configSync.git

`GcpGkeFleetFeatureConfigSyncGit`

Sync from a Git repository.

### spec.membershipConfigs[].configmanagement.configSync.git.secretType

`string` · required

How Config Sync authenticates to the repository:
  "none" (public repository), "ssh", "cookiefile", "token",
  "gcenode" (the node's service account), "gcpserviceaccount" (a
  Google service account through Workload Identity, see
  gcp_service_account_email), "githubapp".
Required; Google matches it case-sensitively.

- rule: secret_type must be one of: none, ssh, cookiefile, token, gcenode, gcpserviceaccount, githubapp
- rule: {"required":true}

### spec.membershipConfigs[].configmanagement.configSync.git.gcpServiceAccountEmail

`string | valueFrom`

The Google service account Config Sync reads as when secret_type is
"gcpserviceaccount": a GcpServiceAccount reference or a literal email.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.membershipConfigs[].configmanagement.configSync.git.httpsProxy

`string`

An HTTPS proxy for reaching the repository.

### spec.membershipConfigs[].configmanagement.configSync.git.policyDir

`string`

The directory within the repository to sync. Empty is the root.

### spec.membershipConfigs[].configmanagement.configSync.git.syncBranch

`string`

The branch to sync. Empty is Google's default ("master").

### spec.membershipConfigs[].configmanagement.configSync.git.syncRepo

`string` · required

The repository URL. Required.

- rule: {"required":true}

### spec.membershipConfigs[].configmanagement.configSync.git.syncRev

`string`

A tag or commit to check out instead of the branch head.

### spec.membershipConfigs[].configmanagement.configSync.git.syncWaitSecs

`int64`

Seconds between syncs. 0 takes Google's default (15).

- rule: {"int64":{"gte":"0"}}

### spec.membershipConfigs[].configmanagement.configSync.oci

`GcpGkeFleetFeatureConfigSyncOci`

Sync from an OCI image in Artifact Registry.

### spec.membershipConfigs[].configmanagement.configSync.oci.secretType

`string` · required

How Config Sync authenticates to the registry:
  "none", "gcenode" (the node's service account), "gcpserviceaccount"
  (a Google service account through Workload Identity, see
  gcp_service_account_email), "k8sserviceaccount".
Required; Google matches it case-sensitively.

- rule: secret_type must be one of: none, gcenode, gcpserviceaccount, k8sserviceaccount
- rule: {"required":true}

### spec.membershipConfigs[].configmanagement.configSync.oci.gcpServiceAccountEmail

`string | valueFrom`

The Google service account Config Sync pulls as when secret_type is
"gcpserviceaccount": a GcpServiceAccount reference or a literal email.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.membershipConfigs[].configmanagement.configSync.oci.policyDir

`string`

The directory within the image to sync. Empty is the image root.

### spec.membershipConfigs[].configmanagement.configSync.oci.syncRepo

`string` · required

The image repository, e.g.
"us-docker.pkg.dev/my-project/configs/platform". Required.

- rule: {"required":true}

### spec.membershipConfigs[].configmanagement.configSync.oci.syncWaitSecs

`int64`

Seconds between syncs. 0 takes Google's default (15).

- rule: {"int64":{"gte":"0"}}

### spec.membershipConfigs[].configmanagement.configSync.metricsGcpServiceAccountEmail

`string | valueFrom`

The service account Config Sync exports metrics as: a
GcpServiceAccount reference or a literal email.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.membershipConfigs[].configmanagement.configSync.preventDrift

`bool` · optional (explicit presence)

true turns on the admission webhook that refuses manual changes to
synced objects. Unset takes Google's default.

### spec.membershipConfigs[].configmanagement.configSync.sourceFormat

`string`

The repository's layout: "unstructured" or "hierarchy".

- rule: source_format must be unstructured or hierarchy

### spec.membershipConfigs[].configmanagement.configSync.stopSyncing

`bool` · optional (explicit presence)

true pauses syncing on this cluster (a break-glass lever during an
incident); Config Sync stays installed. Unset or false keeps syncing.

### spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides

`[]GcpGkeFleetFeatureConfigSyncDeploymentOverride`

Resource overrides for Config Sync's own Deployments on this cluster,
e.g. more memory for the reconciler of a large repository.

### spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].deploymentName

`string`

The Deployment's name, e.g. "root-reconciler".

### spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].deploymentNamespace

`string`

The Deployment's namespace, e.g. "config-management-system".

### spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers

`[]GcpGkeFleetFeatureConfigSyncContainerOverride`

Per-container resource overrides.

### spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].containerName

`string`

The container's name, e.g. "reconciler".

### spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].cpuLimit

`string`

CPU limit.

### spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].cpuRequest

`string`

CPU request.

### spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].memoryLimit

`string`

Memory limit.

### spec.membershipConfigs[].configmanagement.configSync.deploymentOverrides[].containers[].memoryRequest

`string`

Memory request.

### spec.membershipConfigs[].mesh

`GcpGkeFleetFeatureMesh`

Cloud Service Mesh for this cluster (feature "servicemesh").

### spec.membershipConfigs[].mesh.management

`string` · required

"MANAGEMENT_AUTOMATIC" lets Google provision and upgrade the managed
mesh; "MANAGEMENT_MANUAL" leaves mesh components to you. Required: it
is the one mesh setting Google offers here.

- rule: management must be MANAGEMENT_AUTOMATIC or MANAGEMENT_MANUAL
- rule: {"required":true}

### spec.membershipConfigs[].policycontroller

`GcpGkeFleetFeaturePolicyController`

Policy Controller for this cluster (feature "policycontroller").

### spec.membershipConfigs[].policycontroller.version

`string`

The Policy Controller version to install. Empty takes the current
release.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig

`GcpGkeFleetFeaturePolicyControllerHubConfig` · required

How Policy Controller is installed and what it enforces. Required.

- rule: {"required":true}
- rule: each Policy Controller component appears at most once in deployment_configs

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.installSpec

`string` · required

The installation state:
  "INSTALL_SPEC_ENABLED"       -- installed and enforcing
  "INSTALL_SPEC_SUSPENDED"     -- installed with its webhooks off
  "INSTALL_SPEC_NOT_INSTALLED" -- uninstalled
  "INSTALL_SPEC_DETACHED"      -- Google stops reconciling (break-glass)
Required.

- rule: install_spec must be INSTALL_SPEC_ENABLED, INSTALL_SPEC_SUSPENDED, INSTALL_SPEC_NOT_INSTALLED, or INSTALL_SPEC_DETACHED
- rule: {"required":true}

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.auditIntervalSeconds

`int64` · optional (explicit presence)

Seconds between audit scans; 0 turns auditing off. Unset takes
Google's default.

- rule: {"int64":{"gte":"0"}}

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.constraintViolationLimit

`int64` · optional (explicit presence)

How many violations a constraint stores. Unset takes Google's default
(20).

- rule: {"int64":{"gte":"0"}}

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.exemptableNamespaces

`[]string`

Namespaces Policy Controller never checks; they need not exist yet.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.logDeniesEnabled

`bool`

Log every denial and dry-run failure.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.mutationEnabled

`bool`

Allow mutation policies, which change objects at admission.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.referentialRulesEnabled

`bool`

Allow constraint templates that read objects other than the one being
checked (referential rules).

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.monitoring

`GcpGkeFleetFeaturePolicyControllerMonitoring`

Where Policy Controller exports metrics. Unset takes Google's default;
set with no backends to turn metrics export off.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.monitoring.backends

`[]string`

"PROMETHEUS" and/or "CLOUD_MONITORING". Empty turns export off.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["PROMETHEUS","CLOUD_MONITORING"]}}}}

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs

`[]GcpGkeFleetFeaturePolicyControllerDeploymentConfig`

Sizing and placement for Policy Controller's components.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].component

`string` · required

The component: "admission", "audit", or "mutation". Required.

- rule: component must be admission, audit, or mutation
- rule: {"required":true}

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].replicaCount

`int64` · optional (explicit presence)

Pod replicas. Unset takes Google's default.

- rule: {"int64":{"gte":"0"}}

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podAffinity

`string`

"ANTI_AFFINITY" spreads replicas across nodes; "NO_AFFINITY" does not.
Empty takes Google's default.

- rule: pod_affinity must be NO_AFFINITY or ANTI_AFFINITY

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources

`GcpGkeFleetFeaturePolicyControllerContainerResources`

Container resources.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits

`GcpGkeFleetFeaturePolicyControllerResourceList`

The most the container may use.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits.cpu

`string`

CPU.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.limits.memory

`string`

Memory.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests

`GcpGkeFleetFeaturePolicyControllerResourceList`

What the scheduler reserves for the container.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests.cpu

`string`

CPU.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].containerResources.requests.memory

`string`

Memory.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations

`[]GcpGkeFleetFeaturePolicyControllerToleration`

Tolerations, so the component can run on tainted nodes.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].effect

`string`

The taint effect matched, e.g. "NoSchedule".

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].key

`string`

The taint key matched.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].operator

`string`

"Equal" or "Exists".

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.deploymentConfigs[].podTolerations[].value

`string`

The taint value matched (with "Equal").

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent

`GcpGkeFleetFeaturePolicyControllerPolicyContent`

The policies installed.

- rule: each bundle appears at most once

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.bundles

`[]GcpGkeFleetFeaturePolicyControllerBundle`

Google's policy bundles to install, e.g. "cis-k8s-v1.5.1",
"pss-baseline-v2022", "policy-essentials-v2022".

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.bundles[].bundle

`string` · required

The bundle's name. Required.

- rule: {"required":true}

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.bundles[].exemptedNamespaces

`[]string`

Namespaces the bundle's constraints skip.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.templateLibrary

`GcpGkeFleetFeaturePolicyControllerTemplateLibrary`

The constraint template library.

### spec.membershipConfigs[].policycontroller.policyControllerHubConfig.policyContent.templateLibrary.installation

`string`

"ALL" installs Google's whole template library; "NOT_INSTALLED"
installs none. Empty takes Google's default.

- rule: installation must be ALL or NOT_INSTALLED

### spec.deletionPolicy

`string`

What destroy does, for the feature and its per-cluster settings:
  "" / "DELETE" -- the feature is turned off and the per-cluster
                   settings removed ("rbacrolebindingactuation" only
                   empties its allowlist)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays on in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.multiclusteringress_on_its_feature`: multiclusteringress settings apply only to feature multiclusteringress
- `spec.fleetobservability_on_its_feature`: fleetobservability settings apply only to feature fleetobservability
- `spec.clusterupgrade_on_its_feature`: clusterupgrade settings apply only to feature clusterupgrade
- `spec.rbacrolebindingactuation_on_its_feature`: rbacrolebindingactuation settings apply only to feature rbacrolebindingactuation
- `spec.workloadidentity_on_its_feature`: workloadidentity settings apply only to feature workloadidentity
- `spec.default_configmanagement_on_its_feature`: fleet_default_member_config.configmanagement applies only to feature configmanagement
- `spec.default_mesh_on_its_feature`: fleet_default_member_config.mesh applies only to feature servicemesh
- `spec.default_policycontroller_on_its_feature`: fleet_default_member_config.policycontroller applies only to feature policycontroller
- `spec.membership_configs_on_member_features`: membership_configs apply only to features configmanagement, servicemesh, and policycontroller
- `spec.membership_config_block_matches_feature`: each membership_configs entry carries the block of this feature: configmanagement, mesh (servicemesh), or policycontroller
- `spec.unique_membership_configs`: a membership appears at most once in membership_configs

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGkeFleetFeature, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/features/{feature}. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpGkeFleet | `status.outputs.project_id` |
| `spec.multiclusteringress.configMembership` | GcpGkeFleetMembership | `status.outputs.name` |
| `spec.multiclusteringress.configMembership` | GcpGkeCluster | `status.outputs.fleet_membership` |
| `spec.clusterupgrade.upstreamFleets` | GcpGkeFleet | `status.outputs.project_id` |
| `spec.workloadidentity.scopeTenancyPool` | GcpWorkloadIdentityPool | `status.outputs.name` |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.git.gcpServiceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.oci.gcpServiceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.fleetDefaultMemberConfig.configmanagement.configSync.metricsGcpServiceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.membershipConfigs[].membership` | GcpGkeFleetMembership | `status.outputs.name` |
| `spec.membershipConfigs[].membership` | GcpGkeCluster | `status.outputs.fleet_membership` |
| `spec.membershipConfigs[].configmanagement.configSync.git.gcpServiceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.membershipConfigs[].configmanagement.configSync.oci.gcpServiceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.membershipConfigs[].configmanagement.configSync.metricsGcpServiceAccountEmail` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)
