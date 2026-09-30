# AzureContainerAppEnvironment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureContainerAppEnvironmentSpec** defines the configuration for creating
an Azure Container Apps Managed Environment (Microsoft.App/managedEnvironments),
the hosting platform for Azure Container Apps.

A Container App Environment is a secure boundary around one or more container
apps and jobs. All workloads within an environment share the same virtual
network, logging configuration, Dapr infrastructure, and compute capacity.
It is the execution boundary for containerized workloads -- the environment
hosts, the apps run.

**Networking modes**:

External (default):
  - Apps can receive traffic from the public internet
  - Azure assigns a public static IP and default domain
  - No subnet required (Azure manages networking)

VNet-injected (infrastructure_subnet_id provided):
  - Apps run inside a customer-managed VNet
  - Enables private connectivity to databases, storage, and other VNet resources
  - Subnet must be /21 or larger (minimum 2048 IPs for Container Apps infrastructure)
  - Unlocks internal_load_balancer_enabled and zone_redundancy_enabled

Internal (internal_load_balancer_enabled = true):
  - Apps are only accessible from within the VNet
  - Requires infrastructure_subnet_id
  - Static IP is private, default domain resolves to the private IP
  - Used for backend services, microservice meshes, internal APIs

**Workload profiles**:

By default, environments run on the Consumption plan (serverless,
pay-per-use, scale to zero). For workloads requiring dedicated compute,
GPU access, or guaranteed resources, add workload profiles. Apps and jobs
select a profile by name; anything that names no profile runs on
Consumption.

**Logging**:

`logs_destination` selects where application logs are persisted:
LOG_ANALYTICS pairs with `log_analytics_workspace_id` (query via KQL,
alerting, dashboards); AZURE_MONITOR routes through Azure Monitor
diagnostic settings configured separately on the environment; leaving it
unspecified without a workspace means logs are streaming-only (visible in
`az containerapp logs show`, never stored).

**Custom DNS suffix**:

The optional `custom_domain` block replaces the environment's generated
`*.{region}.azurecontainerapps.io` default domain with your own DNS suffix
(e.g. apps.example.com) backed by a wildcard certificate. This is the
environment-wide suffix; per-app custom domains bind on the app itself.

**ForceNew fields** (changing these destroys and recreates the environment
-- and every app in it): `environment_name`, `region`, `resource_group`,
`dapr_application_insights_connection_string`,
`infrastructure_resource_group_name`, `infrastructure_subnet_id`,
`internal_load_balancer_enabled`, `zone_redundancy_enabled`. Going from
zero workload profiles to some (or some to zero) also forces replacement;
changing the profiles within a non-empty set updates in place.

**Referenced by**: AzureContainerApp, AzureContainerAppJob,
AzureContainerAppEnvironmentStorage, AzureContainerAppEnvironmentDaprComponent
(container_app_environment_id).

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerAppEnvironment
metadata:
  name: test-env
