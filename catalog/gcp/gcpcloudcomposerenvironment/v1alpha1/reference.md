# GcpCloudComposerEnvironment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCloudComposerEnvironmentSpec defines the configuration for a Google Cloud
Composer environment -- a managed Apache Airflow service for authoring,
scheduling, and monitoring data pipelines.

Cloud Composer provisions and manages the underlying GKE cluster, Cloud SQL
metadata database, web server, and Cloud Storage DAG bucket. This component
targets Composer 2.x and 3 (Composer 1.x fields are excluded as that version
is deprecated by Google).

Important behavioral notes:

  - environment_name, region, node networking, encryption, and the private
    environment configuration are immutable after creation. The workloads
    sizing, environment size, resilience mode, software configuration
    (packages, overrides, env vars), maintenance window, web-server access
    control, and labels update in place.

  - Environment creation takes 25-45 minutes: Composer assembles a GKE
    cluster, Cloud SQL database, and web server behind the scenes.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudComposerEnvironment
metadata:
  name: test-composer
spec:
  # GCP project for the environment. Replace with your project ID.
  # Omit entirely to ride the provider's default project.
  projectId:
    value: my-gcp-project-123

  region: us-central1

  # Composer 3 image — supports the DAG processor workload below.
  softwareConfig:
    imageVersion: composer-3-airflow-2.10.5 # replace with a current image version
    airflowConfigOverrides:
      core-dags_are_paused_at_creation: "True"
    pypiPackages:
      apache-airflow-providers-google: ""
    envVariables:
      DATA_ENV: test

  environmentSize: ENVIRONMENT_SIZE_SMALL

  # All five Airflow components. The triggerer's cpu/memory/count are all
  # required by the API when the block is present.
  workloadsConfig:
    scheduler:
      cpu: 1.0
      memoryGb: 4.0
      storageGb: 1.0
      count: 1
    webServer:
      cpu: 1.0
      memoryGb: 2.0
      storageGb: 1.0
    worker:
      cpu: 1.0
      memoryGb: 4.0
      storageGb: 1.0
      minCount: 1
      maxCount: 3
    triggerer:
      cpu: 0.5
      memoryGb: 1.0
      count: 1
    dagProcessor:
      cpu: 1.0
      memoryGb: 2.0
      storageGb: 1.0
      count: 1

  # Composer requires at least a 12-hour window.
  maintenanceWindow:
    startTime: "2026-01-01T00:00:00Z"
    endTime: "2026-01-01T12:00:00Z"
    recurrence: "FREQ=WEEKLY;BYDAY=SA,SU"

  # Only these ranges reach the Airflow web UI.
  webServerNetworkAccessControl:
    allowedIpRanges:
      - value: "10.0.0.0/8"
        description: Internal network
      - value: "203.0.113.0/24" # replace with your office/VPN range
        description: Office network

  labels:
    team: data-platform
    purpose: smoke-test

  # Explicit DELETE (the provider default) proves the round-trip; the
  # auto-created DAG bucket survives the destroy either way.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.environmentName` | `string` |  |  |  |
