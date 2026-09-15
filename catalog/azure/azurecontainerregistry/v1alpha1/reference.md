# AzureContainerRegistry

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureContainerRegistrySpec** defines the configuration for creating an
Azure Container Registry (ACR): the managed, private OCI registry that
stores the container images and artifacts a platform's workloads run.

The SKU is the registry's feature gate, and the spec mirrors Azure's own
tiering rather than hiding it:
- BASIC/STANDARD carry the core push/pull surface (Standard additionally
  allows anonymous pull).
- PREMIUM unlocks the enterprise surface: geo-replication, zone
  redundancy, network isolation (network_rule_set, disabling public
  access, dedicated data endpoints), quarantine/retention policies,
  and customer-managed-key encryption.
Spec-level validation enforces the same SKU gates ARM does, so a
misconfigured manifest fails at validation, not at apply.

What the registry composes with is referenced, never created here: the
user-assigned identity that unwraps the CMK encryption key is a
first-class AzureUserAssignedIdentity, and an AKS cluster pulls from the
registry by referencing its container_registry_id output.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerRegistry
metadata:
  name: test-acr
  labels:
    environment: production
    team: platform
spec:
  # The home replica's region.
  region: eastus

  # The resource group the registry lives in (literal value here; a
  # manifest can also reference an AzureResourceGroup's name output via
  # valueFrom).
  resourceGroup:
    value: test-rg

  # Globally unique registry name (5-50 lowercase alphanumerics; becomes
  # {name}.azurecr.io).
  registryName: testcompanyacr

  # PREMIUM to exercise the enterprise surface below.
  sku: PREMIUM

  # Keep the static admin account off; Entra-based auth is the production
  # path.
  adminUserEnabled: false

  # Zone-redundant home replica (Premium; fixed at creation).
  zoneRedundancyEnabled: true

  # Dedicated regional data endpoints for exact firewall allowlisting
  # (Premium).
  dataEndpointEnabled: true

  # Purge untagged manifests after a week (Premium).
  retentionPolicyInDays: 7

  # Public registry restricted to a CIDR allowlist (Premium).
  networkRuleSet:
    defaultAction: DENY
    ipRules:
      - ipRange: 203.0.113.0/24

  # Geo-replications (Premium; must not contain the home region).
  georeplications:
    - location: westeurope
      zoneRedundancyEnabled: true
    - location: southeastasia
      globalEndpointRoutingEnabled: true

  # User tags merged over the metadata-derived tags.
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.registryName` | `string` | yes |  |  |
| `spec.sku` | `enum` |  |  |  |
| `spec.adminUserEnabled` | `bool` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.zoneRedundancyEnabled` | `bool` |  |  |  |
| `spec.anonymousPullEnabled` | `bool` |  |  |  |
| `spec.dataEndpointEnabled` | `bool` |  |  |  |
| `spec.quarantinePolicyEnabled` | `bool` |  |  |  |
| `spec.retentionPolicyInDays` | `int32` |  |  |  |
| `spec.exportPolicyEnabled` | `bool` |  | `true` |  |
| `spec.networkRuleBypassOption` | `enum` |  |  |  |
| `spec.networkRuleSet` | `AzureContainerRegistryNetworkRuleSet` |  |  |  |
| `spec.networkRuleSet.defaultAction` | `enum` |  |  |  |
| `spec.networkRuleSet.ipRules` | `[]AzureContainerRegistryIpRule` |  |  |  |
| `spec.networkRuleSet.ipRules[].ipRange` | `string` | yes |  |  |
| `spec.georeplications` | `[]AzureContainerRegistryGeoreplication` |  |  |  |
| `spec.georeplications[].location` | `string` | yes |  |  |
| `spec.georeplications[].zoneRedundancyEnabled` | `bool` |  |  |  |
| `spec.georeplications[].globalEndpointRoutingEnabled` | `bool` |  |  |  |
| `spec.georeplications[].tags` | `map<string, string>` |  |  |  |
| `spec.identity` | `AzureContainerRegistryIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.encryption` | `AzureContainerRegistryEncryption` |  |  |  |
| `spec.encryption.identityClientId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.client_id`) |
| `spec.encryption.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the registry's home replica lives in, e.g. "eastus",
"westeurope". Additional regions are geo-replications, not a region
change. Changing the region replaces the registry.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the registry will be created in. Can be a
literal resource-group name or a reference to an AzureResourceGroup's
name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.registryName

`string` · required

The name of the registry: 5-50 lowercase alphanumerics, GLOBALLY
unique across all of Azure because it becomes the registry's DNS name
({name}.azurecr.io -- the login_server output). Changing the name
replaces the registry and its contents do not migrate; every image
would need to be re-pushed.

- rule: {"required":true,"string":{"pattern":"^[a-z0-9]{5,50}$"}}

### spec.sku

`enum`

The pricing tier -- the registry's feature gate. Unspecified applies
STANDARD: the production baseline (100 GiB included storage, webhooks,
anonymous pull available). BASIC is a cost-optimized dev/test tier
with the same API surface but low storage and throughput limits.
PREMIUM adds the enterprise surface gated throughout this spec
(geo-replication, zone redundancy, network isolation, policies, CMK)
plus the highest throughput and 500 GiB included storage. The SKU can
be changed in place -- but downgrading requires the Premium-only
features below to be unset first.

Allowed values (use exactly as shown):

- `azure_container_registry_sku_unspecified` -- Not specified: STANDARD.
- `BASIC` -- Cost-optimized dev/test tier: full API surface, low storage and throughput limits, no premium features.
- `STANDARD` -- The production baseline: 100 GiB included storage, webhooks, anonymous pull available.
- `PREMIUM` -- The enterprise tier: geo-replication, zone redundancy, network isolation, quarantine/retention policies, CMK encryption, highest throughput.

### spec.adminUserEnabled

`bool`

Enables the registry's built-in admin account: a single username (the
registry name) with two rotatable passwords, surfaced in the
admin_username/admin_password outputs. Azure's default is false --
production authentication should be Microsoft Entra ID (managed
identities, service principals, or repo-scoped tokens). Enable it only
for services that can consume nothing but a static username/password
(e.g. some app-hosting image pulls). Updatable in place.

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the registry accepts connections from the public internet.
Azure's default is true. Setting false takes the registry fully
private -- reachable only through private endpoints -- and requires
the PREMIUM SKU. For a public registry restricted to known addresses,
keep this true and use network_rule_set instead. Updatable in place.

- default: `true`

### spec.zoneRedundancyEnabled

`bool`

Whether the home replica's data is zone-redundant: spread across
availability zones so a zone outage does not take the registry down.
PREMIUM only, and fixed at creation -- enabling it later means
replacing the registry. Each geo-replication declares its own zone
redundancy separately.

### spec.anonymousPullEnabled

`bool`

Whether unauthenticated (anonymous) image pulls are allowed. STANDARD
or PREMIUM only. This makes every repository in the registry publicly
readable -- the right shape for a public artifact distribution
registry, and only that. Updatable in place.

### spec.dataEndpointEnabled

`bool`

Whether the registry gets dedicated regional data endpoints
({name}.{region}.data.azurecr.io, surfaced in the
data_endpoint_host_names output) instead of serving blobs from shared
storage endpoints. PREMIUM only. Enable it when egress firewalls must
allowlist registry traffic by hostname -- dedicated endpoints make the
allowlist exact. Updatable in place.

### spec.quarantinePolicyEnabled

`bool`

Whether newly pushed images are quarantined until a scanner (or other
automation) marks them passed; unquarantined clients cannot pull them.
PREMIUM only. The quarantine workflow itself is driven through the
registry's data-plane API by the scanning tooling. Updatable in place.

### spec.retentionPolicyInDays

`int32` · optional (explicit presence)

Retention window, in days, after which UNTAGGED manifests are
automatically purged (0 means delete untagged manifests immediately).
PREMIUM only; leave unset to keep untagged manifests forever (Azure's
default). This is the built-in hygiene lever that stops CI push churn
from growing storage without bound. Updatable in place.

- rule: {"int32":{"lte":365,"gte":0}}

### spec.exportPolicyEnabled

`bool` · optional (explicit presence)

Whether registry artifacts can be exported (e.g. imported into another
registry or transferred out). Azure's default is true. Disabling
export is a data-exfiltration control for locked-down registries; it
requires the PREMIUM SKU and public_network_access_enabled explicitly
false (ARM enforces the pairing; so does spec validation). Updatable
in place.

- default: `true`

### spec.networkRuleBypassOption

`enum`

Whether trusted Azure services (e.g. ACR Tasks, Microsoft Defender)
may reach a network-restricted registry despite network_rule_set /
private-only access. Unspecified applies Azure's default
(AZURE_SERVICES) -- the pragmatic choice that keeps first-party
integrations working. NONE closes even that door. Updatable in place.

Allowed values (use exactly as shown):

- `azure_container_registry_network_rule_bypass_option_unspecified` -- Not specified: Azure's default (AzureServices).
- `AZURE_SERVICES` -- Trusted Azure services (ACR Tasks, Microsoft Defender) may reach the registry despite network restrictions.
- `NONE` -- No bypass: network rules apply to everything, first-party services included.

### spec.networkRuleSet

`AzureContainerRegistryNetworkRuleSet`

Network access rules for a PUBLIC registry: a default action plus an
IP allowlist. PREMIUM only. This is the middle ground between "open
to the internet" and "private endpoints only" -- keep
public_network_access_enabled true, set default_action DENY, and
allowlist the CIDR ranges (office egress, CI runners) that may reach
the registry. Updatable in place.

### spec.networkRuleSet.defaultAction

`enum`

What happens to requests matching no ip_rules. Unspecified applies
Azure's default (ALLOW) -- which makes the rule set a no-op; a real
allowlist sets DENY and enumerates ip_rules.

Allowed values (use exactly as shown):

- `azure_container_registry_network_rule_default_action_unspecified` -- Not specified: Azure's default (Allow).
- `ALLOW` -- Requests matching no rules are allowed (the rule set is effectively off).
- `DENY` -- Requests matching no rules are denied -- the allowlist posture.

### spec.networkRuleSet.ipRules

`[]AzureContainerRegistryIpRule`

The public CIDR ranges allowed to reach the registry (ARM only
supports allow rules, so the entries carry no per-rule action).
IPv4 only, e.g. "203.0.113.0/24".

### spec.networkRuleSet.ipRules[].ipRange

`string` · required

The public IPv4 range being allowed, in CIDR form (a single address is
/32).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.georeplications

`[]AzureContainerRegistryGeoreplication`

Additional Azure regions the registry is replicated to. PREMIUM only,
and the list must not contain the registry's own region (the home
replica is implicit). Each replica serves pulls locally (lower
latency, no cross-region egress fees) and the registry stays
available for pulls if a replicated region goes down; pushes
propagate to all replicas automatically. Replicas can be added and
removed in place.

### spec.georeplications[].location

`string` · required

The Azure region to replicate into, e.g. "westus2". Must differ from
the registry's own region and from every other replication.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.georeplications[].zoneRedundancyEnabled

`bool`

Whether this replica's data is spread across the region's availability
zones. Changing it replaces the underlying replication (a re-sync, not
a registry replacement).

### spec.georeplications[].globalEndpointRoutingEnabled

`bool`

Whether this replica participates in global endpoint routing for the
registry's geo-replicated login server: on, Azure may route pulls
addressed to the global endpoint ({name}.azurecr.io) to this replica;
off, the replica still syncs data but is excluded from global
routing. The provider requires an explicit choice; unset deploys
false.

### spec.georeplications[].tags

`map<string, string>`

Tags applied to this replication (each replication is its own tracked
ARM resource); independent of the registry's tags.

### spec.identity

`AzureContainerRegistryIdentity`

The registry's managed identity. Required for customer-managed-key
encryption (which needs a user-assigned identity to unwrap the key at
boot); also usable by first-party integrations that authenticate as
the registry. Leave unset for registries that need no identity.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure with
the registry; USER_ASSIGNED brings identities you manage (required for
CMK encryption, whose key must be unwrappable before the registry's
own identity exists); SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_container_registry_identity_type_unspecified` -- Not specified: the registry has no managed identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the registry.
- `USER_ASSIGNED` -- Bring-your-own user-assigned identities (set identity_ids).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned identity and the listed user-assigned ones.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the registry, by ARM ID. Reference
AzureUserAssignedIdentity resources so grants (Key Vault crypto
access for CMK) can be composed before the registry is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.encryption