spec:
  region: eastus
  resource_group:
    value: test-rg
  environment_name: test-container-env
  logs_destination: LOG_ANALYTICS
  log_analytics_workspace_id:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.OperationalInsights/workspaces/test-law
  infrastructure_subnet_id:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/apps
  internal_load_balancer_enabled: false
  zone_redundancy_enabled: true
  public_network_access: DISABLED
  mutual_tls_enabled: true
  infrastructure_resource_group_name: me-test-container-env
  workload_profiles:
    - name: dedicated-d4
      workload_profile_type: D4
      minimum_count: 0
      maximum_count: 3
    - name: gpu-serverless
      workload_profile_type: CONSUMPTION_GPU_NC8AS_T4
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    user_assigned_identity_ids:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-uai
  tags:
    team: platform
    costcenter: eng-1234
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.environmentName` | `string` | yes |  |  |
| `spec.logsDestination` | `enum` |  |  |  |
| `spec.logAnalyticsWorkspaceId` | `string \| valueFrom` |  |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`) |
| `spec.daprApplicationInsightsConnectionString` | `string` (sensitive) |  |  |  |
| `spec.infrastructureSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.infrastructureResourceGroupName` | `string` |  |  |  |
| `spec.internalLoadBalancerEnabled` | `bool` |  | `false` |  |
| `spec.zoneRedundancyEnabled` | `bool` |  | `false` |  |
| `spec.publicNetworkAccess` | `enum` |  |  |  |
| `spec.mutualTlsEnabled` | `bool` |  | `false` |  |
| `spec.workloadProfiles` | `[]AzureContainerAppEnvironmentWorkloadProfile` |  |  |  |
| `spec.workloadProfiles[].name` | `string` | yes |  |  |
| `spec.workloadProfiles[].workloadProfileType` | `enum` | yes |  |  |
| `spec.workloadProfiles[].minimumCount` | `int32` |  |  |  |
| `spec.workloadProfiles[].maximumCount` | `int32` |  |  |  |
| `spec.customDomain` | `AzureContainerAppEnvironmentCustomDomain` |  |  |  |
| `spec.customDomain.dnsSuffix` | `string` | yes |  |  |
| `spec.customDomain.certificateBlobBase64` | `string` (sensitive) | yes |  |  |
| `spec.customDomain.certificatePassword` | `string` (sensitive) | yes |  |  |
| `spec.identity` | `AzureContainerAppEnvironmentIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Container App Environment will be created.
Examples: "eastus", "westus2", "westeurope", "southeastasia".

**ForceNew**: Changing this destroys and recreates the environment.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the Container App Environment will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

**ForceNew**: Changing this destroys and recreates the environment.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.environmentName

`string` · required

The name of the Container App Environment.
Alphanumeric characters and hyphens; may not start or end with a
hyphen; 2 to 60 characters (Azure accepts upper- and lowercase).

This name appears in the Azure portal and CLI, and is embedded in the
environment's default domain ({app-name}.{env-default-domain}).

**ForceNew**: Changing this destroys and recreates the environment.

- rule: Environment name must contain only alphanumeric characters and hyphens, and may not start or end with a hyphen
- rule: {"required":true,"string":{"minLen":"2","maxLen":"60"}}

### spec.logsDestination

`enum`

Where application logs are persisted for this environment.

LOG_ANALYTICS: logs land in the Log Analytics workspace named by
`log_analytics_workspace_id` (required with this choice) -- enables KQL
queries, alert rules, and Azure Monitor dashboards.

AZURE_MONITOR: logs route through Azure Monitor diagnostic settings,
which are configured separately on the environment; do not set
`log_analytics_workspace_id` with this choice.

Unspecified: when `log_analytics_workspace_id` is provided the modules
deploy LOG_ANALYTICS (the destination the workspace implies); without a
workspace, logs are streaming-only (no persistent storage).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_container_app_environment_logs_destination_unspecified` -- Not specified -- with a log_analytics_workspace_id the modules deploy LOG_ANALYTICS; without one, logs are streaming-only (never stored).
- `LOG_ANALYTICS` -- Persist logs in the Log Analytics workspace named by log_analytics_workspace_id (KQL queries, alerting, dashboards).
- `AZURE_MONITOR` -- Route logs through Azure Monitor diagnostic settings (configured separately on the environment).

### spec.logAnalyticsWorkspaceId

`string | valueFrom`

The Log Analytics Workspace to link for centralized log collection.
Required when `logs_destination` is LOG_ANALYTICS; must be omitted when
it is AZURE_MONITOR. Providing a workspace with `logs_destination`
unspecified also deploys log-analytics.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_id}} -- a bare string does not parse

### spec.daprApplicationInsightsConnectionString

`string` · sensitive

Application Insights connection string used by Dapr to export
service-to-service communication telemetry. Only meaningful for
Dapr-enabled apps; leave empty otherwise.

The connection string embeds the instrumentation key -- a write
credential for your telemetry stream -- so it is handled as a secret.

**ForceNew**: Changing this destroys and recreates the environment
(Azure never returns it on read, so it cannot be updated in place).

### spec.infrastructureSubnetId

`string | valueFrom`

The existing Subnet to use for the Container Apps infrastructure.
When provided, the environment is VNet-injected, enabling private
connectivity to other VNet resources (databases, storage, etc.).