| `spec.nodeConfig` | `GcpCloudComposerNodeConfig` |  |  |  |
| `spec.nodeConfig.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.nodeConfig.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.nodeConfig.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.nodeConfig.tags` | `[]string` |  |  |  |
| `spec.nodeConfig.composerNetworkAttachment` | `string` |  |  |  |
| `spec.nodeConfig.composerInternalIpv4CidrBlock` | `string` |  |  |  |
| `spec.nodeConfig.enableIpMasqAgent` | `bool` |  |  |  |
| `spec.nodeConfig.ipAllocationPolicy` | `GcpCloudComposerIpAllocationPolicy` |  |  |  |
| `spec.nodeConfig.ipAllocationPolicy.clusterSecondaryRangeName` | `string` |  |  |  |
| `spec.nodeConfig.ipAllocationPolicy.clusterIpv4CidrBlock` | `string` |  |  |  |
| `spec.nodeConfig.ipAllocationPolicy.servicesSecondaryRangeName` | `string` |  |  |  |
| `spec.nodeConfig.ipAllocationPolicy.servicesIpv4CidrBlock` | `string` |  |  |  |
| `spec.softwareConfig` | `GcpCloudComposerSoftwareConfig` |  |  |  |
| `spec.softwareConfig.imageVersion` | `string` |  |  |  |
| `spec.softwareConfig.airflowConfigOverrides` | `map<string, string>` |  |  |  |
| `spec.softwareConfig.pypiPackages` | `map<string, string>` |  |  |  |
| `spec.softwareConfig.envVariables` | `map<string, string>` (no secrets: use `secretEnvVariables`) |  |  |  |
| `spec.softwareConfig.webServerPluginsMode` | `string` |  |  |  |
| `spec.softwareConfig.cloudDataLineageIntegration` | `GcpCloudComposerCloudDataLineageIntegration` |  |  |  |
| `spec.softwareConfig.cloudDataLineageIntegration.enabled` | `bool` |  |  |  |
| `spec.softwareConfig.secretEnvVariables` | `map<string, string>` (sensitive) |  |  |  |
| `spec.privateEnvironmentConfig` | `GcpCloudComposerPrivateEnvironmentConfig` |  |  |  |
| `spec.privateEnvironmentConfig.enablePrivateEndpoint` | `bool` |  |  |  |
| `spec.privateEnvironmentConfig.connectionType` | `string` |  |  |  |
| `spec.privateEnvironmentConfig.masterIpv4CidrBlock` | `string` |  |  |  |
| `spec.privateEnvironmentConfig.cloudSqlIpv4CidrBlock` | `string` |  |  |  |
| `spec.privateEnvironmentConfig.cloudComposerNetworkIpv4CidrBlock` | `string` |  |  |  |
| `spec.privateEnvironmentConfig.cloudComposerConnectionSubnetwork` | `string` |  |  |  |
| `spec.privateEnvironmentConfig.enablePrivatelyUsedPublicIps` | `bool` |  |  |  |
| `spec.workloadsConfig` | `GcpCloudComposerWorkloadsConfig` |  |  |  |
| `spec.workloadsConfig.scheduler` | `GcpCloudComposerWorkloadResource` |  |  |  |
| `spec.workloadsConfig.scheduler.cpu` | `double` |  |  |  |
| `spec.workloadsConfig.scheduler.memoryGb` | `double` |  |  |  |
| `spec.workloadsConfig.scheduler.storageGb` | `double` |  |  |  |
| `spec.workloadsConfig.scheduler.count` | `int32` |  |  |  |
| `spec.workloadsConfig.webServer` | `GcpCloudComposerWebServerResource` |  |  |  |
| `spec.workloadsConfig.webServer.cpu` | `double` |  |  |  |
| `spec.workloadsConfig.webServer.memoryGb` | `double` |  |  |  |
| `spec.workloadsConfig.webServer.storageGb` | `double` |  |  |  |
| `spec.workloadsConfig.worker` | `GcpCloudComposerWorkerResource` |  |  |  |
| `spec.workloadsConfig.worker.cpu` | `double` |  |  |  |
| `spec.workloadsConfig.worker.memoryGb` | `double` |  |  |  |
| `spec.workloadsConfig.worker.storageGb` | `double` |  |  |  |
| `spec.workloadsConfig.worker.minCount` | `int32` |  |  |  |
| `spec.workloadsConfig.worker.maxCount` | `int32` |  |  |  |
| `spec.workloadsConfig.triggerer` | `GcpCloudComposerTriggererResource` |  |  |  |
| `spec.workloadsConfig.triggerer.cpu` | `double` |  |  |  |
| `spec.workloadsConfig.triggerer.memoryGb` | `double` |  |  |  |
| `spec.workloadsConfig.triggerer.count` | `int32` |  |  |  |
| `spec.workloadsConfig.dagProcessor` | `GcpCloudComposerWorkloadResource` |  |  |  |
| `spec.workloadsConfig.dagProcessor.cpu` | `double` |  |  |  |
| `spec.workloadsConfig.dagProcessor.memoryGb` | `double` |  |  |  |
| `spec.workloadsConfig.dagProcessor.storageGb` | `double` |  |  |  |
| `spec.workloadsConfig.dagProcessor.count` | `int32` |  |  |  |
| `spec.environmentSize` | `string` |  |  |  |
| `spec.resilienceMode` | `string` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`) |
| `spec.maintenanceWindow` | `GcpCloudComposerMaintenanceWindow` |  |  |  |
| `spec.maintenanceWindow.startTime` | `string` | yes |  |  |
| `spec.maintenanceWindow.endTime` | `string` | yes |  |  |
| `spec.maintenanceWindow.recurrence` | `string` | yes |  |  |
| `spec.recoveryConfig` | `GcpCloudComposerRecoveryConfig` |  |  |  |
| `spec.recoveryConfig.enabled` | `bool` |  |  |  |
| `spec.recoveryConfig.snapshotLocation` | `string` |  |  |  |
| `spec.recoveryConfig.snapshotCreationSchedule` | `string` |  |  |  |
| `spec.recoveryConfig.timeZone` | `string` |  |  |  |
| `spec.webServerNetworkAccessControl` | `GcpCloudComposerWebServerAccessControl` |  |  |  |
| `spec.webServerNetworkAccessControl.allowedIpRanges` | `[]GcpCloudComposerAllowedIpRange` |  |  |  |
| `spec.webServerNetworkAccessControl.allowedIpRanges[].value` | `string` | yes |  |  |
| `spec.webServerNetworkAccessControl.allowedIpRanges[].description` | `string` |  |  |  |
| `spec.masterAuthorizedNetworksConfig` | `GcpCloudComposerMasterAuthorizedNetworksConfig` |  |  |  |
| `spec.masterAuthorizedNetworksConfig.enabled` | `bool` |  |  |  |
| `spec.masterAuthorizedNetworksConfig.cidrBlocks` | `[]GcpCloudComposerCidrBlock` |  |  |  |
| `spec.masterAuthorizedNetworksConfig.cidrBlocks[].cidrBlock` | `string` | yes |  |  |
| `spec.masterAuthorizedNetworksConfig.cidrBlocks[].displayName` | `string` |  |  |  |
| `spec.dataRetentionConfig` | `GcpCloudComposerDataRetentionConfig` |  |  |  |
| `spec.dataRetentionConfig.taskLogsStorageMode` | `string` |  |  |  |
| `spec.dataRetentionConfig.airflowMetadataRetentionMode` | `string` |  |  |  |
| `spec.dataRetentionConfig.airflowMetadataRetentionDays` | `int32` |  |  |  |
| `spec.storageBucket` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.enablePrivateEnvironment` | `bool` |  |  |  |
| `spec.enablePrivateBuildsOnly` | `bool` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project in which to create the Cloud Composer environment.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

