# GcpProject

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpProjectSpec creates one Google Cloud project — the Layer-0 container
every other GCP resource lives in. It attaches the project to the
resource hierarchy (organization or folder), links a billing account,
applies labels and resource-manager tags, controls the default-network
behavior, pre-enables Cloud APIs, and sets the deletion policy.

IAM grants on the project are deliberately NOT bundled here — model them
as first-class GcpProjectIamMember resources, one additive grant each.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpProject
metadata:
  name: test-project
spec:
  projectId: test-project-123456
  parentType: organization
  parentId: "123456789012"
  billingAccountId: 0123AB-4567CD-89EFGH
  enabledApis:
    - compute.googleapis.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.parentType` | `enum` |  |  |  |
| `spec.parentId` | `string` |  |  |  |
| `spec.billingAccountId` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.autoCreateNetwork` | `bool` |  | `false` |  |
| `spec.enabledApis` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |

## Field Details

### spec.projectId

`string` · required

Unique project ID for the GCP project. Must be 6-30 characters long,
contain only lowercase letters, digits, and hyphens, must start with
a letter, and cannot end with a hyphen. Globally unique across all of
GCP and IMMUTABLE — deleted project IDs stay reserved for up to 30
days, so they cannot be reused quickly.

- rule: {"required":true,"string":{"minLen":"6","maxLen":"30","pattern":"^[a-z][a-z0-9-]*[a-z0-9]$"}}

### spec.displayName

`string`

Human-readable display name shown in the console (4-30 characters).
Mutable, unlike project_id. If not specified, defaults to
metadata.name.

### spec.parentType

`enum`

The type of parent node the project is created under. Changing the
parent migrates the project within the hierarchy. May be left empty
when folder_id names the parent.

Allowed values (use exactly as shown):

- `gcp_project_parent_type_unspecified`
- `organization`
- `folder`

### spec.parentId

`string`

Organization ID or Folder ID (numeric string) matching parent_type.
For a folder declared in the same chart, prefer folder_id (a
reference) and leave this empty.

- rule: parent_id must be the numeric organization/folder ID

### spec.billingAccountId

`string`

Billing account ID in the form "0123AB-4567CD-89EFGH".
Strongly recommended for any project that will use billable services.
The deploying identity needs roles/billing.user on the account.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[A-Z0-9]{6}-[A-Z0-9]{6}-[A-Z0-9]{6}$"}}

### spec.labels

`map<string, string>`

User labels merged onto the project beneath the platform's attribution
labels (platform keys win on conflicts). Project labels are the
primary cost-allocation dimension in billing exports.
Keys/values: lowercase letters, digits, underscores, hyphens.

### spec.tags

`map<string, string>`

Resource Manager tags bound to the project at CREATE TIME only
(tagKeys/{id} -> tagValues/{id}, the `name` outputs of GcpTagKey and
GcpTagValue). Tags drive org policies and IAM conditions. Changing this
after creation recreates the project — for tags on an existing project,
bind tag values with GcpTagBinding instead, which attaches and detaches
without touching the project.

### spec.autoCreateNetwork

`bool` · optional (explicit presence)

Whether GCP auto-creates the "default" VPC network in the new project.
Defaults to false: deleting the auto-created network is a standard
security-hardening step, and explicit GcpVpcNetwork resources are the
composable path. Note the project still needs one network slot of
quota available even when false (the network exists momentarily).

- default: `false`

### spec.enabledApis

`[]string`

List of Cloud APIs to enable at project creation
(e.g. "compute.googleapis.com"). Individual component kinds also
enable the APIs they need, so this is a convenience for pre-warming a
known set. Each entry must end with ".googleapis.com".

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-z0-9]+\\.googleapis\\.com$"}}}}

### spec.deletionPolicy

`string`

What destroying this resource does to the project:
  DELETE (default): the project is shut down (30-day pending-deletion
    window during which it can be restored).
  PREVENT: destroy fails — protection for shared foundation projects.
  ABANDON: the resource is removed from state and the project lives
    on unmanaged — the safe hand-off when ownership moves elsewhere.

- rule: deletion_policy must be DELETE, PREVENT, or ABANDON

### spec.folderId

`string | valueFrom`