**Important**: The subnet must have a /21 or larger address space
(minimum 2048 IPs). Container Apps infrastructure reserves a significant
portion of the address space for platform services.

When set, unlocks `internal_load_balancer_enabled` and
`zone_redundancy_enabled`.

**ForceNew**: Changing this destroys and recreates the environment.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.infrastructureResourceGroupName

`string`

Name of the platform-managed resource group Azure creates to host the
environment's infrastructure resources (load balancers, IPs). Only
valid when `workload_profiles` is non-empty; when omitted, Azure
generates a name (ME_{environment}_{rg}_{region}). Setting it gives
governance-conscious orgs a predictable name to scope policies on.

**ForceNew**: Changing this destroys and recreates the environment.

### spec.internalLoadBalancerEnabled

`bool` · optional (explicit presence)

Enable Internal Load Balancing mode. When true, apps in this environment
are only accessible from within the VNet (no public internet access).

Requires `infrastructure_subnet_id` to be set.

Use this for backend microservices, internal APIs, and workloads that
should not be exposed to the public internet.

Default: false (external mode -- apps are publicly accessible)

**ForceNew**: Changing this destroys and recreates the environment.

- default: `false`

### spec.zoneRedundancyEnabled

`bool` · optional (explicit presence)

Enable zone redundancy. Distributes the environment's infrastructure
across multiple availability zones for higher resilience.

Requires `infrastructure_subnet_id` to be set.

When enabled, Container Apps infrastructure is spread across all
available zones in the region (typically 3). Recommended for
production workloads requiring high availability.

Default: false

**ForceNew**: Changing this destroys and recreates the environment.

- default: `false`

### spec.publicNetworkAccess

`enum`

Whether the environment accepts traffic from the public internet at
the platform level. Unspecified lets Azure derive it from the network
configuration (Enabled for external environments, Disabled behind an
internal load balancer). Set DISABLED to pair the environment with
private endpoints; ENABLED cannot be combined with
`internal_load_balancer_enabled`.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_container_app_environment_public_network_access_unspecified` -- Not specified -- Azure derives the value from the network configuration (Enabled for external environments, Disabled behind an internal load balancer).
- `ENABLED` -- The environment accepts traffic from the public internet.
- `DISABLED` -- The environment only accepts traffic through private connectivity (VNet / private endpoints).

### spec.mutualTlsEnabled

`bool` · optional (explicit presence)

Enable mutual TLS (mTLS) between apps in the environment: Azure
issues per-app certificates and encrypts + authenticates all
app-to-app traffic. Adds response latency and reduces maximum
throughput in high-load scenarios -- enable when the compliance
posture requires encrypted east-west traffic.

Default: false

- default: `false`

### spec.workloadProfiles

`[]AzureContainerAppEnvironmentWorkloadProfile`

Workload profiles for dedicated or GPU compute.

By default, environments include the serverless "Consumption" profile.
Declare additional profiles here for dedicated compute (D/E families),
GPU compute (NC*-A100), or serverless GPU (Consumption-GPU-*); apps
and jobs select a profile by name via `workload_profile_name`.

Azure always includes the standard Consumption profile -- do not
declare it here.

**Note**: Going from zero profiles to some (or some to zero) forces
environment replacement; changing profiles within a non-empty set
updates in place.

- rule: minimum_count and maximum_count only apply to dedicated profiles (D, E, and NC families) -- Consumption-family profiles are serverless and manage capacity themselves

### spec.workloadProfiles[].name

`string` · required

The name of the workload profile.
Apps and jobs reference this name in `workload_profile_name` to select
which profile to run on (e.g., "gpu-pool", "high-memory").

Must be unique within the environment.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.workloadProfiles[].workloadProfileType

`enum` · required

The profile type (SKU) that determines the VM size -- see the
AzureContainerAppEnvironmentWorkloadProfileType enum for the family
guide (dedicated D/E, GPU NC*-A100, serverless GPU Consumption-GPU-*).

- rule: {"required":true,"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_environment_workload_profile_type_unspecified` -- Not specified -- invalid; pick an explicit profile type.
- `CONSUMPTION` -- Serverless consumption (pay-per-use, scale to zero). Azure includes this profile automatically -- declare it only when pinning a named profile to it.
- `CONSUMPTION_GPU_NC8AS_T4` -- Serverless GPU: NVIDIA T4 (8 vCPU class) on consumption billing -- scale-to-zero inference workloads.
- `CONSUMPTION_GPU_NC24_A100` -- Serverless GPU: NVIDIA A100 (24 vCPU class) on consumption billing.
- `D4` -- General-purpose dedicated compute (4/8/16/32 vCPU).
- `D8`
- `D16`
- `D32`
- `E4` -- Memory-optimized dedicated compute (4/8/16/32 vCPU, double the memory-per-core of the D family).
- `E8`
- `E16`
- `E32`
- `NC24_A100` -- Dedicated GPU compute: NVIDIA A100 (24/48/96 vCPU classes).
- `NC48_A100`
- `NC96_A100`