GCP region for the Composer environment (e.g., "us-central1",
"europe-west12"). Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.environmentName

`string`

Name of the Composer environment resource in GCP.
Must be lowercase letters, numbers, and hyphens; start with a letter;
end with a letter or number; 1-64 characters.
Optional: when empty, defaults to metadata.name. Immutable after creation.

- rule: {"string":{"pattern":"^([a-z]([a-z0-9-]{0,62}[a-z0-9])?)?$"}}

### spec.nodeConfig

`GcpCloudComposerNodeConfig`

Networking and node configuration for the Composer environment.

### spec.nodeConfig.network

`string | valueFrom`

VPC network for the Composer environment.
Used for Composer 2.x with VPC peering networking.
Not applicable when composer_network_attachment is set (Composer 3 PSC).

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.nodeConfig.subnetwork

`string | valueFrom`

VPC subnetwork for the Composer environment.
Used for Composer 2.x with VPC peering networking.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.nodeConfig.serviceAccount

`string | valueFrom`

Service account the environment's workloads run as. Must hold
roles/composer.worker. Composer 3 REQUIRES this to be explicitly
specified (the API rejects environment creation without it); only
legacy Composer 2 environments fall back to the default Compute
Engine service account.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.nodeConfig.tags

`[]string`

Network tags applied to Composer environment GKE nodes.
Used for firewall rule targeting.