The folder the project lives in, by reference: a GcpFolder resource
(its folder_id output) or the folder's numeric ID as a literal. This is
how a chart places a project inside a folder it also declares -- the
project waits for the folder to exist. When set it IS the parent:
leave parent_id empty and parent_type empty (or `folder`). Changing it
moves the project into the new folder in place; nothing is recreated,
but the IAM and organization policies inherited from the old folder
stop applying and the new folder's start.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

## Validation Rules

- `folder_id_is_the_parent`: when folder_id is set it IS the parent: leave parent_id empty and parent_type either empty or folder

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpProject, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Display name of the project (mirrors spec.name). |
| `status.outputs.project_id` | `string` | Immutable project ID (mirrors spec.project_id). |
| `status.outputs.project_number` | `string` | Numeric project number assigned by Google. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.folderId` | GcpFolder | `status.outputs.folder_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpAddress | `spec.projectId` | `status.outputs.project_id` |
| GcpAlloydbCluster | `spec.projectId` | `status.outputs.project_id` |
| GcpAlloydbInstance | `spec.projectId` | `status.outputs.project_id` |
| GcpAlloydbUser | `spec.projectId` | `status.outputs.project_id` |
| GcpApiKey | `spec.projectId` | `status.outputs.project_id` |
| GcpArtifactRegistryRepo | `spec.projectId` | `status.outputs.project_id` |
| GcpBackendBucket | `spec.projectId` | `status.outputs.project_id` |
| GcpBackendService | `spec.projectId` | `status.outputs.project_id` |
| GcpBigQueryDataset | `spec.projectId` | `status.outputs.project_id` |
| GcpBigQueryTable | `spec.projectId` | `status.outputs.project_id` |
| GcpBigtableInstance | `spec.projectId` | `status.outputs.project_id` |
| GcpBigtableTable | `spec.projectId` | `status.outputs.project_id` |
| GcpBillingBudget | `spec.budgetFilter.projects` | `status.outputs.project_number` |
| GcpCertManagerCert | `spec.projectId` | `status.outputs.project_id` |
| GcpCertManagerDnsAuthorization | `spec.projectId` | `status.outputs.project_id` |
| GcpCertificateMap | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudArmorPolicy | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudComposerEnvironment | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudComposerUserWorkloadsConfigMap | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudComposerUserWorkloadsSecret | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudFunction | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudRun | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudRunDomainMapping | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudRunJob | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudRunWorkerPool | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudSchedulerJob | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudSql | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudSqlDatabase | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudSqlUser | `spec.projectId` | `status.outputs.project_id` |
| GcpCloudTasksQueue | `spec.projectId` | `status.outputs.project_id` |
| GcpColabRuntime | `spec.projectId` | `status.outputs.project_id` |
| GcpColabRuntimeTemplate | `spec.projectId` | `status.outputs.project_id` |
| GcpColabSchedule | `spec.projectId` | `status.outputs.project_id` |
| GcpColabSchedule | `spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | `status.outputs.project_id` |
| GcpComputeDisk | `spec.projectId` | `status.outputs.project_id` |
| GcpComputeInstance | `spec.projectId` | `status.outputs.project_id` |
| GcpComputeMig | `spec.projectId` | `status.outputs.project_id` |
| GcpDataprocAutoscalingPolicy | `spec.projectId` | `status.outputs.project_id` |
| GcpDataprocCluster | `spec.projectId` | `status.outputs.project_id` |
| GcpDialogflowCxAgent | `spec.projectId` | `status.outputs.project_id` |
| GcpDialogflowCxSecuritySettings | `spec.projectId` | `status.outputs.project_id` |
| GcpDnsRecord | `spec.projectId` | `status.outputs.project_id` |
| GcpDnsRecord | `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].project` | `status.outputs.project_id` |
| GcpDnsRecord | `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].project` | `status.outputs.project_id` |
| GcpDnsRecord | `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].project` | `status.outputs.project_id` |
| GcpDnsRecord | `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].project` | `status.outputs.project_id` |
| GcpDnsZone | `spec.projectId` | `status.outputs.project_id` |
| GcpDocumentAiProcessor | `spec.projectId` | `status.outputs.project_id` |
| GcpEventarcMessageBus | `spec.projectId` | `status.outputs.project_id` |
| GcpEventarcTrigger | `spec.projectId` | `status.outputs.project_id` |
| GcpFilestoreInstance | `spec.projectId` | `status.outputs.project_id` |
| GcpFilestoreInstance | `spec.networkConfig.pscEndpointProject` | `status.outputs.project_id` |
| GcpFirebaseProject | `spec.projectId` | `status.outputs.project_id` |
| GcpFirestoreBackupSchedule | `spec.projectId` | `status.outputs.project_id` |
| GcpFirestoreDatabase | `spec.projectId` | `status.outputs.project_id` |
| GcpFirestoreIndex | `spec.projectId` | `status.outputs.project_id` |
| GcpFirewallRule | `spec.projectId` | `status.outputs.project_id` |
| GcpGcsBucket | `spec.projectId` | `status.outputs.project_id` |
| GcpGkeCluster | `spec.projectId` | `status.outputs.project_id` |
| GcpGkeNodePool | `spec.projectId` | `status.outputs.project_id` |
| GcpGkeWorkloadIdentityBinding | `spec.projectId` | `status.outputs.project_id` |
| GcpGlobalAddress | `spec.projectId` | `status.outputs.project_id` |
| GcpGlobalForwardingRule | `spec.projectId` | `status.outputs.project_id` |
| GcpHaVpnConnection | `spec.projectId` | `status.outputs.project_id` |
| GcpHaVpnGateway | `spec.projectId` | `status.outputs.project_id` |
| GcpHealthCheck | `spec.projectId` | `status.outputs.project_id` |
| GcpIamCustomRole | `spec.projectId` | `status.outputs.project_id` |
| GcpIamDenyPolicy | `spec.parent.projectId` | `status.outputs.project_id` |
| GcpIamOauthClient | `spec.projectId` | `status.outputs.project_id` |
| GcpIdentityPlatformConfig | `spec.projectId` | `status.outputs.project_id` |
| GcpIdentityPlatformTenant | `spec.projectId` | `status.outputs.project_id` |
| GcpKmsKeyRing | `spec.projectId` | `status.outputs.project_id` |
| GcpLogBucket | `spec.scope.projectId` | `status.outputs.project_id` |
| GcpLogMetric | `spec.projectId` | `status.outputs.project_id` |
| GcpLoggingSink | `spec.scope.projectId` | `status.outputs.project_id` |
| GcpManagedSslCertificate | `spec.projectId` | `status.outputs.project_id` |
| GcpMemorystoreInstance | `spec.projectId` | `status.outputs.project_id` |
| GcpMemorystoreInstance | `spec.pscAutoConnections[].projectId` | `status.outputs.project_id` |
| GcpModelArmorFloorSetting | `spec.scope.projectId` | `status.outputs.project_id` |
| GcpModelArmorTemplate | `spec.projectId` | `status.outputs.project_id` |
| GcpMonitoringAlertPolicy | `spec.projectId` | `status.outputs.project_id` |
| GcpMonitoringDashboard | `spec.projectId` | `status.outputs.project_id` |
| GcpMonitoringNotificationChannel | `spec.projectId` | `status.outputs.project_id` |
| GcpMonitoringSlo | `spec.projectId` | `status.outputs.project_id` |
| GcpMonitoringUptimeCheck | `spec.projectId` | `status.outputs.project_id` |
| GcpNetworkEndpointGroup | `spec.projectId` | `status.outputs.project_id` |
| GcpNetworkFirewallPolicy | `spec.projectId` | `status.outputs.project_id` |
| GcpOrgPolicy | `spec.scope.projectId` | `status.outputs.project_id` |
| GcpPlantonRunner | `spec.projectId` | `status.outputs.project_id` |
| GcpProjectIamMember | `spec.projectId` | `status.outputs.project_id` |
| GcpPscServiceAttachment | `spec.projectId` | `status.outputs.project_id` |
| GcpPscServiceAttachment | `spec.consumerAcceptLists[].projectId` | `status.outputs.project_id` |
| GcpPscServiceAttachment | `spec.consumerRejectLists` | `status.outputs.project_id` |
| GcpPubSubSchema | `spec.projectId` | `status.outputs.project_id` |
| GcpPubSubSubscription | `spec.projectId` | `status.outputs.project_id` |
| GcpPubSubTopic | `spec.projectId` | `status.outputs.project_id` |
| GcpRedisCluster | `spec.projectId` | `status.outputs.project_id` |
| GcpRedisClusterEndpointSet | `spec.projectId` | `status.outputs.project_id` |
| GcpRedisClusterEndpointSet | `spec.endpoints[].connections[].projectId` | `status.outputs.project_id` |
| GcpRedisInstance | `spec.projectId` | `status.outputs.project_id` |
| GcpRegionNetworkEndpointGroup | `spec.projectId` | `status.outputs.project_id` |
| GcpRouterNat | `spec.projectId` | `status.outputs.project_id` |
| GcpSecretManagerSecret | `spec.projectId` | `status.outputs.project_id` |
| GcpServerlessVpcConnector | `spec.projectId` | `status.outputs.project_id` |
| GcpServiceAccount | `spec.projectId` | `status.outputs.project_id` |
| GcpServiceConnectionPolicy | `spec.projectId` | `status.outputs.project_id` |
| GcpServiceNetworkingConnection | `spec.projectId` | `status.outputs.project_id` |
| GcpSharedVpcHost | `spec.projectId` | `status.outputs.project_id` |
| GcpSharedVpcServiceProject | `spec.serviceProjectId` | `status.outputs.project_id` |
| GcpSpannerBackupSchedule | `spec.projectId` | `status.outputs.project_id` |
| GcpSpannerDatabase | `spec.projectId` | `status.outputs.project_id` |
| GcpSpannerInstance | `spec.projectId` | `status.outputs.project_id` |
| GcpSslCertificate | `spec.projectId` | `status.outputs.project_id` |
| GcpSslPolicy | `spec.projectId` | `status.outputs.project_id` |
| GcpSubnetwork | `spec.projectId` | `status.outputs.project_id` |
| GcpTagBinding | `spec.parent.projectId` | `status.outputs.project_number` |
| GcpTagKey | `spec.parent.projectId` | `status.outputs.project_id` |
| GcpTargetHttpProxy | `spec.projectId` | `status.outputs.project_id` |
| GcpTargetHttpsProxy | `spec.projectId` | `status.outputs.project_id` |
| GcpTpuQueuedResource | `spec.projectId` | `status.outputs.project_id` |
| GcpTpuVm | `spec.projectId` | `status.outputs.project_id` |
| GcpUrlMap | `spec.projectId` | `status.outputs.project_id` |
| GcpVectorSearchCollection | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiAgentEngine | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiAgentEngine | `spec.spec.deploymentSpec.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | `status.outputs.project_id` |
| GcpVertexAiDataset | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiEndpoint | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiEndpoint | `spec.privateServiceConnectConfig.pscAutomationConfigs[].projectId` | `status.outputs.project_id` |
| GcpVertexAiFeatureGroup | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiFeatureOnlineStore | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiFeatureOnlineStore | `spec.featureViews[].featureRegistrySource.projectNumber` | `status.outputs.project_number` |
| GcpVertexAiIndex | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiIndexEndpoint | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiIndexEndpoint | `spec.privateServiceConnectConfig.pscAutomationConfigs[].projectId` | `status.outputs.project_id` |
| GcpVertexAiModelGardenDeployment | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiModelGardenDeployment | `spec.endpointConfig.privateServiceConnectConfig.projectAllowlist` | `status.outputs.project_id` |
| GcpVertexAiModelGardenDeployment | `spec.endpointConfig.privateServiceConnectConfig.pscAutomationConfig.projectId` | `status.outputs.project_id` |
| GcpVertexAiNotebook | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiPersistentResource | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiPersistentResource | `spec.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | `status.outputs.project_id` |
| GcpVertexAiRagEngineConfig | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiSearchDataConnector | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiSearchDataStore | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiSearchEngine | `spec.projectId` | `status.outputs.project_id` |
| GcpVertexAiTensorboard | `spec.projectId` | `status.outputs.project_id` |
| GcpVpcNetwork | `spec.projectId` | `status.outputs.project_id` |
| GcpWorkflow | `spec.projectId` | `status.outputs.project_id` |
| GcpWorkloadIdentityPool | `spec.projectId` | `status.outputs.project_id` |
| GcpWorkloadIdentityPoolProvider | `spec.projectId` | `status.outputs.project_id` |
| KubernetesClusterSecretStore | `spec.config.gcpSecretManager.projectId` | `status.outputs.project_id` |
| KubernetesExternalDns | `spec.googleCloudDns.project` | `status.outputs.project_id` |
| KubernetesOpenBao | `spec.autoUnseal.gcpKms.project` | `status.outputs.project_id` |
| KubernetesPlantonPlatform | `spec.vault.autoUnseal.gcpKms.project` | `status.outputs.project_id` |
| KubernetesSecretStore | `spec.config.gcpSecretManager.projectId` | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