### spec.workloadProfiles[].minimumCount

`int32` · optional (explicit presence)

Minimum number of instances for this profile.
Set to 0 to allow scale-to-zero (instances are deallocated when idle).
Set to 1+ to keep pre-warmed instances for reduced cold-start latency.

Only meaningful for dedicated profiles (D/E/NC families) -- the
Consumption-family profiles are serverless and manage capacity
themselves.

- rule: {"int32":{"gte":0}}

### spec.workloadProfiles[].maximumCount

`int32` · optional (explicit presence)

Maximum number of instances this profile can scale to.
Controls the upper bound on cost and capacity.

Only meaningful for dedicated profiles (D/E/NC families).

- rule: {"int32":{"gte":0}}

### spec.customDomain

`AzureContainerAppEnvironmentCustomDomain`

Replace the environment's generated default domain
(*.{region}.azurecontainerapps.io) with your own DNS suffix backed by
a wildcard certificate. Azure models this as part of the environment
itself (one custom DNS suffix per environment), so it is configured
here rather than as a separate resource. Per-app custom domains are a
different mechanism and bind on the individual app.

### spec.customDomain.dnsSuffix

`string` · required

The DNS suffix that replaces the environment's default domain --
apps become {app-name}.{dns_suffix}. Example: "apps.example.com".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.customDomain.certificateBlobBase64

`string` · required · sensitive

The wildcard certificate for *.{dns_suffix}, as a base64-encoded
PFX/PKCS12 blob. The blob bundles the private key, so it is handled
as a secret.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.customDomain.certificatePassword

`string` · required · sensitive

The password protecting the certificate blob.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.identity

`AzureContainerAppEnvironmentIdentity`

Managed identity for the environment. Used by environment-level
integrations -- for example pulling images for jobs, or Key Vault
certificate references in the custom-domain configuration.

- rule: user_assigned_identity_ids is required with USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED, and must be empty with SYSTEM_ASSIGNED

### spec.identity.type

`enum`