### spec.nodeConfig.composerNetworkAttachment

`string`

PSC Network Attachment for Composer 3 networking.
Format: projects/{project}/regions/{region}/networkAttachments/{name}
Mutually exclusive with network/subnetwork (VPC peering).

### spec.nodeConfig.composerInternalIpv4CidrBlock

`string`

IPv4 CIDR block for Composer 3 internal components.
Must be a /20 range. Only applicable when using Composer 3.

### spec.nodeConfig.enableIpMasqAgent

`bool`

Deploy the ip-masq-agent DaemonSet on the environment's GKE nodes,
SNATing pod traffic to node IPs — needed when the pod CIDR is not
routable in the wider network.

### spec.nodeConfig.ipAllocationPolicy

`GcpCloudComposerIpAllocationPolicy`

VPC-native (alias IP) range assignment for the environment's GKE
pods and services.

- rule: cluster_secondary_range_name and cluster_ipv4_cidr_block are mutually exclusive
- rule: services_secondary_range_name and services_ipv4_cidr_block are mutually exclusive

### spec.nodeConfig.ipAllocationPolicy.clusterSecondaryRangeName

`string`

Name of the subnetwork's existing secondary range to use for pods.
Mutually exclusive with cluster_ipv4_cidr_block.

### spec.nodeConfig.ipAllocationPolicy.clusterIpv4CidrBlock

`string`

CIDR block for the pod range when no named secondary range is used
(e.g. "10.4.0.0/14", or a netmask size like "/14").
Mutually exclusive with cluster_secondary_range_name.

### spec.nodeConfig.ipAllocationPolicy.servicesSecondaryRangeName

`string`

Name of the subnetwork's existing secondary range to use for services.
Mutually exclusive with services_ipv4_cidr_block.

### spec.nodeConfig.ipAllocationPolicy.servicesIpv4CidrBlock

`string`

CIDR block for the services range when no named secondary range is
used. Mutually exclusive with services_secondary_range_name.

### spec.softwareConfig

`GcpCloudComposerSoftwareConfig`

Airflow software configuration including image version, packages,
and configuration overrides.

- rule: a variable is either in env_variables or in secret_env_variables, never both

### spec.softwareConfig.imageVersion

`string`

Composer and Airflow image version.
Format: "composer-A.B.C-airflow-X.Y.Z" (e.g., "composer-2.9.7-airflow-2.9.3").
If empty, the latest stable version is used.

### spec.softwareConfig.airflowConfigOverrides

`map<string, string>`

Airflow configuration property overrides.
Keys are section-key pairs (e.g., "core-dags_are_paused_at_creation": "True").
Values that conflict with managed Composer settings are rejected.

### spec.softwareConfig.pypiPackages

`map<string, string>`

Custom PyPI packages to install in the environment.
Keys are package names, values are version specifiers or empty strings.
Example: {"numpy": ">=1.21", "requests": ""}

### spec.softwareConfig.envVariables

`map<string, string>` · no secrets