`AzureContainerRegistryEncryption`

Customer-managed-key (CMK) encryption: the registry's data is
encrypted with a Key Vault key you own and can rotate or revoke,
instead of Microsoft-managed keys. PREMIUM only, and fixed at
creation. Requires a USER_ASSIGNED identity (declared in `identity`)
that holds get/wrapKey/unwrapKey permission on the key's vault.

### spec.encryption.identityClientId

`string | valueFrom` · required

The client ID of the user-assigned identity that unwraps the key --
the same identity declared in identity.identity_ids. Can be a literal
UUID or a reference to an AzureUserAssignedIdentity's client_id
output.

- references: AzureUserAssignedIdentity (`status.outputs.client_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.encryption.keyVaultKeyId

`string | valueFrom` · required

The Key Vault key encrypting the registry, as the key's full Key
Vault ID, e.g. "https://{vault}.vault.azure.net/keys/{name}" (pin a
version suffix to freeze rotation). Defaults to referencing an
AzureKeyVaultKey's versionless_id output in composed environments,
so key rotation propagates to the registry automatically.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Free-form tags applied to the registry, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins. Tags are Azure's governance
surface -- Azure Policy enforces them and Microsoft Cost Management
groups by them. Updatable in place.

## Validation Rules

- `acr_georeplications_premium_only`: georeplications requires the PREMIUM SKU
- `acr_georeplications_exclude_home_region`: georeplications must not contain the registry's own region -- the home replica is implicit
- `acr_zone_redundancy_premium_only`: zone_redundancy_enabled requires the PREMIUM SKU
- `acr_anonymous_pull_standard_or_premium`: anonymous_pull_enabled requires the STANDARD or PREMIUM SKU
- `acr_data_endpoint_premium_only`: data_endpoint_enabled requires the PREMIUM SKU
- `acr_quarantine_premium_only`: quarantine_policy_enabled requires the PREMIUM SKU
- `acr_retention_premium_only`: retention_policy_in_days requires the PREMIUM SKU
- `acr_export_disable_premium_and_private`: disabling export_policy_enabled requires the PREMIUM SKU and public_network_access_enabled explicitly false
- `acr_network_rule_set_premium_only`: network_rule_set requires the PREMIUM SKU
- `acr_encryption_premium_only`: encryption (customer-managed keys) requires the PREMIUM SKU
- `acr_encryption_requires_user_assigned_identity`: encryption requires a USER_ASSIGNED (or SYSTEM_AND_USER_ASSIGNED) identity -- the user-assigned identity unwraps the key

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureContainerRegistry, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.container_registry_id` | `string` | The Azure Resource Manager ID of the registry. This is the primary output: AzureAksCluster's container_registry_id references it to wire image pulls, and role assignments (AcrPull/AcrPush) scope to it. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ContainerRegistry/registries/{name} |
| `status.outputs.container_registry_name` | `string` | The name of the registry. |
| `status.outputs.login_server` | `string` | The registry's login server -- the hostname images are tagged with and pulled from, e.g. "myregistry.azurecr.io". |
| `status.outputs.admin_username` | `string` | The admin account's username (the registry name), populated only when admin_user_enabled is true. |
| `status.outputs.admin_password` | `string` | One of the admin account's two rotatable passwords, populated only when admin_user_enabled is true. Static credential material -- prefer Entra-based authentication wherever the consumer supports it. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the registry's system-assigned identity, populated only when the identity type includes SYSTEM_ASSIGNED. Grant this principal roles to let the registry act on other resources. |
| `status.outputs.data_endpoint_host_names` | `[]string` | The dedicated regional data-endpoint hostnames (home region plus each geo-replication), populated only when data_endpoint_enabled is true -- the exact hostnames an egress firewall must allowlist. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.encryption.identityClientId` | AzureUserAssignedIdentity | `status.outputs.client_id` |
| `spec.encryption.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAiFoundry | `spec.containerRegistryId` | `status.outputs.container_registry_id` |
| AzureAksCluster | `spec.bootstrapProfile.containerRegistryId` | `status.outputs.container_registry_id` |
| AzureContainerApp | `spec.registries[].server` | `status.outputs.login_server` |
| AzureContainerAppJob | `spec.registries[].server` | `status.outputs.login_server` |
| AzureContainerInstance | `spec.imageRegistryCredentials[].server` | `status.outputs.login_server` |
| AzureMachineLearningWorkspace | `spec.containerRegistryId` | `status.outputs.container_registry_id` |

## See Also

- [Overview](../README.md)