The identity model: SYSTEM_ASSIGNED (Azure creates and rotates a
service principal bound to the environment's lifecycle), USER_ASSIGNED
(bring identities from user_assigned_identity_ids, shareable across
resources), or SYSTEM_AND_USER_ASSIGNED (both).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_container_app_environment_identity_type_unspecified` -- Not specified -- invalid; choose an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates a service principal bound to the environment's lifecycle.
- `USER_ASSIGNED` -- Bring your own AzureUserAssignedIdentity entries -- shareable across resources and grantable before the environment exists.
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned principal and user-assigned identities.

### spec.identity.userAssignedIdentityIds

`[]string | valueFrom`

The user-assigned identities to attach -- required when (and only
meaningful when) type includes USER_ASSIGNED. Each entry references
an AzureUserAssignedIdentity's ARM id.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Free-form Azure resource tags applied to the environment, merged over
the platform's metadata-derived tags (user tags win on key collision)
-- the hooks for cost allocation, chargeback reports, and Azure Policy
governance rules that filter or group by them. Updatable in place.

## Validation Rules

- `environment_ilb_requires_subnet`: internal_load_balancer_enabled requires infrastructure_subnet_id -- internal load balancing only exists for VNet-injected environments
- `environment_zone_redundancy_requires_subnet`: zone_redundancy_enabled requires infrastructure_subnet_id -- zone redundancy only exists for VNet-injected environments
- `environment_log_analytics_requires_workspace`: logs_destination LOG_ANALYTICS requires log_analytics_workspace_id -- name the workspace the logs should land in
- `environment_azure_monitor_forbids_workspace`: logs_destination AZURE_MONITOR cannot be combined with log_analytics_workspace_id -- Azure Monitor routing is configured through diagnostic settings, not a linked workspace
- `environment_infra_rg_requires_workload_profiles`: infrastructure_resource_group_name is only valid when workload_profiles is non-empty (consumption-only environments do not support a custom infrastructure resource group)
- `environment_public_access_conflicts_with_ilb`: public_network_access ENABLED cannot be combined with internal_load_balancer_enabled -- an internal environment has no public entry point

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureContainerAppEnvironment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.environment_id` | `string` | The Azure Resource Manager ID of the Container App Environment. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.App/managedEnvironments/{name} This is the primary reference output -- every kind that lives inside the environment consumes it. |
| `status.outputs.environment_name` | `string` | The name of the Container App Environment. |
| `status.outputs.default_domain` | `string` | The default, publicly resolvable domain name for apps in this environment. Apps are accessible at {app-name}.{default_domain}. Useful for DNS CNAME records and verifying environment connectivity. |
| `status.outputs.static_ip_address` | `string` | The static IP address of the environment. For external environments, this is a public IP. For internal environments (internal_load_balancer_enabled=true), this is a private IP. Useful for DNS A records, firewall allowlists, and network debugging. |
| `status.outputs.platform_reserved_cidr` | `string` | The IP range (CIDR notation) reserved for environment infrastructure. This range is used by the Container Apps platform for internal services. Only populated for VNet-injected environments. Useful for network planning and ensuring no CIDR conflicts with other VNet resources. |
| `status.outputs.platform_reserved_dns_ip_address` | `string` | The IP address from platform_reserved_cidr reserved for the internal DNS server. Used by container apps for service discovery within the environment. Only populated for VNet-injected environments. Useful for custom DNS configuration and network debugging. |
| `status.outputs.docker_bridge_cidr` | `string` | The Docker bridge network address (CIDR) used inside the environment's infrastructure. Only populated for VNet-injected environments; useful when diagnosing address-space overlaps. |
| `status.outputs.custom_domain_verification_id` | `string` | The value Azure expects in the asuid.{dns_suffix} TXT record to prove ownership of a custom DNS suffix (and of per-app custom domains). Publish it before configuring `custom_domain`. |
| `status.outputs.identity_principal_id` | `string` | The principal (object) ID of the environment's system-assigned managed identity. Empty unless the identity block enables SYSTEM_ASSIGNED. Grant this principal roles (e.g. AcrPull on a container registry) to let the environment's platform operations authenticate keylessly. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.logAnalyticsWorkspaceId` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_id` |
| `spec.infrastructureSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.identity.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureContainerApp | `spec.containerAppEnvironmentId` | `status.outputs.environment_id` |
| AzureContainerAppEnvironmentCertificate | `spec.containerAppEnvironmentId` | `status.outputs.environment_id` |
| AzureContainerAppEnvironmentDaprComponent | `spec.containerAppEnvironmentId` | `status.outputs.environment_id` |
| AzureContainerAppEnvironmentManagedCertificate | `spec.containerAppEnvironmentId` | `status.outputs.environment_id` |
| AzureContainerAppEnvironmentStorage | `spec.containerAppEnvironmentId` | `status.outputs.environment_id` |
| AzureContainerAppJob | `spec.containerAppEnvironmentId` | `status.outputs.environment_id` |
| AzureFrontDoorOrigin | `spec.privateLink.privateLinkTargetId` | `status.outputs.environment_id` |
| AzurePlantonRunner | `spec.containerAppEnvironmentId` | `status.outputs.environment_id` |

## See Also

- [Overview](../README.md)