Additional environment variables available to all Airflow components.
Values are written into the environment's configuration, where anyone
who can view the environment reads them: configuration only, never a
credential -- a credential goes in secret_env_variables. Names must
match [a-zA-Z_][a-zA-Z0-9_]*, must not be Airflow configuration
overrides (AIRFLOW__<SECTION>__<KEY>; use airflow_config_overrides),
and must not be one of Composer's reserved names (AIRFLOW_HOME,
C_FORCE_ROOT, CONTAINER_NAME, DAGS_FOLDER, GCP_PROJECT, GCS_BUCKET,
GKE_CLUSTER_NAME, SQL_DATABASE, SQL_INSTANCE, SQL_PASSWORD,
SQL_PROJECT, SQL_REGION, SQL_USER, among others Google lists). Map keys
are not CEL-addressable, so the API enforces these at deploy.

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `secretEnvVariables`, which keeps it in a secret store the workload reads by reference
### spec.softwareConfig.webServerPluginsMode

`string`

Web server plugins mode for Composer 3 environments.
When DISABLED, custom Airflow UI plugins are not loaded.
Only applicable to Composer 3.

- rule: {"string":{"in":["","ENABLED","DISABLED"]}}

### spec.softwareConfig.cloudDataLineageIntegration

`GcpCloudComposerCloudDataLineageIntegration`

Cloud Data Lineage integration: Airflow operators report dataset
lineage into Dataplex Data Lineage automatically.
Applies to Composer 2.1.2+.

### spec.softwareConfig.cloudDataLineageIntegration.enabled

`bool`

Whether the integration is enabled.

### spec.softwareConfig.secretEnvVariables

`map<string, string>` · sensitive

Secret values Airflow reads at run time, keyed by environment variable
name. Composer has no secret field for environment variables, so the
component never puts the value on the environment: it keeps each one
in a Secret Manager secret it owns (id
composer_<region>_<environment name>_<key>, replicated only in the
environment's region), grants the environment's node service account
(node_config.service_account, or the project's Compute Engine default
service account when unset) secretAccessor on that secret alone, and
sets the variable to the version's resource name,
projects/<project>/secrets/<id>/versions/<n>. DAG and plugin code reads
the value with the Secret Manager client:

  from google.cloud import secretmanager
  token = secretmanager.SecretManagerServiceClient().access_secret_version(
      name=os.environ["API_TOKEN"]).payload.data.decode()

A changed value adds a version, which changes the variable and runs an
environment update (Composer restarts its Airflow components, which
takes several minutes), so rotation is a deploy; destroying the
environment removes the secrets. The same name rules as env_variables
apply, and a name may not appear in both maps.

### spec.privateEnvironmentConfig

`GcpCloudComposerPrivateEnvironmentConfig`

Private networking configuration for Composer 2.x environments using
VPC peering or Private Service Connect. Not applicable to Composer 3
which uses enable_private_environment and composer_network_attachment instead.

### spec.privateEnvironmentConfig.enablePrivateEndpoint

`bool`

Whether to deny access to the public Airflow web server endpoint.
When true, the web server is only accessible via private IP.

### spec.privateEnvironmentConfig.connectionType

`string`

Connection type for the Composer 2.x private environment.

- rule: {"string":{"in":["","VPC_PEERING","PRIVATE_SERVICE_CONNECT"]}}

### spec.privateEnvironmentConfig.masterIpv4CidrBlock

`string`

IP range for the GKE master network in CIDR notation.
Default: 172.16.0.0/28.

### spec.privateEnvironmentConfig.cloudSqlIpv4CidrBlock

`string`

IP range for the Cloud SQL instance in CIDR notation.

### spec.privateEnvironmentConfig.cloudComposerNetworkIpv4CidrBlock

`string`

IP range for Cloud Composer internal components in CIDR notation.
Applies to Composer 2.x and newer.

### spec.privateEnvironmentConfig.cloudComposerConnectionSubnetwork

`string`

PSC connection subnetwork for Composer 2.x with Private Service Connect.

### spec.privateEnvironmentConfig.enablePrivatelyUsedPublicIps

`bool`

Whether to allow public IPs from non-RFC1918 ranges for IP allocation
in the environment.

### spec.workloadsConfig

`GcpCloudComposerWorkloadsConfig`

Workload resource allocation for Airflow components (scheduler, worker,
web server, triggerer, DAG processor). Applies to Composer 2.x and 3.

### spec.workloadsConfig.scheduler

`GcpCloudComposerWorkloadResource`

Resource allocation for the Airflow scheduler.
The scheduler parses DAGs, manages task scheduling, and triggers task instances.

### spec.workloadsConfig.scheduler.cpu

`double`

CPU allocation in vCPUs (e.g., 0.5, 1.0, 2.0).

### spec.workloadsConfig.scheduler.memoryGb

`double`

Memory allocation in GB (e.g., 1.0, 2.0, 4.0).

### spec.workloadsConfig.scheduler.storageGb

`double`

Storage allocation in GB (e.g., 1.0, 5.0, 10.0).

### spec.workloadsConfig.scheduler.count

`int32`

Number of replicas for this component.

### spec.workloadsConfig.webServer

`GcpCloudComposerWebServerResource`

Resource allocation for the Airflow web server (UI).

### spec.workloadsConfig.webServer.cpu

`double`

CPU allocation in vCPUs.

### spec.workloadsConfig.webServer.memoryGb

`double`

Memory allocation in GB.

### spec.workloadsConfig.webServer.storageGb

`double`

Storage allocation in GB.

### spec.workloadsConfig.worker

`GcpCloudComposerWorkerResource`

Resource allocation for Airflow workers.
Workers execute the actual tasks defined in DAGs.

- rule: max_count must be >= min_count when both are specified

### spec.workloadsConfig.worker.cpu

`double`

CPU allocation per worker in vCPUs.

### spec.workloadsConfig.worker.memoryGb

`double`

Memory allocation per worker in GB.

### spec.workloadsConfig.worker.storageGb

`double`

Storage allocation per worker in GB.

### spec.workloadsConfig.worker.minCount

`int32`

Minimum number of workers. Must be >= 0.

- rule: {"int32":{"gte":0}}

### spec.workloadsConfig.worker.maxCount

`int32`

Maximum number of workers. Must be >= min_count.

- rule: {"int32":{"gte":0}}

### spec.workloadsConfig.triggerer

`GcpCloudComposerTriggererResource`

Resource allocation for the Airflow triggerer.
The triggerer monitors deferred tasks and resumes them when conditions are met.
Critical for deferrable operators in Airflow 2.x+.

### spec.workloadsConfig.triggerer.cpu

`double`

CPU allocation per triggerer in vCPUs.

### spec.workloadsConfig.triggerer.memoryGb

`double`

Memory allocation per triggerer in GB.

### spec.workloadsConfig.triggerer.count

`int32`

Number of triggerer replicas. Set to 0 to disable the triggerer.

### spec.workloadsConfig.dagProcessor

`GcpCloudComposerWorkloadResource`

Resource allocation for the DAG processor.
The DAG processor parses DAG files independently of the scheduler.
Only applicable to Composer 3 (replica count is capped at 3).

### spec.workloadsConfig.dagProcessor.cpu

`double`

CPU allocation in vCPUs (e.g., 0.5, 1.0, 2.0).

### spec.workloadsConfig.dagProcessor.memoryGb

`double`

Memory allocation in GB (e.g., 1.0, 2.0, 4.0).

### spec.workloadsConfig.dagProcessor.storageGb

`double`

Storage allocation in GB (e.g., 1.0, 5.0, 10.0).

### spec.workloadsConfig.dagProcessor.count

`int32`

Number of replicas for this component.

### spec.environmentSize

`string`

Size of the Composer environment. Controls the managed infrastructure
capacity (GKE cluster and database sizing behind the scenes).

- rule: {"string":{"in":["","ENVIRONMENT_SIZE_SMALL","ENVIRONMENT_SIZE_MEDIUM","ENVIRONMENT_SIZE_LARGE","ENVIRONMENT_SIZE_EXTRA_LARGE"]}}

### spec.resilienceMode

`string`

Resilience mode for the Composer environment. HIGH_RESILIENCE provides
multi-zone redundancy for increased availability. Applies to Composer 2.1.15+.

- rule: {"string":{"in":["","STANDARD_RESILIENCE","HIGH_RESILIENCE"]}}

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key for the Composer environment.
All Composer-managed resources (GKE nodes, Cloud SQL, Cloud Storage) are
encrypted with this key. Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.maintenanceWindow

`GcpCloudComposerMaintenanceWindow`

Maintenance window configuration for the Composer environment.
Defines when GCP may perform maintenance operations on the environment.

### spec.maintenanceWindow.startTime

`string` · required

Start time of the maintenance window in RFC3339 format.
Example: "2026-01-01T00:00:00Z"

- rule: {"required":true}

### spec.maintenanceWindow.endTime

`string` · required

End time of the maintenance window in RFC3339 format.
Must be after start_time. The window duration must be at least 12 hours.

- rule: {"required":true}

### spec.maintenanceWindow.recurrence

`string` · required

Recurrence specification in RFC5545 RRULE format.
Examples: "FREQ=WEEKLY;BYDAY=TU,WE,TH" or "FREQ=DAILY".

- rule: {"required":true}

### spec.recoveryConfig

`GcpCloudComposerRecoveryConfig`

Recovery configuration with scheduled snapshots for disaster recovery.

### spec.recoveryConfig.enabled

`bool`

Whether scheduled snapshots are enabled.

### spec.recoveryConfig.snapshotLocation

`string`

Cloud Storage location for snapshots (GCS bucket folder URI).
Example: "gs://my-bucket/composer-snapshots"

### spec.recoveryConfig.snapshotCreationSchedule

`string`

Cron schedule for snapshot creation in Unix-cron format.
Example: "0 4 * * *" (daily at 4 AM).

### spec.recoveryConfig.timeZone

`string`

Time zone for the cron schedule (e.g., "America/Los_Angeles", "UTC").

### spec.webServerNetworkAccessControl

`GcpCloudComposerWebServerAccessControl`

Network-level access restrictions for the Airflow web server UI.
When configured, only requests from the specified IP ranges are allowed.

### spec.webServerNetworkAccessControl.allowedIpRanges

`[]GcpCloudComposerAllowedIpRange`

Allowed IP ranges that can access the Airflow web server.

### spec.webServerNetworkAccessControl.allowedIpRanges[].value

`string` · required

IP address or CIDR range (e.g., "10.0.0.0/8", "203.0.113.0/24").

- rule: {"required":true}

### spec.webServerNetworkAccessControl.allowedIpRanges[].description

`string`

Optional human-readable description of this IP range.

### spec.masterAuthorizedNetworksConfig

`GcpCloudComposerMasterAuthorizedNetworksConfig`

IP-based access control for the environment's GKE cluster master.
Restricts which networks can reach the Kubernetes control plane that
runs the Airflow workloads.

### spec.masterAuthorizedNetworksConfig.enabled

`bool`

Whether master authorized networks are enforced.

### spec.masterAuthorizedNetworksConfig.cidrBlocks

`[]GcpCloudComposerCidrBlock`

Networks allowed to reach the Kubernetes control plane.

### spec.masterAuthorizedNetworksConfig.cidrBlocks[].cidrBlock

`string` · required

CIDR block allowed to access the cluster master (e.g., "10.0.0.0/8").

- rule: {"required":true}

### spec.masterAuthorizedNetworksConfig.cidrBlocks[].displayName

`string`

Optional display name for the network.

### spec.dataRetentionConfig

`GcpCloudComposerDataRetentionConfig`

Retention policies for Airflow task logs and metadata database rows —
the levers that keep long-lived environments from accumulating
unbounded operational data.

- rule: airflow_metadata_retention_days requires airflow_metadata_retention_mode to be set

### spec.dataRetentionConfig.taskLogsStorageMode

`string`

Where Airflow task logs are stored.
CLOUD_LOGGING_ONLY: task logs go to Cloud Logging only.
CLOUD_LOGGING_AND_CLOUD_STORAGE: logs also land in the environment's
bucket. Applies to Composer 2.0.32+ (not Composer 3).

- rule: {"string":{"in":["","CLOUD_LOGGING_ONLY","CLOUD_LOGGING_AND_CLOUD_STORAGE"]}}

### spec.dataRetentionConfig.airflowMetadataRetentionMode

`string`

Whether Airflow metadata database retention is enforced.
Composer 3 only.

- rule: {"string":{"in":["","RETENTION_MODE_ENABLED","RETENTION_MODE_DISABLED"]}}

### spec.dataRetentionConfig.airflowMetadataRetentionDays

`int32`

Days of Airflow metadata (task history, XComs, logs metadata) to
retain when retention is enabled (30-730). Composer 3 only.

- rule: airflow_metadata_retention_days must be between 30 and 730

### spec.storageBucket

`string | valueFrom`

Existing Cloud Storage bucket for the environment's DAGs, plugins, and
data (instead of the bucket Composer auto-creates). Resolves to the
bucket name. Immutable after creation.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.enablePrivateEnvironment

`bool`

Enable private environment for Composer 3 environments. When true, the
environment does not have a public IP endpoint for the web server.
Only applicable to Composer 3.

### spec.enablePrivateBuildsOnly

`bool`

Enable private builds only for Composer 3 environments. When true,
only builds using private connectivity are allowed for Python packages.
Only applicable to Composer 3.

### spec.labels

`map<string, string>`

User-defined labels to organize and track the environment. Merged
beneath Planton's platform attribution labels (platform keys win on
conflict).

### spec.deletionPolicy

`string`

Deletion policy for the environment — what happens when this
resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the environment is deleted (10-15 minutes: Composer
               tears down the managed GKE cluster and database).
               The DAG bucket Composer auto-created survives — GCP
               never deletes it with the environment
  "PREVENT" -- destroy FAILS; protects the environment whose DAGs
               a data platform runs on
  "ABANDON" -- the environment is removed from management but keeps
               running (and billing meaningfully — Composer bills
               for its infrastructure even when idle) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `secret_env_variables_need_email_identity`: with software_config.secret_env_variables, node_config.service_account must name the account by email (or projects/{project}/serviceAccounts/{email}): the secret grant is made to that email

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudComposerEnvironment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.environment_id` | `string` | Fully qualified resource ID of the Composer environment. Format: projects/{project}/locations/{region}/environments/{name} |
| `status.outputs.environment_name` | `string` | Short name of the Composer environment. |
| `status.outputs.airflow_uri` | `string` | URI of the Apache Airflow web UI hosted by the Composer environment. |
| `status.outputs.dag_gcs_prefix` | `string` | Cloud Storage prefix where DAG files should be uploaded. Format: gs://{bucket}/dags |
| `status.outputs.gke_cluster` | `string` | Name of the underlying GKE cluster managed by Cloud Composer. Populated for Composer 2 environments (the cluster runs in your project); empty for Composer 3, whose cluster is Google-managed in a tenant project and never surfaced. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.nodeConfig.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.nodeConfig.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.nodeConfig.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.kmsKeyName` | GcpKmsKeyHandle | `status.outputs.kms_key` |
| `spec.storageBucket` | GcpGcsBucket | `status.outputs.bucket_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCloudComposerUserWorkloadsConfigMap | `spec.environment` | `status.outputs.environment_name` |
| GcpCloudComposerUserWorkloadsSecret | `spec.environment` | `status.outputs.environment_name` |

## See Also

- [Overview](../README.md)
